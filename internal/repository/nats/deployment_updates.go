package nats

import (
	"github.com/code-xd/k8s-deployment-manager/internal/repository/nats/common"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	"github.com/nats-io/nats.go"
)

// DeploymentUpdateProducer produces deployment update messages to the producer channel
type DeploymentUpdateProducer struct {
	producer *common.Producer
	channel  string
}

// NewDeploymentUpdateProducer creates a new deployment update producer
func NewDeploymentUpdateProducer(producer *common.Producer, cfg *dto.ProducerConfig) *DeploymentUpdateProducer {
	return &DeploymentUpdateProducer{
		producer: producer,
		channel:  cfg.DeploymentUpdateChannel,
	}
}

// Publish sends a message with headers request_id and user_id and body DeploymentUpdateMessage to the deployment update channel
func (p *DeploymentUpdateProducer) Publish(identifier, requestID, userID string) error {
	header := nats.Header{}
	header.Set(dto.HeaderKeyRequestID, requestID)
	header.Set(dto.HeaderKeyUserID, userID)
	body := &dto.DeploymentUpdateMessage{Identifier: identifier}
	return p.producer.PublishWithHeader(p.channel, body, header)
}
