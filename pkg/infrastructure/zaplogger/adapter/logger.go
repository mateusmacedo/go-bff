package adapter

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/mateusmacedo/go-bff/pkg/application"
)

type zapAppLoggerAdapter struct {
	zapLogger *zap.Logger
}

func NewZapAppLogger() (application.AppLogger, error) {
	config := zap.NewProductionConfig()
	config.InitialFields = map[string]interface{}{"app": "bff-watermill"}
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zapLogger, err := config.Build()
	zapLogger = zapLogger.WithOptions(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return &zapAppLoggerAdapter{zapLogger: zapLogger}, nil
}

func (l *zapAppLoggerAdapter) Info(ctx context.Context, msg string, fields map[string]interface{}) {
	zapFields := convertFields(ctx, fields)
	l.zapLogger.With(zapFields...).Info(msg)
}

func (l *zapAppLoggerAdapter) Debug(ctx context.Context, msg string, fields map[string]interface{}) {
	zapFields := convertFields(ctx, fields)
	l.zapLogger.With(zapFields...).Debug(msg)
}

func (l *zapAppLoggerAdapter) Error(ctx context.Context, msg string, fields map[string]interface{}) {
	zapFields := convertFields(ctx, fields)
	l.zapLogger.With(zapFields...).Error(msg)
}

func (l *zapAppLoggerAdapter) Trace(ctx context.Context, msg string, fields map[string]interface{}) {
	zapFields := convertFields(ctx, fields)
	l.zapLogger.With(zapFields...).Debug(msg)
}

func convertFields(ctx context.Context, fields map[string]interface{}) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))

	if requestID, ok := ctx.Value("requestID").(string); ok {
		zapFields = append(zapFields, zap.String("requestID", requestID))
	}

	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return zapFields
}
