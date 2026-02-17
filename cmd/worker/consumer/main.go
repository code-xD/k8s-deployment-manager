package main

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/code-xd/k8s-deployment-manager/internal/repository/nats/common"
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

	// Initialize NATS connection
	natsConn, err := common.NewNATS(&workerCfg.Nats, log)
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

	// Create consumer
	nc := consumer.NewNATSConsumer(natsConn.JS, natsConn.Conn, log)

	// Wire services and register routes via router (like API stack)
	deploymentRequest := workerService.NewDeploymentRequestService(log)
	worker.SetupRouter(nc, &workerCfg.Consumer, deploymentRequest, log)

	// Start consuming
	if err := nc.Run(); err != nil {
		log.Fatal("Failed to start consumer", zap.Error(err))
	}

	log.Info("Consumer started", zap.String("channel", prod.DeploymentRequestChannel))

	// Defer shutdown - runs when main returns (after WaitForShutdown)
	defer func() {
		timeout := workerCfg.Consumer.ShutdownTimeout
		if timeout == 0 {
			timeout = 30 * time.Second
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := nc.Shutdown(ctx); err != nil {
			log.Warn("Consumer shutdown error", zap.Error(err))
		} else {
			log.Info("Consumer shutdown complete")
		}
	}()

	// Block until shutdown signal
	utils.WaitForShutdown()
}
