package blockchain

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"streamr_api/common"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type TxManager struct {
	client       *ethclient.Client
	privateKey   *ecdsa.PrivateKey
	contractAddr ethcommon.Address
	contractAbi  abi.ABI

	nonce               chan uint64
	sendContractTxQueue chan types.Transaction
}

func NewTxManager(privateKey *ecdsa.PrivateKey, contractAddr ethcommon.Address, contractAbi abi.ABI) (*TxManager, error) {
	client, err := ethclient.Dial(common.GetStringEnvWithDefault("RPC_ADDR", "https://polygon-rpc.com"))
	if err != nil {
		panic(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, err
	}

	nChan := make(chan uint64, 1)
	nChan <- nonce
	tm := TxManager{
		client:       client,
		privateKey:   privateKey,
		contractAddr: contractAddr,
		contractAbi:  contractAbi,

		nonce:               nChan,
		sendContractTxQueue: make(chan types.Transaction, 1000),
	}
	return &tm, nil
}

func (tm *TxManager) ContractCall(method string, params []interface{}) ([]interface{}, error) {
	// Creating a call message
	data, err := tm.contractAbi.Pack(method, params...)
	if err != nil {
		return nil, err
	}

	callMsg := ethereum.CallMsg{
		To:   &tm.contractAddr,
		Data: data,
	}

	output, err := tm.client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		log.Fatalf("Failed to execute contract call: %v", err)
		return nil, err
	}

	result, err := tm.contractAbi.Unpack(method, output)
	if err != nil {
		log.Fatalf("Failed to unpack the output: %v", err)
		return nil, err
	}

	fmt.Println("Result:", result)
	return result, nil
}

// created a separate function for calling functions with non-standard return types. This returns a byte string that the calling function can handle as needed.
func (tm *TxManager) ContractCallSpecial(method string, params []interface{}) ([]byte, error) {
	// Creating a call message
	data, err := tm.contractAbi.Pack(method, params...)
	if err != nil {
		return nil, err
	}

	callMsg := ethereum.CallMsg{
		To:   &tm.contractAddr,
		Data: data,
	}

	output, err := tm.client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		log.Fatalf("Failed to execute contract call: %v", err)
		return nil, err
	}

	result, err := tm.contractAbi.Unpack(method, output)
	if err != nil {
		log.Fatalf("Failed to unpack the output: %v", err)
		return nil, err
	}

	fmt.Println("Result:", result)
	byteResult, _ := json.Marshal(result)
	return byteResult, nil
}

func (tm *TxManager) ContractSendTxWithWait(method string, params []interface{}, duration time.Duration) (*types.Transaction, error) {
	txHash, err := tm.ContractSendTx(method, params)
	if err != nil {
		return nil, err
	}

	tx, err := tm.PolygonWaitForTx(txHash, duration)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (tm *TxManager) ContractSendTx(method string, params []interface{}) (string, error) {
	gasPrice, err := tm.client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}

	chainID, err := tm.client.NetworkID(context.Background())
	if err != nil {
		return "", err
	}

	// Prepare the transaction options
	auth, err := bind.NewKeyedTransactorWithChainID(tm.privateKey, chainID) // Chain ID for Polygon Mainnet
	if err != nil {
		return "", err
	}

	nonce := <-tm.nonce
	defer func() { tm.nonce <- nonce }()

	auth.Nonce = big.NewInt(int64(nonce))
	log.Printf("Nonce: %d\n", nonce)
	auth.Value = big.NewInt(0)      // in wei (0 if your function is not payable)
	auth.GasLimit = uint64(3000000) // set the gas limit to a suitable value
	auth.GasPrice = gasPrice
	// Pack the data to send in the transaction
	inputData, err := tm.contractAbi.Pack(method, params...)
	if err != nil {
		return "", err
	}

	// Create the transaction
	tx := types.NewTransaction(auth.Nonce.Uint64(), tm.contractAddr, auth.Value, auth.GasLimit, auth.GasPrice, inputData)

	// Sign the transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), tm.privateKey)
	if err != nil {
		return "", err
	}

	// Send the transaction
	err = tm.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}

	nonce += 1
	fmt.Printf("Transaction sent! TX Hash: %s\n", signedTx.Hash().Hex())

	return signedTx.Hash().Hex(), nil
}

func (tm *TxManager) PolygonWaitForTx(txHash string, duration time.Duration) (*types.Transaction, error) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	timeout := time.After(duration)

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("timeout reached while waiting for transaction to be mined")
		case <-ticker.C:
			tx, isPending, err := tm.client.TransactionByHash(context.Background(), ethcommon.HexToHash(txHash))
			// Note: It's normal for a transaction not to be found immediately after submission.
			if err != nil {
				fmt.Printf("Waiting for transaction to be recognized by the network: %v\n", err)
				continue // Skip this iteration and try again.
			}

			if !isPending {
				fmt.Println("Transaction has been mined.")
				return tx, nil // Transaction is not pending anymore, break the loop.
			}
		}
	}
}
