package handlers

import (
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
// @Success      200  {array}  models.Operator
// @Router       /operator/valuewithoutearnings [get]
func OperatorValueWithoutEarnings(o *models.Operator) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		result := o.GetValueWithoutEarnings()
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
// @Summary      Get the sponsorships and earnings.
// @Description  Responds with the list of sponsorships and uncollected earnings.
// @Tags         Operator
// @Produce      json
// @Success      200  {array}  models.Operator
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
// @Success      200  {array}  models.Operator
// @Router       /operator/withdrawearnings [get]
func OperatorWithdrawEarnings(o *models.Operator) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		result := o.WithdrawEarnings()
		c.JSON(http.StatusOK, result)
	}

	return gin.HandlerFunc(fn)
}
