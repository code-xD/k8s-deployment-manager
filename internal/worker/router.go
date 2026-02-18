package worker

import (
	"github.com/code-xd/k8s-deployment-manager/internal/worker/handler"
	"github.com/code-xd/k8s-deployment-manager/pkg/consumer"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	portsworker "github.com/code-xd/k8s-deployment-manager/pkg/ports/service/workerService"
	"go.uber.org/zap"
)

// SetupRouter registers all consumer routes on the NATS consumer.
// Handler initialization and route registration happen here, similar to the API stack.
func SetupRouter(
	nc *consumer.NATSConsumer,
	cfg *dto.ConsumerConfig,
	deploymentRequest portsworker.DeploymentRequest,
	deploymentUpdate portsworker.DeploymentUpdate,
	log *zap.Logger,
) {
	registerDeploymentRequestRoutes(nc, cfg, deploymentRequest, log)
	registerDeploymentUpdateRoutes(nc, cfg, deploymentUpdate, log)
}

// registerDeploymentRequestRoutes initializes the deployment request handler and registers its route
func registerDeploymentRequestRoutes(
	nc *consumer.NATSConsumer,
	cfg *dto.ConsumerConfig,
	deploymentRequest portsworker.DeploymentRequest,
	log *zap.Logger,
) {
	deploymentRequestHandler := handler.NewDeploymentRequestHandler(deploymentRequest, log)

	taskCfg := cfg.DeploymentRequestTask
	channel := taskCfg.Channel
	queueGroup := taskCfg.QueueGroup
	if queueGroup == "" {
		queueGroup = dto.QueueGroupDeploymentWorkers
	}

	opts := []consumer.OptionFunc{}
	if taskCfg.TaskTimeout != nil && *taskCfg.TaskTimeout > 0 {
		opts = append(opts, consumer.TaskTimeout(*taskCfg.TaskTimeout))
	}
	if taskCfg.RetryCount != nil && *taskCfg.RetryCount >= 0 {
		opts = append(opts, consumer.RetryCount(*taskCfg.RetryCount))
	}

	nc.Route(channel, queueGroup, deploymentRequestHandler.Handle, opts...)
}

// registerDeploymentUpdateRoutes initializes the deployment update handler and registers its route
func registerDeploymentUpdateRoutes(
	nc *consumer.NATSConsumer,
	cfg *dto.ConsumerConfig,
	deploymentUpdate portsworker.DeploymentUpdate,
	log *zap.Logger,
) {
	deploymentUpdateHandler := handler.NewDeploymentUpdateHandler(deploymentUpdate, log)

	taskCfg := cfg.DeploymentUpdateTask
	channel := taskCfg.Channel
	queueGroup := taskCfg.QueueGroup
	if queueGroup == "" {
		queueGroup = dto.QueueGroupDeploymentWorkers
	}

	opts := []consumer.OptionFunc{}
	if taskCfg.TaskTimeout != nil && *taskCfg.TaskTimeout > 0 {
		opts = append(opts, consumer.TaskTimeout(*taskCfg.TaskTimeout))
	}
	if taskCfg.RetryCount != nil && *taskCfg.RetryCount >= 0 {
		opts = append(opts, consumer.RetryCount(*taskCfg.RetryCount))
	}

	nc.Route(channel, queueGroup, deploymentUpdateHandler.Handle, opts...)
}
