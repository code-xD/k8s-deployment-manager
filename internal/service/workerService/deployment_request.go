package workerService

import (
	"context"
	"fmt"

	"github.com/code-xd/k8s-deployment-manager/pkg/consumer"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto/models"
	portsdb "github.com/code-xd/k8s-deployment-manager/pkg/ports/repo/db"
	portsk8s "github.com/code-xd/k8s-deployment-manager/pkg/ports/repo/k8s"
	portsworker "github.com/code-xd/k8s-deployment-manager/pkg/ports/service/workerService"
	"github.com/code-xd/k8s-deployment-manager/pkg/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DeploymentRequestService implements worker-side deployment request processing
type DeploymentRequestService struct {
	deploymentRequestRepo portsdb.DeploymentRequest
	k8sDeploymentManager  portsk8s.DeploymentManager
	logger                *zap.Logger
}

// NewDeploymentRequestService creates a new worker deployment request service
func NewDeploymentRequestService(
	deploymentRequestRepo portsdb.DeploymentRequest,
	k8sDeploymentManager portsk8s.DeploymentManager,
	logger *zap.Logger,
) portsworker.DeploymentRequest {
	return &DeploymentRequestService{
		deploymentRequestRepo: deploymentRequestRepo,
		k8sDeploymentManager:  k8sDeploymentManager,
		logger:                logger,
	}
}

// ProcessDeploymentRequest fetches the deployment request, validates state and ownership,
// then dispatches to the appropriate handler based on request type.
func (s *DeploymentRequestService) ProcessDeploymentRequest(ctx context.Context, msg *dto.DeploymentRequestMessage) error {
	req, found, err := s.deploymentRequestRepo.GetByRequestID(ctx, msg.RequestID)
	if err != nil {
		return fmt.Errorf("fetch deployment request: %w", err)
	}
	if !found {
		return fmt.Errorf("deployment request not found: request_id=%s", msg.RequestID)
	}

	if req.Status != models.DeploymentRequestStatusCreated {
		return fmt.Errorf("deployment request is in terminal or invalid state: status=%s", req.Status)
	}

	headerUserID := utils.GetUserIDFromWorkerHeader(consumer.HeadersFromContext(ctx))
	if headerUserID == uuid.Nil || headerUserID != req.UserID {
		return fmt.Errorf("user_id from header does not match deployment request owner")
	}

	lastRetryAttempt := consumer.LastAttemptFromContext(ctx)

	switch req.RequestType {
	case models.DeploymentRequestTypeCreate:
		return s.processCreate(ctx, req, lastRetryAttempt)
	case models.DeploymentRequestTypeUpdate:
		return s.processUpdate(ctx, req, lastRetryAttempt)
	case models.DeploymentRequestTypeDelete:
		return s.processDelete(ctx, req, lastRetryAttempt)
	default:
		return fmt.Errorf("unknown request type: %s", req.RequestType)
	}
}

// processCreate invokes k8s deployment creation and updates the deployment request status.
func (s *DeploymentRequestService) processCreate(ctx context.Context, req *models.DeploymentRequest, lastRetryAttempt bool) error {
	_, err := s.k8sDeploymentManager.Create(ctx, req)
	if err != nil {
		if lastRetryAttempt {
			errMsg := err.Error()
			if updateErr := s.deploymentRequestRepo.UpdateStatus(ctx, req.ID, models.DeploymentRequestStatusFailure, &errMsg); updateErr != nil {
				s.logger.Error("Failed to mark deployment request as FAILURE", zap.Error(updateErr))
			}
		}
		return fmt.Errorf("create deployment: %w", err)
	}

	if err := s.deploymentRequestRepo.UpdateStatus(ctx, req.ID, models.DeploymentRequestStatusSuccess, nil); err != nil {
		return fmt.Errorf("update status to SUCCESS: %w", err)
	}
	return nil
}

// processUpdate invokes k8s deployment update and updates the deployment request status.
func (s *DeploymentRequestService) processUpdate(ctx context.Context, req *models.DeploymentRequest, lastRetryAttempt bool) error {
	// Get existing deployment from K8s
	existingDeployment, found, err := s.k8sDeploymentManager.GetOptional(ctx, req.Namespace, req.Identifier)
	if err != nil {
		if lastRetryAttempt {
			errMsg := fmt.Sprintf("failed to get existing deployment: %v", err)
			if updateErr := s.deploymentRequestRepo.UpdateStatus(ctx, req.ID, models.DeploymentRequestStatusFailure, &errMsg); updateErr != nil {
				s.logger.Error("Failed to mark deployment request as FAILURE", zap.Error(updateErr))
			}
		}
		return fmt.Errorf("get existing deployment: %w", err)
	}
	if !found {
		if lastRetryAttempt {
			errMsg := fmt.Sprintf("deployment not found in Kubernetes: namespace=%s, name=%s", req.Namespace, req.Name)
			if updateErr := s.deploymentRequestRepo.UpdateStatus(ctx, req.ID, models.DeploymentRequestStatusFailure, &errMsg); updateErr != nil {
				s.logger.Error("Failed to mark deployment request as FAILURE", zap.Error(updateErr))
			}
		}
		return fmt.Errorf("deployment not found in Kubernetes: namespace=%s, name=%s", req.Namespace, req.Name)
	}

	// Update the deployment in K8s
	_, err = s.k8sDeploymentManager.Update(ctx, req, existingDeployment)
	if err != nil {
		if lastRetryAttempt {
			errMsg := err.Error()
			if updateErr := s.deploymentRequestRepo.UpdateStatus(ctx, req.ID, models.DeploymentRequestStatusFailure, &errMsg); updateErr != nil {
				s.logger.Error("Failed to mark deployment request as FAILURE", zap.Error(updateErr))
			}
		}
		return fmt.Errorf("update deployment: %w", err)
	}

	if err := s.deploymentRequestRepo.UpdateStatus(ctx, req.ID, models.DeploymentRequestStatusSuccess, nil); err != nil {
		return fmt.Errorf("update status to SUCCESS: %w", err)
	}
	return nil
}

// processDelete invokes k8s deployment deletion and updates the deployment request status.
func (s *DeploymentRequestService) processDelete(ctx context.Context, req *models.DeploymentRequest, lastRetryAttempt bool) error {
	// Delete the deployment from K8s
	err := s.k8sDeploymentManager.Delete(ctx, req.Namespace, req.Identifier)
	if err != nil {
		if lastRetryAttempt {
			errMsg := err.Error()
			if updateErr := s.deploymentRequestRepo.UpdateStatus(ctx, req.ID, models.DeploymentRequestStatusFailure, &errMsg); updateErr != nil {
				s.logger.Error("Failed to mark deployment request as FAILURE", zap.Error(updateErr))
			}
		}
		return fmt.Errorf("delete deployment: %w", err)
	}

	if err := s.deploymentRequestRepo.UpdateStatus(ctx, req.ID, models.DeploymentRequestStatusSuccess, nil); err != nil {
		return fmt.Errorf("update status to SUCCESS: %w", err)
	}
	return nil
}
