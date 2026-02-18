package routedinformer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/informers/internalinterfaces"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// DeploymentEventHandler is called for each deployment event that passes all filters.
// EventType indicates add/update/delete. The context is cancelled after the configured task timeout.
// For delete events, deployment may be nil if the object was already removed from the cache (e.g. DeletedFinalStateUnknown).
type DeploymentEventHandler func(ctx context.Context, eventType DeploymentEventType, deployment *appsv1.Deployment)

// RoutedInformer watches deployments; optional list-option tweaks (before Run) and in-process filters apply. Use WithTweakListOptions for server-side filtering (e.g. label selector).
type RoutedInformer struct {
	clientset        kubernetes.Interface
	resyncPeriod     time.Duration
	taskTimeout      time.Duration
	handler          DeploymentEventHandler
	logger           *zap.Logger
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	informer         cache.SharedInformer
	stopCh           chan struct{}
}

// NewRoutedInformer creates an informer for deployments. Use WithTweakListOptions (e.g. label selector) for server-side filtering before Run; optional WithFilter for in-process filtering.
func NewRoutedInformer(
	clientset kubernetes.Interface,
	handler DeploymentEventHandler,
	opts ...Option,
) *RoutedInformer {
	if handler == nil {
		panic("routedinformer: handler is required")
	}

	ri := &RoutedInformer{
		clientset:    clientset,
		resyncPeriod: 0,
		taskTimeout:  0,
		handler:      handler,
		logger:       zap.NewNop(),
	}

	for _, opt := range opts {
		opt(ri)
	}

	var factory informers.SharedInformerFactory
	if ri.tweakListOptions != nil {
		tweak := ri.tweakListOptions
		factory = informers.NewSharedInformerFactoryWithOptions(
			clientset,
			ri.resyncPeriod,
			informers.WithTweakListOptions(func(opts *metav1.ListOptions) { tweak(opts) }),
		)
	} else {
		factory = informers.NewSharedInformerFactory(clientset, ri.resyncPeriod)
	}

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
			namespace, name := ri.getMetaNamespaceKey(obj)
			d := &appsv1.Deployment{}
			d.SetName(name)
			d.SetNamespace(namespace)
			ri.dispatchDeployment(DeploymentEventDelete, d)
		},
	}
}

// getMetaNamespaceKey returns namespace and name from obj using cache.MetaNamespaceKeyFunc.
// Key is "namespace/name" for namespaced resources, or "name" for cluster-scoped.
func (ri *RoutedInformer) getMetaNamespaceKey(obj interface{}) (namespace, name string) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil || key == "" {
		return "", ""
	}
	parts := strings.SplitN(key, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", parts[0]
}

func (ri *RoutedInformer) dispatch(eventType DeploymentEventType, obj interface{}) {
	d := ri.toDeployment(obj)
	if d == nil {
		return
	}
	ri.dispatchDeployment(eventType, d)
}

func (ri *RoutedInformer) dispatchDeployment(eventType DeploymentEventType, d *appsv1.Deployment) {
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
