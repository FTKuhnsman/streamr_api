package models

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"streamr_api/common"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Operator struct {
	ContractAddr ethcommon.Address `json:"contractAddr"`
	ContractAbi  abi.ABI           `json:"contractAbi"`
	OwnerAddr    ethcommon.Address `json:"ownerAddr"`
	privateKey   *ecdsa.PrivateKey
	client       *ethclient.Client
}

type GetSponsorshipsAndEarningsResponse struct {
	Addresses          []ethcommon.Address `json:"addresses"`
	Earnings           []*big.Int          `json:"earnings"`
	MaxAllowedEarnings *big.Int            `json:"maxAllowedEarnings"`
}

type DeployedStakeResponse struct {
	DeployedBySponsorship map[ethcommon.Address]*big.Int `json:"deployedBySponsorship"`
	TotalDeployed         []*big.Int                     `json:"totalDeployed"`
}

type StakedIntoResponse struct {
	StakedInto *big.Int `json:"stakedInto"`
}

func NewOperator(contractAddr, ownerAddr, privateKey string) *Operator {
	client, err := ethclient.Dial(common.GetStringEnvWithDefault("RPC_ADDR", "https://polygon-rpc.com"))
	if err != nil {
		panic(err)
	}

	abiStr, err := common.FetchContractABI(contractAddr)
	if err != nil {
		panic(err)
	}

	contractABI, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		log.Fatalf("Invalid ABI: %v", err)
	}

	privKey, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	return &Operator{
		ContractAddr: ethcommon.HexToAddress(contractAddr),
		ContractAbi:  contractABI,
		OwnerAddr:    ethcommon.HexToAddress(ownerAddr),
		privateKey:   privKey,
		client:       client,
	}
}

func (o *Operator) GetValueWithoutEarnings() interface{} {
	result, err := PolygonCall(o.client, o.ContractAddr, o.ContractAbi, "valueWithoutEarnings", []interface{}{})
	if err != nil {
		log.Fatalf("Failed to unpack the output: %v", err)
		return err
	}

	fmt.Println("Result:", result)
	return result
}

func (o *Operator) WithdrawEarnings() interface{} {
	params := []interface{}{} // The parameters for your method, if any
	addr1 := ethcommon.HexToAddress("0xc95e7aa2436a2ab8eebae10079d4cbf556adc55c")
	addList := []ethcommon.Address{addr1}

	params = append(params, addList)

	result, err := PolygonSendTx(o.client, o.privateKey, o.ContractAddr, o.ContractAbi, "withdrawEarningsFromSponsorships", params)
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
		return err
	}

	return result
}

func (o *Operator) GetSponsorshipsAndEarnings() (GetSponsorshipsAndEarningsResponse, error) {
	result, err := PolygonCall(o.client, o.ContractAddr, o.ContractAbi, "getSponsorshipsAndEarnings", []interface{}{})
	if err != nil {
		log.Fatalf("Failed to unpack the output: %v", err)
		return GetSponsorshipsAndEarningsResponse{}, err
	}
	var jsonResult GetSponsorshipsAndEarningsResponse = GetSponsorshipsAndEarningsResponse{
		Addresses:          result[0].([]ethcommon.Address),
		Earnings:           result[1].([]*big.Int),
		MaxAllowedEarnings: result[2].(*big.Int),
	}

	fmt.Println("Result:", jsonResult)
	return jsonResult, nil
}

func (o *Operator) StakedInto(sponsorshipAddr ethcommon.Address) (StakedIntoResponse, error) {
	var params []interface{}
	params = append(params, sponsorshipAddr)
	result, err := PolygonCall(o.client, o.ContractAddr, o.ContractAbi, "stakedInto", params)
	if err != nil {
		log.Fatalf("Failed to unpack the output: %v", err)
		return StakedIntoResponse{}, err
	}

	bigIntPointer, ok := result[0].(*big.Int)
	if !ok {
		log.Fatalf("Failed to cast the result to *big.Int: %v", err)
		return StakedIntoResponse{}, err
	}

	var jsonResult StakedIntoResponse = StakedIntoResponse{
		StakedInto: bigIntPointer,
	}

	fmt.Println("Result:", jsonResult)
	return jsonResult, nil
}

func (o *Operator) GetDeployedStake() (DeployedStakeResponse, error) {
	SAE, err := o.GetSponsorshipsAndEarnings()
	if err != nil {
		return DeployedStakeResponse{}, err
	}

	for _, addr := range SAE.Addresses {
		//
		log.Printf("Address: %s\n", addr.Hex())
		// finish this
	}

	return DeployedStakeResponse{}, nil
	// update return statement with actual value
}

func PolygonCall(client *ethclient.Client, contractAddress ethcommon.Address, contractABI abi.ABI, method string, params []interface{}) ([]interface{}, error) {
	// Creating a call message
	data, err := contractABI.Pack(method, params...)
	if err != nil {
		return nil, err
	}

	callMsg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}

	output, err := client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		log.Fatalf("Failed to execute contract call: %v", err)
		return nil, err
	}

	result, err := contractABI.Unpack(method, output)
	if err != nil {
		log.Fatalf("Failed to unpack the output: %v", err)
		return nil, err
	}

	fmt.Println("Result:", result)
	return result, nil
}

func PolygonSendTx(client *ethclient.Client, privateKey *ecdsa.PrivateKey, contractAddress ethcommon.Address, contractABI abi.ABI, method string, params []interface{}) (interface{}, error) {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("Cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, err
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}

	// Prepare the transaction options
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID) // Chain ID for Polygon Mainnet
	if err != nil {
		return nil, err
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)      // in wei (0 if your function is not payable)
	auth.GasLimit = uint64(3000000) // set the gas limit to a suitable value
	auth.GasPrice = gasPrice
	// Pack the data to send in the transaction
	inputData, err := contractABI.Pack(method, params...)
	if err != nil {
		return nil, err
	}

	// Create the transaction
	tx := types.NewTransaction(auth.Nonce.Uint64(), contractAddress, auth.Value, auth.GasLimit, auth.GasPrice, inputData)

	// Sign the transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return nil, err
	}

	// Send the transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Transaction sent! TX Hash: %s\n", signedTx.Hash().Hex())

	return signedTx.Hash().Hex(), nil
}
