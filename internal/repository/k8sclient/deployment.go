package k8sclient

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto/models"
	"github.com/code-xd/k8s-deployment-manager/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Deployment handles Kubernetes deployment operations.
type Deployment struct {
	templatesBasePath string
	clientset         *kubernetes.Clientset
}

// NewDeployment creates a new Deployment repository.
// templatesBasePath is the directory containing the templates folder (e.g. project root or ".").
// cfg controls whether to use in-cluster config or kubeconfig. If nil, in-cluster is used.
func NewDeployment(templatesBasePath string, cfg *dto.K8sConfig) (*Deployment, error) {
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

	return &Deployment{
		templatesBasePath: basePath,
		clientset:         clientset,
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

// Create fetches the template for the image, replaces placeholders with DeploymentRequest
// details, validates the manifest, and creates the deployment in Kubernetes.
// When metadata contains inline HTML (keys: "html", "content", or "body"), a ConfigMap is created and mounted into the nginx container.
func (d *Deployment) Create(ctx context.Context, req *models.DeploymentRequest) (*appsv1.Deployment, error) {
	renderer := utils.NewTemplateRenderer[dto.CreateTemplateData](d.templatesBasePath, req.Image)
	if renderer.TemplateName() != "nginx" {
		return nil, fmt.Errorf("unsupported image: only nginx is supported, got %q", req.Image)
	}

	if err := renderer.Load(); err != nil {
		return nil, fmt.Errorf("load template: %w", err)
	}

	indexHTML := d.extractIndexHTML(req.Metadata)
	if indexHTML != "" {
		configMap := d.buildHTMLConfigMap(req.Name, req.Namespace, indexHTML)
		if _, err := d.clientset.CoreV1().ConfigMaps(req.Namespace).Create(ctx, configMap, metav1.CreateOptions{}); err != nil {
			return nil, fmt.Errorf("create configmap for index.html: %w", err)
		}
	}

	data := dto.CreateTemplateData{
		Name:                req.Name,
		Namespace:           req.Namespace,
		Identifier:          req.Identifier,
		Image:               req.Image,
		UserID:              req.UserID.String(),
		RequestID:           req.RequestID,
		DeploymentRequestID: req.ID.String(),
		HasCustomHTML:       indexHTML != "",
	}

	manifest, err := renderer.Execute(data)
	if err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	deployment, err := d.parseAndValidate(manifest)
	if err != nil {
		return nil, fmt.Errorf("parse and validate: %w", err)
	}

	if err := d.getOrCreateNamespace(ctx, req.Namespace); err != nil {
		return nil, fmt.Errorf("failed to get or create namespace: %w", err)
	}

	created, err := d.clientset.AppsV1().Deployments(req.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("create deployment in cluster: %w", err)
	}

	return created, nil
}

func (d *Deployment) getOrCreateNamespace(ctx context.Context, namespace string) error {
	_, err := d.clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil || !apierrors.IsNotFound(err) {
		return err
	}

	_, err = d.clientset.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: namespace},
	}, metav1.CreateOptions{})

	return err
}

// parseAndValidate decodes the YAML manifest into an appsv1.Deployment and validates it.
func (d *Deployment) parseAndValidate(manifest string) (*appsv1.Deployment, error) {
	var deployment appsv1.Deployment
	if err := yaml.Unmarshal([]byte(manifest), &deployment); err != nil {
		return nil, fmt.Errorf("yaml unmarshal: %w", err)
	}

	if deployment.Name == "" {
		return nil, fmt.Errorf("deployment name is required")
	}
	if deployment.Namespace == "" {
		deployment.Namespace = corev1.NamespaceDefault
	}
	if len(deployment.Spec.Template.Spec.Containers) == 0 {
		return nil, fmt.Errorf("deployment must have at least one container")
	}
	if deployment.Spec.Template.Spec.Containers[0].Image == "" {
		return nil, fmt.Errorf("container image is required")
	}

	return &deployment, nil
}

// extractIndexHTML returns the inline HTML content from metadata if present.
// Supports keys: "html", "content", "body"
func (d *Deployment) extractIndexHTML(metadata models.JSONB) string {
	if metadata == nil {
		return ""
	}
	for _, key := range []string{"html", "content", "body"} {
		if v, ok := metadata[key]; ok && v != nil {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

// buildHTMLConfigMap creates a ConfigMap with index.html content for nginx to serve.
func (d *Deployment) buildHTMLConfigMap(name, namespace, indexHTML string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-html",
			Namespace: namespace,
		},
		Data: map[string]string{
			"index.html": indexHTML,
		},
	}
}
