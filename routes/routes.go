package routes

import (
	"streamr_api/handlers"
	"streamr_api/models"

	"github.com/gin-gonic/gin"
)

func SetupRouter(o *models.Operator) *gin.Engine {
	router := gin.Default()
	root := router.Group("/")
	{
		root.GET("/", func(c *gin.Context) {
			c.Redirect(301, "/docs/index.html")
		})
	}
	v1 := router.Group("/api/v1")
	{
		v1.GET("/operator", handlers.GetOperator(o))
		v1.GET("/operator/valuewithoutearnings", handlers.OperatorValueWithoutEarnings(o))
		v1.GET("/operator/withdrawearnings", handlers.OperatorWithdrawEarnings(o))
		v1.GET("/operator/withdrawearningsandcompound", handlers.WithdrawEarningsAndCompound(o))
		v1.GET("/operator/sponsorshipsandearnings", handlers.SponsorshipsAndEarnings(o))
		v1.GET("/operator/stakedinto/:address", handlers.StakedInto(o))
		v1.GET("/operator/deployedstake", handlers.DeployedStake(o))
		v1.GET("/operator/reducestaketo/:sponsorship/:amount", handlers.ReduceStakeTo(o))
		v1.GET("/operator/stake/:sponsorship/:amount", handlers.Stake(o))
	}

	return router
}
