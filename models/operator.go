package models

import (
	"crypto/ecdsa"
	"log"
	"math/big"
	"streamr_api/blockchain"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Operator struct {
	ContractAddr ethcommon.Address `json:"contractAddr"`
	ContractAbi  abi.ABI           `json:"contractAbi"`
	OwnerAddr    ethcommon.Address `json:"ownerAddr"`
	PrivateKey   *ecdsa.PrivateKey
	TxManager    *blockchain.TxManager
}

type GetSponsorshipsAndEarningsResponse struct {
	Addresses          []ethcommon.Address `json:"addresses"`
	Earnings           []*big.Int          `json:"earnings"`
	MaxAllowedEarnings *big.Int            `json:"maxAllowedEarnings"`
}

type DeployedStakeResponse struct {
	DeployedBySponsorship map[ethcommon.Address]*big.Int `json:"deployedBySponsorship"`
	TotalDeployed         *big.Int                       `json:"totalDeployed"`
}

type StakedIntoResponse struct {
	StakedInto *big.Int `json:"stakedInto"`
}

func NewOperator(contractAddr string, ownerAddr string, privateKey string) *Operator {
	abiStr, err := blockchain.FetchContractABI(contractAddr)
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

	txManager, err := blockchain.NewTxManager(privKey, ethcommon.HexToAddress(contractAddr), contractABI)
	if err != nil {
		log.Fatalf("Failed to create tx manager: %v", err)
	}

	return &Operator{
		ContractAddr: ethcommon.HexToAddress(contractAddr),
		ContractAbi:  contractABI,
		OwnerAddr:    ethcommon.HexToAddress(ownerAddr),
		PrivateKey:   privKey,
		TxManager:    txManager,
	}
}

func (o *Operator) GetValueWithoutEarnings() interface{} {
	result, err := o.TxManager.ContractCall("valueWithoutEarnings", []interface{}{})
	if err != nil {
		log.Fatalf("Failed to unpack the output: %v", err)
		return err
	}
	return result
}

func (o *Operator) WithdrawEarnings() (string, error) {

	sponsors, err := o.GetSponsorshipsAndEarnings()
	if err != nil {
		log.Fatalf("Failed to get sponsorships and earnings: %v", err)
		return "", err
	}

	params := []interface{}{} // The parameters for your method, if any

	params = append(params, sponsors.Addresses)

	result, err := o.TxManager.ContractSendTx("withdrawEarningsFromSponsorships", params)
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
		return "", err
	}

	return result, nil
}

func (o *Operator) WithdrawEarningsAndCompound() ([]string, error) {
	txList := []string{}
	sponsors, err := o.GetSponsorshipsAndEarnings()
	if err != nil {
		log.Fatalf("Failed to get sponsorships and earnings: %v", err)
		return nil, err
	}

	params := []interface{}{} // The parameters for your method, if any

	params = append(params, sponsors.Addresses)

	result, err := o.TxManager.ContractSendTx("withdrawEarningsFromSponsorships", params)
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
		return nil, err
	}

	txList = append(txList, result)

	// get current deploy stake
	deployedStake, err := o.GetDeployedStake()
	if err != nil {
		log.Fatalf("Failed to get deployed stake: %v", err)
		return nil, err
	}

	// iterate each sponsor and set stake to the current amount plus the earnings withdrawn in the previous transaction
	for sponsorA, currentAmount := range deployedStake.DeployedBySponsorship {
		for i, sponsorB := range sponsors.Addresses {
			if sponsorA == sponsorB {
				newAmount := new(big.Int).Add(currentAmount, sponsors.Earnings[i])
				log.Printf("Adding %s stake to %s\n", newAmount.String(), sponsorA.Hex())
				tx, err := o.Stake(sponsorA, sponsors.Earnings[i])
				go func() {
					txResult, err := o.TxManager.PolygonWaitForTx(tx, 120*time.Second)
					if err != nil {
						log.Printf("Failed to wait for transaction: %v", err)
					} else {
						log.Printf("Transaction %s completed as: %v\n", txResult.Hash().Hex(), txResult)
					}
				}()

				if err != nil {
					log.Fatalf("Failed to reduce stake: %v", err)
					return nil, err
				}
				txList = append(txList, tx)
			}
		}
	}

	return txList, nil
}

func (o *Operator) GetSponsorshipsAndEarnings() (GetSponsorshipsAndEarningsResponse, error) {
	result, err := o.TxManager.ContractCall("getSponsorshipsAndEarnings", []interface{}{})
	if err != nil {
		log.Fatalf("Failed to unpack the output: %v", err)
		return GetSponsorshipsAndEarningsResponse{}, err
	}
	var jsonResult GetSponsorshipsAndEarningsResponse = GetSponsorshipsAndEarningsResponse{
		Addresses:          result[0].([]ethcommon.Address),
		Earnings:           result[1].([]*big.Int),
		MaxAllowedEarnings: result[2].(*big.Int),
	}
	return jsonResult, nil
}

func (o *Operator) StakedInto(sponsorshipAddr ethcommon.Address) (StakedIntoResponse, error) {
	var params []interface{}
	params = append(params, sponsorshipAddr)

	result, err := o.TxManager.ContractCall("stakedInto", params)

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

	return jsonResult, nil
}

func (o *Operator) GetDeployedStake() (DeployedStakeResponse, error) {
	SAE, err := o.GetSponsorshipsAndEarnings()
	if err != nil {
		return DeployedStakeResponse{}, err
	}

	response := DeployedStakeResponse{
		DeployedBySponsorship: make(map[ethcommon.Address]*big.Int),
		TotalDeployed:         big.NewInt(0),
	}

	for _, addr := range SAE.Addresses {
		log.Printf("Address: %s\n", addr.Hex())
		sponsorDeployed, err := o.StakedInto(addr)
		if err != nil {
			return DeployedStakeResponse{}, err
		}
		response.DeployedBySponsorship[addr] = sponsorDeployed.StakedInto
		response.TotalDeployed.Add(response.TotalDeployed, sponsorDeployed.StakedInto)
	}
	log.Printf("Total Deployed: %s\n", response.TotalDeployed.String())
	return response, nil
	// update return statement with actual value
}

func (o *Operator) ReduceStakeTo(addr ethcommon.Address, targetStake *big.Int) (string, error) {
	params := []interface{}{addr, targetStake}
	result, err := o.TxManager.ContractSendTx("reduceStakeTo", params)
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
		return "", err
	}

	return result, nil
}

func (o *Operator) Stake(addr ethcommon.Address, targetStake *big.Int) (string, error) {
	params := []interface{}{addr, targetStake}
	result, err := o.TxManager.ContractSendTx("stake", params)
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
		return "", err
	}

	return result, nil
}
