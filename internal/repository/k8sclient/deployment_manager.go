package k8sclient

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto/models"
	"github.com/code-xd/k8s-deployment-manager/pkg/utils"
	"github.com/go-viper/mapstructure/v2"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// DeploymentManager handles Kubernetes deployment operations.
type DeploymentManager struct {
	templatesBasePath string
	clientset         *kubernetes.Clientset
	logger            *zap.Logger
	managerTag        string
}

// NewDeploymentManager creates a new DeploymentManager.
// templatesBasePath is the directory containing the templates folder (e.g. project root or ".").
// cfg controls whether to use in-cluster config or kubeconfig. If nil, in-cluster is used.
func NewDeploymentManager(templatesBasePath string, cfg *dto.K8sConfig, logger *zap.Logger) (*DeploymentManager, error) {
	restConfig, err := buildRestConfig(cfg)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("kubernetes client: %w", err)
	}

	basePath, err := filepath.Abs(templatesBasePath)
	if err != nil {
		return nil, fmt.Errorf("resolve templates path: %w", err)
	}

	if cfg == nil || cfg.ManagerTag == "" {
		return nil, fmt.Errorf("%s", dto.ErrMsgK8sManagerTagRequired)
	}

	return &DeploymentManager{
		templatesBasePath: basePath,
		clientset:         clientset,
		logger:            logger,
		managerTag:        cfg.ManagerTag,
	}, nil
}

func buildRestConfig(cfg *dto.K8sConfig) (*rest.Config, error) {
	if cfg == nil || cfg.InCluster {
		return rest.InClusterConfig()
	}
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if cfg.Kubeconfig != "" {
		loadingRules.ExplicitPath = cfg.Kubeconfig
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
}

// NewClientSet returns a Kubernetes clientset for the given config (e.g. for use with informers).
func NewClientSet(cfg *dto.K8sConfig) (kubernetes.Interface, error) {
	restConfig, err := buildRestConfig(cfg)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(restConfig)
}

// Create fetches the template for the image, replaces placeholders with DeploymentRequest
// details, validates the manifest, and creates the deployment in Kubernetes.
// When metadata contains inline HTML (keys: "html", "content", or "body"), a ConfigMap is created and mounted into the nginx container.
func (dm *DeploymentManager) Create(ctx context.Context, req *models.DeploymentRequest) (*appsv1.Deployment, error) {
	renderer := utils.NewTemplateRenderer[dto.CreateTemplateData](dm.templatesBasePath, req.Image)
	if renderer.TemplateName() != dto.TemplateNginx {
		return nil, fmt.Errorf("unsupported image: only nginx is supported, got %q", req.Image)
	}

	if err := renderer.Load(); err != nil {
		return nil, fmt.Errorf("load template: %w", err)
	}

	indexHTML := dm.extractIndexHTML(req.Metadata)

	data := dto.CreateTemplateData{
		Name:                req.Name,
		Namespace:           req.Namespace,
		Identifier:          req.Identifier,
		Image:               req.Image,
		UserID:              req.UserID.String(),
		RequestID:           req.RequestID,
		DeploymentRequestID: req.ID.String(),
		HasCustomHTML:       indexHTML != "",
		ManagedBy:           dm.managerTag,
	}

	manifest, err := renderer.Execute(data)
	if err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	deployment, err := dm.parseAndValidate(manifest)
	if err != nil {
		return nil, fmt.Errorf("parse and validate: %w", err)
	}

	if err := dm.getOrCreateNamespace(ctx, req.Namespace); err != nil {
		return nil, fmt.Errorf("failed to get or create namespace: %w", err)
	}

	if indexHTML != "" {
		configMap := dm.buildHTMLConfigMap(req.Identifier, req.Namespace, indexHTML)
		if _, err := dm.clientset.CoreV1().ConfigMaps(req.Namespace).Create(ctx, configMap, metav1.CreateOptions{}); err != nil {
			return nil, fmt.Errorf("create configmap for %s: %w", dto.ConfigMapIndexHTML, err)
		}
	}

	created, err := dm.clientset.AppsV1().Deployments(req.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("create deployment in cluster: %w", err)
	}

	return created, nil
}

func (dm *DeploymentManager) getOrCreateNamespace(ctx context.Context, namespace string) error {
	_, err := dm.clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil || !apierrors.IsNotFound(err) {
		return err
	}

	_, err = dm.clientset.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: namespace},
	}, metav1.CreateOptions{})

	return err
}

// parseAndValidate decodes the YAML manifest into an appsv1.Deployment and validates it.
func (dm *DeploymentManager) parseAndValidate(manifest string) (*appsv1.Deployment, error) {
	var depl appsv1.Deployment
	if err := yaml.Unmarshal([]byte(manifest), &depl); err != nil {
		return nil, fmt.Errorf("yaml unmarshal: %w", err)
	}

	if depl.Name == "" {
		return nil, fmt.Errorf("deployment name is required")
	}
	if depl.Namespace == "" {
		depl.Namespace = corev1.NamespaceDefault
	}
	if len(depl.Spec.Template.Spec.Containers) == 0 {
		return nil, fmt.Errorf("deployment must have at least one container")
	}
	if depl.Spec.Template.Spec.Containers[0].Image == "" {
		return nil, fmt.Errorf("container image is required")
	}

	return &depl, nil
}

// extractIndexHTML returns the inline HTML content from metadata if present.
// Supports keys: "html", "content", "body"
func (dm *DeploymentManager) extractIndexHTML(metadata models.JSONB) string {
	if metadata == nil {
		return ""
	}

	var deploymentMetadata dto.DeploymentMetadata
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &deploymentMetadata,
		TagName: dto.MapstructureTagJSON,
	})
	if err != nil {
		dm.logger.Error("failed to decode deployment metadata", zap.Error(err))
		return ""
	}

	err = decoder.Decode(metadata)
	if err != nil {
		dm.logger.Error("failed to decode deployment metadata", zap.Error(err))
		return ""
	}

	return deploymentMetadata.DocHTML
}

// Get retrieves a deployment from Kubernetes by namespace and name
func (dm *DeploymentManager) Get(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	deployment, err := dm.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get deployment from cluster: %w", err)
	}
	return deployment, nil
}

// GetOptional returns the deployment if it exists in the cluster.
// It returns (nil, false, nil) when the deployment is not found (NotFound error).
func (dm *DeploymentManager) GetOptional(ctx context.Context, namespace, name string) (*appsv1.Deployment, bool, error) {
	deployment, err := dm.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("get deployment from cluster: %w", err)
	}
	return deployment, true, nil
}

// buildHTMLConfigMap creates a ConfigMap with index.html content for nginx to serve.
func (dm *DeploymentManager) buildHTMLConfigMap(identifier, namespace, indexHTML string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      identifier + dto.ConfigMapHTMLSuffix,
			Namespace: namespace,
		},
		Data: map[string]string{
			dto.ConfigMapIndexHTML: indexHTML,
		},
	}
}

// Update updates an existing deployment in Kubernetes based on the deployment request metadata.
// It applies changes to replica count, resource limits, and doc_html (ConfigMap) if provided.
func (dm *DeploymentManager) Update(ctx context.Context, req *models.DeploymentRequest, existingDeployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	// Make a copy to avoid modifying the original
	updatedDeployment := existingDeployment.DeepCopy()

	// Extract metadata from request
	var updateMetadata dto.UpdateDeploymentRequestMetadata
	if req.Metadata != nil {
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Result:  &updateMetadata,
			TagName: dto.MapstructureTagJSON,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create decoder: %w", err)
		}
		if err := decoder.Decode(req.Metadata); err != nil {
			return nil, fmt.Errorf("failed to decode update metadata: %w", err)
		}
	}

	// Apply updates based on provided metadata fields
	if updateMetadata.ReplicaCount != nil {
		if err := dm.updateReplicaCount(updatedDeployment, *updateMetadata.ReplicaCount); err != nil {
			return nil, fmt.Errorf("update replica count: %w", err)
		}
	}

	if updateMetadata.ResourceLimit != nil {
		if err := dm.updateResourceLimits(updatedDeployment, updateMetadata.ResourceLimit); err != nil {
			return nil, fmt.Errorf("update resource limits: %w", err)
		}
	}

	if updateMetadata.DocHTML != nil {
		if err := dm.updateConfigMap(ctx, req, *updateMetadata.DocHTML); err != nil {
			return nil, fmt.Errorf("update configmap: %w", err)
		}
	}

	// Update the deployment in Kubernetes
	updated, err := dm.clientset.AppsV1().Deployments(req.Namespace).Update(ctx, updatedDeployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("update deployment in cluster: %w", err)
	}

	return updated, nil
}

// updateReplicaCount updates the replica count of a deployment.
func (dm *DeploymentManager) updateReplicaCount(deployment *appsv1.Deployment, replicaCount int) error {
	replicas := int32(replicaCount)
	deployment.Spec.Replicas = &replicas
	return nil
}

// updateResourceLimits updates the resource limits and requests of the first container in a deployment.
func (dm *DeploymentManager) updateResourceLimits(deployment *appsv1.Deployment, resourceLimit *dto.ResourceMetadata) error {
	if len(deployment.Spec.Template.Spec.Containers) == 0 {
		return fmt.Errorf("deployment has no containers")
	}

	container := &deployment.Spec.Template.Spec.Containers[0]

	// Parse CPU and memory for requests
	requestCPU, err := resource.ParseQuantity(resourceLimit.Request.CPU)
	if err != nil {
		return fmt.Errorf("invalid CPU request value: %w", err)
	}
	requestMemory, err := resource.ParseQuantity(resourceLimit.Request.Memory)
	if err != nil {
		return fmt.Errorf("invalid memory request value: %w", err)
	}

	// Parse CPU and memory for limits
	limitCPU, err := resource.ParseQuantity(resourceLimit.Limit.CPU)
	if err != nil {
		return fmt.Errorf("invalid CPU limit value: %w", err)
	}
	limitMemory, err := resource.ParseQuantity(resourceLimit.Limit.Memory)
	if err != nil {
		return fmt.Errorf("invalid memory limit value: %w", err)
	}

	container.Resources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    requestCPU,
			corev1.ResourceMemory: requestMemory,
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    limitCPU,
			corev1.ResourceMemory: limitMemory,
		},
	}

	return nil
}

// updateConfigMap updates or creates a ConfigMap with HTML content for the deployment.
func (dm *DeploymentManager) updateConfigMap(ctx context.Context, req *models.DeploymentRequest, docHTML string) error {
	if docHTML == "" {
		// Empty doc_html means no update needed
		return nil
	}

	configMapName := req.Identifier + dto.ConfigMapHTMLSuffix
	configMap := dm.buildHTMLConfigMap(req.Identifier, req.Namespace, docHTML)

	// Try to update existing ConfigMap, or create if it doesn't exist
	_, err := dm.clientset.CoreV1().ConfigMaps(req.Namespace).Update(ctx, configMap, metav1.UpdateOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Create if it doesn't exist
			_, err = dm.clientset.CoreV1().ConfigMaps(req.Namespace).Create(ctx, configMap, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create configmap %s: %w", configMapName, err)
			}
		} else {
			return fmt.Errorf("failed to update configmap %s: %w", configMapName, err)
		}
	}

	return nil
}

// Delete deletes a deployment from Kubernetes by namespace and name.
func (dm *DeploymentManager) Delete(ctx context.Context, namespace, name string) error {
	deletePolicy := metav1.DeletePropagationForeground
	err := dm.clientset.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Deployment already deleted, consider it success
			return nil
		}
		return fmt.Errorf("delete deployment from cluster: %w", err)
	}
	return nil
}
