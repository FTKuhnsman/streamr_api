package handlers

import (
	"net/http"

	"streamr_api/models"

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
