package application

import (
	"context"
	"encoding/json"
)

type AppLogger interface {
	Info(ctx context.Context, msg string, fields map[string]interface{})
	Debug(ctx context.Context, msg string, fields map[string]interface{})
	Error(ctx context.Context, msg string, fields map[string]interface{})
	Trace(ctx context.Context, msg string, fields map[string]interface{})
}

func LogError(ctx context.Context, logger AppLogger, message string, err error, fields map[string]interface{}) {
	logData := make(map[string]interface{})
	for k, v := range fields {
		logData[k] = v
	}
	if err != nil {
		logData["error"] = err
	}
	logger.Error(ctx, message, logData)
}

func LogInfo(ctx context.Context, logger AppLogger, message string, fields map[string]interface{}) {
	logData := make(map[string]interface{})
	for k, v := range fields {
		logData[k] = v
	}
	logger.Info(ctx, message, logData)
}

func LogDebug(ctx context.Context, logger AppLogger, message string, fields map[string]interface{}) {
	logData := make(map[string]interface{})
	for k, v := range fields {
		logData[k] = v
	}
	logger.Debug(ctx, message, logData)
}

func LogTrace(ctx context.Context, logger AppLogger, message string, fields map[string]interface{}) {
	logData := make(map[string]interface{})
	for k, v := range fields {
		logData[k] = v
	}
	logger.Trace(ctx, message, logData)
}

func MarshalPayload[T any](payload T) ([]byte, error) {
	return json.Marshal(payload)
}
