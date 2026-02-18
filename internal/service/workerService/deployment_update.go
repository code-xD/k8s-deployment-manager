package workerService

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto/models"
	portsdb "github.com/code-xd/k8s-deployment-manager/pkg/ports/repo/db"
	portsk8s "github.com/code-xd/k8s-deployment-manager/pkg/ports/repo/k8s"
	portsworker "github.com/code-xd/k8s-deployment-manager/pkg/ports/service/workerService"
	"github.com/google/uuid"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
)

// DeploymentUpdateService implements worker-side deployment update processing
type DeploymentUpdateService struct {
	deploymentRepo       portsdb.Deployment
	k8sDeploymentManager portsk8s.DeploymentManager
	logger               *zap.Logger
}

// NewDeploymentUpdateService creates a new worker deployment update service
func NewDeploymentUpdateService(
	deploymentRepo portsdb.Deployment,
	k8sDeploymentManager portsk8s.DeploymentManager,
	logger *zap.Logger,
) portsworker.DeploymentUpdate {
	return &DeploymentUpdateService{
		deploymentRepo:       deploymentRepo,
		k8sDeploymentManager: k8sDeploymentManager,
		logger:               logger,
	}
}

// ProcessDeploymentUpdate processes a deployment update message:
// 1. Fetches from both DB (by identifier) and K8s (by namespace/name), with error checks.
// 2. If not in DB and not in K8s → return as is.
// 3. If in DB and not in K8s → mark as deleted.
// 4. Else (in K8s) → extract metadata and upsert as usual.
func (s *DeploymentUpdateService) ProcessDeploymentUpdate(ctx context.Context, msg *dto.DeploymentUpdateMessage) error {
	// Parse identifier (format: namespace/name)
	parts := strings.Split(msg.Identifier, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid identifier format, expected namespace/name: %s", msg.Identifier)
	}
	namespace := parts[0]
	name := parts[1]

	// Fetch from both DB and K8s
	dbDeployment, dbExists, err := s.deploymentRepo.GetByIdentifier(ctx, name)
	if err != nil {
		return fmt.Errorf("get deployment by identifier: %w", err)
	}
	k8sDeployment, k8sExists, err := s.k8sDeploymentManager.GetOptional(ctx, namespace, name)
	if err != nil {
		return fmt.Errorf("get deployment from k8s: %w", err)
	}

	if !dbExists && !k8sExists {
		return nil
	}
	if dbExists && !k8sExists {
		return s.markDeploymentDeleted(ctx, dbDeployment, msg.Identifier)
	}

	// Usual flow: in K8s — extract metadata and upsert
	deployment, err := s.extractDeploymentFromK8s(k8sDeployment)
	if err != nil {
		return fmt.Errorf("extract deployment from k8s object: %w", err)
	}
	if err := s.deploymentRepo.Upsert(ctx, deployment); err != nil {
		return fmt.Errorf("upsert deployment: %w", err)
	}
	s.logger.Info("Processed deployment update",
		zap.String("identifier", deployment.Identifier),
		zap.String("resource_version", deployment.ResourceVersion),
		zap.String("status", string(deployment.Status)),
	)
	return nil
}

// markDeploymentDeleted sets deployment status to DELETED and updates the DB when the deployment exists in DB but not in K8s.
func (s *DeploymentUpdateService) markDeploymentDeleted(ctx context.Context, dbDeployment *models.Deployment, identifier string) error {
	if dbDeployment.Status == models.DeploymentStatusDeleted {
		return nil
	}
	now := time.Now()
	dbDeployment.Status = models.DeploymentStatusDeleted
	dbDeployment.UpdatedOn = &now
	if err := s.deploymentRepo.Update(ctx, dbDeployment); err != nil {
		return fmt.Errorf("update deployment status to deleted: %w", err)
	}
	s.logger.Info("Marked deployment as deleted (not found in k8s)",
		zap.String("identifier", identifier),
	)
	return nil
}

// extractDeploymentFromK8s extracts deployment fields from Kubernetes deployment object
func (s *DeploymentUpdateService) extractDeploymentFromK8s(k8sDeployment *appsv1.Deployment) (*models.Deployment, error) {
	now := time.Now()
	deployment := &models.Deployment{
		Common: models.Common{
			UpdatedOn: &now,
		},
	}

	// Extract identifier from labels
	identifier, ok := k8sDeployment.Labels["identifier"]
	if !ok {
		return nil, fmt.Errorf("deployment missing identifier label")
	}
	deployment.Identifier = identifier

	// Extract user_id from labels
	userIDStr, ok := k8sDeployment.Labels["user-id"]
	if !ok {
		return nil, fmt.Errorf("deployment missing user-id label")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user-id format: %w", err)
	}
	deployment.UserID = userID

	// Extract name and namespace from metadata
	deployment.Name, ok = k8sDeployment.Labels["name"]
	if !ok {
		return nil, fmt.Errorf("deployment missing name label")
	}

	deployment.Namespace = k8sDeployment.Namespace

	// Extract resourceVersion from metadata
	deployment.ResourceVersion = k8sDeployment.ResourceVersion

	// Extract image from first container
	if len(k8sDeployment.Spec.Template.Spec.Containers) > 0 {
		deployment.Image = k8sDeployment.Spec.Template.Spec.Containers[0].Image
	}

	// Determine status from deployment conditions
	deployment.Status = s.determineStatus(k8sDeployment)

	// dump relevant metadata to deployment.Metadata
	deployment.Metadata = models.JSONB{
		"k8s_deployment": k8sDeployment,
	}

	return deployment, nil
}

// determineStatus determines the deployment status from Kubernetes deployment conditions
func (s *DeploymentUpdateService) determineStatus(k8sDeployment *appsv1.Deployment) models.DeploymentStatus {
	// Check if deployment is being deleted
	if k8sDeployment.DeletionTimestamp != nil {
		return models.DeploymentStatusDeleted
	}

	// Check if deployment has ready replicas
	if k8sDeployment.Status.ReadyReplicas > 0 && k8sDeployment.Status.ReadyReplicas == k8sDeployment.Status.Replicas {
		return models.DeploymentStatusCreated
	}

	// Check deployment conditions for progress
	for _, condition := range k8sDeployment.Status.Conditions {
		if condition.Type == appsv1.DeploymentProgressing {
			if condition.Status == "True" {
				// Deployment is progressing but may not be fully ready
				return models.DeploymentStatusUpdating
			}
		}
	}

	// Default to initiated if we can't determine status
	return models.DeploymentStatusInitiated
}
