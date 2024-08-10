package infrastructure

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/mateusmacedo/go-bff/pkg/application"
)

func GenerateUUID() string {
	return uuid.New().String()
}

func LogError(ctx context.Context, logger application.AppLogger, message string, err error, fields map[string]interface{}) {
	logData := make(map[string]interface{})
	for k, v := range fields {
		logData[k] = v
	}
	if err != nil {
		logData["error"] = err
	}
	logger.Error(ctx, message, logData)
}

func MarshalPayload[T any](payload T) ([]byte, error) {
	return json.Marshal(payload)
}
