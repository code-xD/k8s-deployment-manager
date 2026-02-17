package main

import (
	"go.uber.org/zap"

	"github.com/code-xd/k8s-deployment-manager/internal/repository/k8sclient"
	natscommon "github.com/code-xd/k8s-deployment-manager/internal/repository/nats/common"
	"github.com/code-xd/k8s-deployment-manager/internal/repository/postgres"
	pgcommon "github.com/code-xd/k8s-deployment-manager/internal/repository/postgres/common"
	"github.com/code-xd/k8s-deployment-manager/internal/service/workerService"
	"github.com/code-xd/k8s-deployment-manager/internal/worker"
	"github.com/code-xd/k8s-deployment-manager/pkg/config"
	"github.com/code-xd/k8s-deployment-manager/pkg/constants"
	"github.com/code-xd/k8s-deployment-manager/pkg/consumer"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	"github.com/code-xd/k8s-deployment-manager/pkg/logger"
	"github.com/code-xd/k8s-deployment-manager/pkg/utils"
)

func main() {
	// Initialize logger
	log := logger.New()
	defer log.Sync()

	// Load config
	cfg := config.NewConfigLoader[dto.WorkerConfig](
		constants.DEFAULT_CONFIG_PATH,
		constants.DEFAULT_CONFIG_FILE,
	)
	workerCfg, err := cfg.Load()
	if err != nil {
		log.Fatal("Failed to load config", zap.Error(err))
	}

	// Initialize database connection
	db, err := pgcommon.NewDB(&workerCfg.Database, log)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize NATS connection
	natsConn, err := natscommon.NewNATS(&workerCfg.Nats, log)
	if err != nil {
		log.Fatal("Failed to connect to NATS", zap.Error(err))
	}
	defer natsConn.Close()

	// Ensure JetStream stream exists
	prod := &workerCfg.Nats.Producer
	subjects := []string{prod.DeploymentRequestChannel, prod.DeploymentUpdateChannel}
	if err := natsConn.EnsureStream(prod.StreamName, subjects); err != nil {
		log.Fatal("Failed to ensure JetStream stream", zap.Error(err))
	}

	// Initialize repositories
	deploymentRequestRepo := postgres.NewDeploymentRequestRepository(db)
	k8sDeployment, err := k8sclient.NewDeployment(".", &workerCfg.K8s)
	if err != nil {
		log.Fatal("Failed to create k8s deployment client", zap.Error(err))
	}

	// Create consumer and wire services
	nc := consumer.NewNATSConsumer(natsConn.JS, natsConn.Conn, log, workerCfg.Consumer.ShutdownTimeout)
	deploymentRequest := workerService.NewDeploymentRequestService(deploymentRequestRepo, k8sDeployment, log)
	worker.SetupRouter(nc, &workerCfg.Consumer, deploymentRequest, log)

	// Start consuming
	if err := nc.Run(); err != nil {
		log.Fatal("Failed to start consumer", zap.Error(err))
	}

	log.Info("Consumer started", zap.String("channel", prod.DeploymentRequestChannel))

	// Defer shutdown - runs when main returns (after WaitForShutdown)
	defer nc.Shutdown()

	// Block until shutdown signal
	utils.WaitForShutdown()
}
