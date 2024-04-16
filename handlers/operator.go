package handlers

import (
	"math/big"
	"net/http"

	"streamr_api/models"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

// GetOperator             godoc
// @Summary      Get the Streamr Operator details.
// @Description  Responds with the Operator attributes.
// @Tags         Operator
// @Produce      json
// @Success      200  {array}  models.Operator
// @Router       /operator [get]
func GetOperator(o *models.Operator) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		c.JSON(http.StatusOK, o)
	}

	return gin.HandlerFunc(fn)
}

// OperatorValue             godoc
// @Summary      Get the Streamr Operator details.
// @Description  Responds with the Operator attributes.
// @Tags         Operator
// @Produce      json
// @Success      200  {array}  uint64
// @Router       /operator/valuewithoutearnings [get]
func OperatorValueWithoutEarnings(o *models.Operator) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		result, err := o.GetValueWithoutEarnings()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}

	return gin.HandlerFunc(fn)
}

// OperatorValue             godoc
// @Summary      Get the Streamr Operator staked balance in sponsorship.
// @Description  Responds with the Operator stake deployed in sponsorship.
// @Tags         Operator
// @Produce      json
// @Param        address  path      string  true  "get deployed stake by sponsorship"
// @Success      200  {array}  models.StakedIntoResponse
// @Router       /operator/stakedinto/{address} [get]
func StakedInto(o *models.Operator) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		addr := ethcommon.HexToAddress(c.Param("address"))
		result, err := o.StakedInto(addr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}

	return gin.HandlerFunc(fn)
}

// OperatorValue             godoc
// @Summary      Get the Streamr Operator total deployed stake.
// @Description  Responds with the Operator stake deployed in all sponsorships.
// @Tags         Operator
// @Produce      json
// @Success      200  {array}  models.DeployedStakeResponse
// @Router       /operator/deployedstake/ [get]
func DeployedStake(o *models.Operator) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		result, err := o.GetDeployedStake()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}

	return gin.HandlerFunc(fn)
}

// OperatorValue             godoc
// @Summary      Increase Streamr Operator stake on a given sponsor.
// @Description  Responds with the transaction hash.
// @Tags         Operator
// @Produce      json
// @Param        sponsorship  path      string  true  "sponsorship address"
// @Param        amount  path      int64  true  "amount in wei"
// @Success      200  {array}  string
// @Router       /operator/stake/{sponsorship}/{amount} [get]
func Stake(o *models.Operator) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		addr := ethcommon.HexToAddress(c.Param("sponsorship"))
		amount := new(big.Int)
		_, ok := amount.SetString(c.Param("amount"), 10)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount"})
			return
		}

		result, err := o.Stake(addr, amount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}

	return gin.HandlerFunc(fn)
}

// OperatorValue             godoc
// @Summary      Change Streamr Operator stake on a given sponsor.
// @Description  Responds with the transaction hash.
// @Tags         Operator
// @Produce      json
// @Param        sponsorship  path      string  true  "sponsorship address"
// @Param        amount  path      int64  true  "amount in wei"
// @Success      200  {array}  string
// @Router       /operator/reducestaketo/{sponsorship}/{amount} [get]
func ReduceStakeTo(o *models.Operator) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		addr := ethcommon.HexToAddress(c.Param("sponsorship"))
		amount := new(big.Int)
		_, ok := amount.SetString(c.Param("amount"), 10)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount"})
			return
		}

		result, err := o.ReduceStakeTo(addr, amount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}

	return gin.HandlerFunc(fn)
}

// OperatorValue             godoc
// @Summary      Get the sponsorships and earnings.
// @Description  Responds with the list of sponsorships and uncollected earnings.
// @Tags         Operator
// @Produce      json
// @Success      200  {array}  models.GetSponsorshipsAndEarningsResponse
// @Router       /operator/sponsorshipsandearnings [get]
func SponsorshipsAndEarnings(o *models.Operator) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		result, err := o.GetSponsorshipsAndEarnings()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}

	return gin.HandlerFunc(fn)
}

// WithdrawEarnings             godoc
// @Summary      Get the Streamr Operator details.
// @Description  Responds with the Operator attributes.
// @Tags         Operator
// @Produce      json
// @Success      200  {array}  string
// @Router       /operator/withdrawearnings [get]
func OperatorWithdrawEarnings(o *models.Operator) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		result, err := o.WithdrawEarnings()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}

	return gin.HandlerFunc(fn)
}

// CURRENTLY NOT IMPLEMENTED as this cannot currently handle the undelegation queue, if it exists.
// WithdrawEarningsAndCompound            godoc
// @Summary      Withdraw earnings from sponsorship and restake.
// @Description  Withdraws earnings from all sponsorships and restake to compound.
// @Tags         Operator
// @Produce      json
// @Success      200  {array}  []string
// @Router       /operator/withdrawearningsandcompound [get]
func WithdrawEarningsAndCompound(o *models.Operator) gin.HandlerFunc {
	fn := func(c *gin.Context) {

		result, err := o.WithdrawEarningsAndCompound()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}

	return gin.HandlerFunc(fn)
}

// StakeProRata  godoc
// @Summary      Distribute available DATA to all sponsorships.
// @Description  Increase stake on all sponsorships with all available DATA pro-rated by sponsorship current stake.
// @Tags         Operator
// @Produce      json
// @Success      200  {array}  []string
// @Router       /operator/stakeprorata [get]
func StakeProRata(o *models.Operator) gin.HandlerFunc {
	fn := func(c *gin.Context) {

		result, err := o.StakeProRata()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}

	return gin.HandlerFunc(fn)
}

// UndelegationQueue  godoc
// @Summary      Get the undelegation queue.
// @Description  Responds with the undelegation queue.
// @Tags         Operator
// @Produce      json
// @Success      200  {array}  []uint8
// @Router       /operator/undelegationqueue [get]
func UndelegationQueue(o *models.Operator) gin.HandlerFunc {
	fn := func(c *gin.Context) {

		result, err := o.GetUndelegationQueue()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}

	return gin.HandlerFunc(fn)
}
