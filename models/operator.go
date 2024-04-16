package models

import (
	"crypto/ecdsa"
	"encoding/json"
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

type UndelegationRecordResponse struct {
	Delegator ethcommon.Address `json:"delegator"`
	Amount    *big.Int          `json:"amountWei"`
	Timestamp *big.Int          `json:"timestamp"`
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

func (o *Operator) GetValueWithoutEarnings() (*big.Int, error) {
	result, err := o.TxManager.ContractCall("valueWithoutEarnings", []interface{}{})
	if err != nil {
		log.Fatalf("Failed to unpack the output: %v", err)
		return nil, err
	}

	return result[0].(*big.Int), nil
}

func (o *Operator) GetUndelegationQueue() ([][]UndelegationRecordResponse, error) {
	result, err := o.TxManager.ContractCallSpecial("undelegationQueue", []interface{}{})
	if err != nil {
		log.Fatalf("Failed to unpack the output: %v", err)
		return nil, err
	}

	var jsonResult [][]UndelegationRecordResponse
	err = json.Unmarshal(result, &jsonResult)
	if err != nil {
		log.Printf("Failed to unmarshal record: %v", err)
		return nil, err
	}

	return jsonResult, nil
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
	for sponsorA := range deployedStake.DeployedBySponsorship {
		for i, sponsorB := range sponsors.Addresses {
			if sponsorA == sponsorB {

				// the protocol take 5% of earnings, so reduce the earnings by 5% before compounding
				amount := sponsors.Earnings[i]
				amount = amount.Mul(amount, big.NewInt(95))
				amount = amount.Div(amount, big.NewInt(100))

				log.Printf("Adding %s stake to %s\n", sponsors.Earnings[i].String(), sponsorA.Hex())

				tx, err := o.Stake(sponsorA, amount)
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

func (o *Operator) StakeProRata() ([]string, error) {
	deployedStake, err := o.GetDeployedStake()
	if err != nil {
		log.Printf("Failed to get deployed stake: %v", err)
		return nil, err
	}
	log.Printf("Deployed Stake: %v\n", deployedStake.TotalDeployed)
	totalValue, err := o.GetValueWithoutEarnings()
	unstaked := new(big.Int).Sub(totalValue, deployedStake.TotalDeployed)
	log.Printf("Unstaked: %s\nerr: %v", unstaked.String(), err)
	// use a DeployedStakeResponse object to calculate and store amounts of stake to deploy to each sponsorship
	// using this data type for convenience
	var stakeProRata DeployedStakeResponse = DeployedStakeResponse{
		DeployedBySponsorship: make(map[ethcommon.Address]*big.Int),
		TotalDeployed:         totalValue,
	}

	// calculate the total amount of stake to deploy to each sponsoship based on total available DATA and
	// amount of DATA already deployed to each sponsorship
	for sponsorhip, deployed := range deployedStake.DeployedBySponsorship {
		// calculate share of new stake to this sponsorship. Multiply first to avoid rounding errors, then divide by totalDeployed
		numerator := new(big.Int)
		numerator.Mul(unstaked, deployed)
		share := new(big.Int).Div(numerator, deployedStake.TotalDeployed)
		stakeProRata.DeployedBySponsorship[sponsorhip] = share
		stakeProRata.TotalDeployed.Add(stakeProRata.TotalDeployed, share)
		log.Printf("Share for %s: %s\n", sponsorhip.Hex(), share.String())
	}

	// check to see if the total amount of stake to deploy is equal to the total amount of DATA available
	// if not, adjust the stake to deploy to the first sponsorship to make up the difference
	if stakeProRata.TotalDeployed.Cmp(totalValue) != 0 {
		adjustment := new(big.Int).Sub(totalValue, stakeProRata.TotalDeployed)
		lastSponsorship := ethcommon.Address{}
		for sponsorhip := range deployedStake.DeployedBySponsorship {
			lastSponsorship = sponsorhip
			break
		}
		stakeProRata.DeployedBySponsorship[lastSponsorship].Add(stakeProRata.DeployedBySponsorship[lastSponsorship], adjustment)
		stakeProRata.TotalDeployed.Add(stakeProRata.TotalDeployed, adjustment)
	}

	// iterate through the sponsorships and deploy the calculated amount of stake to each
	txList := []string{}
	for sponsorhip, amount := range stakeProRata.DeployedBySponsorship {
		tx, err := o.Stake(sponsorhip, amount)
		if err != nil {
			log.Printf("Failed to increase stake: %v", err)
			return nil, err
		}
		go func() {
			txResult, err := o.TxManager.PolygonWaitForTx(tx, 120*time.Second)
			if err != nil {
				log.Printf("Failed to wait for transaction: %v", err)
			} else {
				log.Printf("Transaction %s completed as: %v\n", txResult.Hash().Hex(), txResult)
			}
		}()
		txList = append(txList, tx)
	}

	return txList, nil
}
