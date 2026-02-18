package main

import (
	"go.uber.org/zap"

	"github.com/code-xd/k8s-deployment-manager/internal/repository/k8sclient"
	"github.com/code-xd/k8s-deployment-manager/internal/repository/nats"
	natscommon "github.com/code-xd/k8s-deployment-manager/internal/repository/nats/common"
	"github.com/code-xd/k8s-deployment-manager/internal/service/watcherService"
	"github.com/code-xd/k8s-deployment-manager/internal/watcher"
	"github.com/code-xd/k8s-deployment-manager/pkg/config"
	"github.com/code-xd/k8s-deployment-manager/pkg/constants"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	"github.com/code-xd/k8s-deployment-manager/pkg/logger"
	"github.com/code-xd/k8s-deployment-manager/pkg/utils"
)

func main() {
	log := logger.New()
	defer log.Sync()
	dto.Log = log

	cfg := config.NewConfigLoader[dto.WorkerConfig](
		constants.DEFAULT_CONFIG_PATH,
		constants.DEFAULT_CONFIG_FILE,
	)
	workerCfg, err := cfg.Load()
	if err != nil {
		log.Fatal("Failed to load config", zap.Error(err))
	}

	natsConn, err := natscommon.NewNATS(&workerCfg.Nats, log)
	if err != nil {
		log.Fatal("Failed to connect to NATS", zap.Error(err))
	}
	defer natsConn.Close()

	prod := &workerCfg.Nats.Producer
	subjects := []string{prod.DeploymentRequestChannel, prod.DeploymentUpdateChannel}
	if err := natsConn.EnsureStream(prod.StreamName, subjects); err != nil {
		log.Fatal("Failed to ensure JetStream stream", zap.Error(err))
	}

	clientset, err := k8sclient.NewClientSet(&workerCfg.K8s)
	if err != nil {
		log.Fatal("Failed to create Kubernetes clientset", zap.Error(err))
	}

	natsProducer := natscommon.NewProducer(natsConn)
	deploymentUpdateProducer := nats.NewDeploymentUpdateProducer(natsProducer, prod)

	watcherSvc := watcherService.NewWatcherService(deploymentUpdateProducer, log)
	informer := watcher.NewDeploymentInformer(workerCfg, clientset, watcherSvc, log)

	go informer.Run()
	defer informer.Stop()

	log.Info("Watcher started", zap.String("managed_by", workerCfg.K8s.ManagerTag))
	utils.WaitForShutdown()
}
