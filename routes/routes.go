package routes

import (
	"streamr_api/handlers"
	"streamr_api/models"

	"github.com/gin-gonic/gin"
)

func SetupRouter(o *models.Operator) *gin.Engine {
	router := gin.Default()

	v1 := router.Group("/api/v1")
	{
		v1.GET("/operator", handlers.GetOperator(o))
		v1.GET("/operator/valuewithoutearnings", handlers.OperatorValueWithoutEarnings(o))
		v1.GET("/operator/withdrawearnings", handlers.OperatorWithdrawEarnings(o))
		v1.GET("/operator/sponsorshipsandearnings", handlers.SponsorshipsAndEarnings(o))
		v1.GET("/operator/stakedinto/:address", handlers.StakedInto(o))
	}

	return router
}
