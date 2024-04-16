package main

import (
	"fmt"
	"streamr_api/common"
	"streamr_api/models"
	"streamr_api/routes"

	_ "streamr_api/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Streamr Operator Service
// @version         1.0
// @description     A Streamr Operator management service API in Go using Gin framework.

// @contact.name   Joseph Kuhnsman
// @contact.email  admin@ftkuhnsman.com

// @BasePath  /api/v1

func init() {

}

func main() {

	operator := models.NewOperator(
		common.GetStringEnvWithDefault("CONTRACT_ADDR", "0x1234567890"),
		common.GetStringEnvWithDefault("OWNER_ADDR", "0x1234567890"),
		common.GetStringEnvWithDefault("PRIVATE_KEY", "0x1234567890"),
	)

	scheduler := models.NewScheduler()

	router := routes.SetupRouter(operator, scheduler)

	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run(fmt.Sprintf(":%d", common.GetIntEnvWithDefault("PORT", 8080)))
}
