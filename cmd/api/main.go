package main

import (
	"go.uber.org/zap"

	"github.com/code-xd/k8s-deployment-manager/internal/api"
	"github.com/code-xd/k8s-deployment-manager/pkg/config"
	"github.com/code-xd/k8s-deployment-manager/pkg/constants"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	"github.com/code-xd/k8s-deployment-manager/pkg/logger"
	"github.com/code-xd/k8s-deployment-manager/pkg/utils"
	_ "github.com/code-xd/k8s-deployment-manager/swagger"
)

// @title           K8s Deployment Manager API
// @version         1.0
// @description     A simple API for Kubernetes deployment management
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath  /api/v1

func main() {
	// Initialize global logger
	dto.Log = logger.New()
	defer dto.Log.Sync()

	// Initialize global config
	cfg := config.NewConfigLoader[dto.APIConfig](
		constants.DEFAULT_CONFIG_PATH,
		constants.DEFAULT_CONFIG_FILE,
	)

	apiCfg, err := cfg.Load()
	if err != nil {
		dto.Log.Fatal("Failed to load config", zap.Error(err))
	}
	dto.APICfg = apiCfg

	router := api.SetupRouter(dto.Log)
	server := api.NewServer(*dto.APICfg, router)

	if err := server.Run(); err != nil {
		dto.Log.Fatal("Server failed", zap.Error(err))
	}

	utils.WaitForShutdown()
	if err := server.Shutdown(); err != nil {
		dto.Log.Fatal("Server failed to shutdown", zap.Error(err))
	}
}
