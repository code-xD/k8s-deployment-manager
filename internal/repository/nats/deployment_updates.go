package nats

import (
	"github.com/code-xd/k8s-deployment-manager/internal/repository/nats/common"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
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

// Publish sends a DeploymentUpdateMessage (identifier, eventType) to the deployment update channel
func (p *DeploymentUpdateProducer) Publish(identifier, eventType string) error {
	body := &dto.DeploymentUpdateMessage{Identifier: identifier, EventType: eventType}
	return p.producer.Publish(p.channel, body)
}
