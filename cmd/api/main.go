package main

import (
	"go.uber.org/zap"

	"github.com/code-xd/k8s-deployment-manager/internal/api"
	"github.com/code-xd/k8s-deployment-manager/internal/repository/nats"
	natscommon "github.com/code-xd/k8s-deployment-manager/internal/repository/nats/common"
	"github.com/code-xd/k8s-deployment-manager/internal/repository/postgres"
	"github.com/code-xd/k8s-deployment-manager/internal/repository/postgres/common"
	"github.com/code-xd/k8s-deployment-manager/internal/service"
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

// main is the composition root - this is the ONLY place where concrete implementations
// should be instantiated. All other layers interact via interfaces from pkg/ports/
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

	// Initialize database connection (concrete implementation - OK in composition root)
	db, err := common.NewDB(&apiCfg.Database, dto.Log)
	if err != nil {
		dto.Log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize NATS connection and producer (concrete implementation - OK in composition root)
	natsConn, err := natscommon.NewNATS(&apiCfg.Nats, dto.Log)
	if err != nil {
		dto.Log.Fatal("Failed to connect to NATS", zap.Error(err))
	}
	defer natsConn.Close()

	// Ensure JetStream stream exists (stream name and subjects from config)
	prod := &apiCfg.Nats.Producer
	subjects := []string{prod.DeploymentRequestChannel, prod.DeploymentUpdateChannel}
	if err := natsConn.EnsureStream(prod.StreamName, subjects); err != nil {
		dto.Log.Fatal("Failed to ensure JetStream stream", zap.Error(err))
	}

	natsProducer := natscommon.NewProducer(natsConn)
	deploymentRequestPublisher := nats.NewDeploymentRequestProducer(natsProducer, prod)

	// Initialize repositories (concrete implementations - OK in composition root)
	// These implement interfaces from pkg/ports/ and are injected as interfaces
	deploymentRequestRepo := postgres.NewDeploymentRequestRepository(db)
	deploymentRepo := postgres.NewDeploymentRepository(db)
	userRepo := postgres.NewUserRepository(db)

	// Initialize services (concrete implementations - OK in composition root)
	// Service receives repository as interface (ports/repo.DeploymentRequest)
	// Service returns interface (ports/service.DeploymentRequest)
	deploymentRequestService := service.NewDeploymentRequestService(
		deploymentRequestRepo,
		deploymentRepo,
		deploymentRequestPublisher,
		dto.Log,
	)

	// Setup router with injected service dependencies (as interface from pkg/ports/)
	router := api.SetupRouter(
		dto.Log,
		deploymentRequestService,
		userRepo,
		deploymentRequestRepo,
	)
	server := api.NewServer(dto.APICfg, router)

	if err := server.Run(); err != nil {
		dto.Log.Fatal("Server failed", zap.Error(err))
	}

	defer server.Shutdown()

	utils.WaitForShutdown()
}
