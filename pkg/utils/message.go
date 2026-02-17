package utils

import (
	"encoding/json"
	"fmt"
)

// UnmarshalMessage unmarshals message data (e.g. from NATS) into the provided type.
// Useful in handlers: v, err := utils.UnmarshalMessage[MyType](msg.Data)
func UnmarshalMessage[T any](data []byte) (*T, error) {
	var out T
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("unmarshal message: %w", err)
	}
	return &out, nil
}
