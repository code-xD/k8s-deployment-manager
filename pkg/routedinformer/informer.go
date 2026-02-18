package routedinformer

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// DeploymentEventHandler is called for each deployment event that passes all filters.
// EventType indicates add/update/delete. The context is cancelled after the configured task timeout.
// For delete events, deployment may be nil if the object was already removed from the cache (e.g. DeletedFinalStateUnknown).
type DeploymentEventHandler func(ctx context.Context, eventType DeploymentEventType, deployment *appsv1.Deployment)

// RoutedInformer watches all deployments, applies stacked filters, and invokes the handler for each event that passes.
// Filtering (e.g. managed-by) is configured at setup via WithFilter. Use Run() to start; handler runs with a task-timeout context.
type RoutedInformer struct {
	clientset    kubernetes.Interface
	resyncPeriod  time.Duration
	taskTimeout   time.Duration
	handler      DeploymentEventHandler
	logger       *zap.Logger
	filters      []Filter
	informer     cache.SharedInformer
	stopCh       chan struct{}
}

// NewRoutedInformer creates an informer that watches all deployments, applies any stacked filters,
// and calls handler for each add/update/delete that passes. Add filters at setup (e.g. LabelFiltering for managed-by).
// Use Run() to start; handler is invoked with a context that times out after the configured task timeout.
func NewRoutedInformer(
	clientset kubernetes.Interface,
	handler DeploymentEventHandler,
	opts ...Option,
) *RoutedInformer {
	if handler == nil {
		panic("routedinformer: handler is required")
	}

	ri := &RoutedInformer{
		clientset:   clientset,
		resyncPeriod: 0,
		taskTimeout:  0,
		handler:     handler,
		logger:      zap.NewNop(),
		filters:     nil,
	}

	for _, opt := range opts {
		opt(ri)
	}

	factory := informers.NewSharedInformerFactory(clientset, ri.resyncPeriod)

	ri.informer = factory.Apps().V1().Deployments().Informer()
	ri.informer.AddEventHandler(ri.resourceEventHandler())
	ri.stopCh = make(chan struct{})

	return ri
}

func (ri *RoutedInformer) resourceEventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ri.dispatch(DeploymentEventAdd, obj)
		},
		UpdateFunc: func(_, newObj interface{}) {
			ri.dispatch(DeploymentEventUpdate, newObj)
		},
		DeleteFunc: func(obj interface{}) {
			d := ri.toDeployment(obj)
			if d == nil {
				if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
					d = ri.toDeployment(tombstone.Obj)
				}
			}
			ri.dispatchDeployment(DeploymentEventDelete, d)
		},
	}
}

func (ri *RoutedInformer) dispatch(eventType DeploymentEventType, obj interface{}) {
	d := ri.toDeployment(obj)
	if d == nil {
		return
	}
	ri.dispatchDeployment(eventType, d)
}

func (ri *RoutedInformer) dispatchDeployment(eventType DeploymentEventType, d *appsv1.Deployment) {
	if d != nil {
		for _, f := range ri.filters {
			if !f(d) {
				return
			}
		}
	}

	ctx := context.Background()
	if ri.taskTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, ri.taskTimeout)
		defer cancel()
	}
	ri.handler(ctx, eventType, d)
}

func (ri *RoutedInformer) toDeployment(obj interface{}) *appsv1.Deployment {
	if obj == nil {
		return nil
	}
	d, ok := obj.(*appsv1.Deployment)
	if !ok {
		ri.logger.Warn("informer received non-deployment object", zap.String("type", fmtType(obj)))
		return nil
	}
	return d
}

func fmtType(obj interface{}) string {
	if obj == nil {
		return "nil"
	}
	return fmt.Sprintf("%T", obj)
}

// Run runs the informer until Stop is called. It blocks.
func (ri *RoutedInformer) Run() {
	ri.logger.Info("starting deployment informer")
	ri.informer.Run(ri.stopCh)
}

// Stop stops the informer. Safe to call multiple times.
func (ri *RoutedInformer) Stop() {
	select {
	case <-ri.stopCh:
		return
	default:
		close(ri.stopCh)
	}
}
