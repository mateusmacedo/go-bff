package infrastructure

import (
	"context"
	// Altere para o caminho correto do seu pacote

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/mateusmacedo/go-bff/pkg/application"
)

// zapAppLoggerAdapter é uma implementação da interface AppLogger usando zap
type zapAppLoggerAdapter struct {
	zapLogger *zap.Logger
}

// NewZapAppLogger cria uma nova instância de ZapLogger
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

// Info registra uma mensagem de nível INFO com campos do contexto
func (l *zapAppLoggerAdapter) Info(ctx context.Context, msg string, fields map[string]interface{}) {
	zapFields := convertFields(ctx, fields)
	l.zapLogger.With(zapFields...).Info(msg)
}

// Debug registra uma mensagem de nível DEBUG com campos do contexto
func (l *zapAppLoggerAdapter) Debug(ctx context.Context, msg string, fields map[string]interface{}) {
	zapFields := convertFields(ctx, fields)
	l.zapLogger.With(zapFields...).Debug(msg)
}

// Error registra uma mensagem de nível ERROR com campos do contexto
func (l *zapAppLoggerAdapter) Error(ctx context.Context, msg string, fields map[string]interface{}) {
	zapFields := convertFields(ctx, fields)
	l.zapLogger.With(zapFields...).Error(msg)
}

// Trace registra uma mensagem de nível TRACE com campos do contexto
func (l *zapAppLoggerAdapter) Trace(ctx context.Context, msg string, fields map[string]interface{}) {
	zapFields := convertFields(ctx, fields)
	l.zapLogger.With(zapFields...).Debug(msg) // Usando Debug para TRACE
}

// convertFields converte um mapa de campos para zap.Fields e inclui campos do contexto
func convertFields(ctx context.Context, fields map[string]interface{}) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))

	// Extraindo dados do contexto (exemplo: requestID)
	if requestID, ok := ctx.Value("requestID").(string); ok {
		zapFields = append(zapFields, zap.String("requestID", requestID))
	}

	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return zapFields
}
