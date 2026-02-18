package utils

import (
	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	"github.com/google/uuid"
)

// GetUserIDFromWorkerHeader extracts and parses user_id from NATS message headers.
// Returns uuid.Nil if the header is missing or invalid.
func GetUserIDFromWorkerHeader(headers map[string][]string) uuid.UUID {
	if headers == nil {
		return uuid.Nil
	}
	vals, ok := headers[dto.HeaderKeyUserID]
	if !ok || len(vals) == 0 {
		return uuid.Nil
	}
	id, err := uuid.Parse(vals[0])
	if err != nil {
		return uuid.Nil
	}
	return id
}
