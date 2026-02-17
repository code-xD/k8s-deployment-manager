package nats

import (
	"github.com/code-xd/k8s-deployment-manager/internal/repository/nats/common"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	"github.com/nats-io/nats.go"
)

// DeploymentRequestProducer produces deployment request messages to the producer channel
type DeploymentRequestProducer struct {
	producer *common.Producer
	channel  string
}

// NewDeploymentRequestProducer creates a new deployment request producer
func NewDeploymentRequestProducer(producer *common.Producer, cfg *dto.ProducerConfig) *DeploymentRequestProducer {
	return &DeploymentRequestProducer{
		producer: producer,
		channel:  cfg.DeploymentRequestChannel,
	}
}

// Publish sends a message with headers request_id and user_id and body DeploymentRequestMessage to the deployment request channel
func (p *DeploymentRequestProducer) Publish(requestID, userID string) error {
	header := nats.Header{}
	header.Set(dto.HeaderKeyRequestID, requestID)
	header.Set(dto.HeaderKeyUserID, userID)
	body := &dto.DeploymentRequestMessage{RequestID: requestID}
	return p.producer.PublishWithHeader(p.channel, body, header)
}
