package watcher

import (
	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	"github.com/code-xd/k8s-deployment-manager/pkg/routedinformer"
	portswatcher "github.com/code-xd/k8s-deployment-manager/pkg/ports/service/watcherService"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

// NewDeploymentInformer creates a deployment informer that filters by managed-by (value from config),
// uses resync period and task timeout from config, and wires the given handler to deployment events.
// The returned RoutedInformer should be Run() by the caller.
func NewDeploymentInformer(
	cfg *dto.WorkerConfig,
	clientset kubernetes.Interface,
	watcherService portswatcher.WatcherService,
	log *zap.Logger,
) *routedinformer.RoutedInformer {
	handler := NewHandler(watcherService, log)
	opts := []routedinformer.Option{
		routedinformer.WithResyncPeriod(cfg.Watcher.ResyncPeriod),
		routedinformer.WithTaskTimeout(cfg.Watcher.TaskTimeout),
		routedinformer.WithTweakListOptions(routedinformer.LabelSelectorTweak(map[string]interface{}{
			dto.LabelKeyManagedBy: cfg.K8s.ManagerTag,
		})),
		routedinformer.WithLogger(log),
	}
	return routedinformer.NewRoutedInformer(clientset, handler.Handle, opts...)
}
