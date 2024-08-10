package adapter

import (
	"context"

	"github.com/ThreeDotsLabs/watermill"

	"github.com/mateusmacedo/go-bff/pkg/application"
)

// watermillLoggerAdapter é uma implementação da interface LoggerAdapter usando AppLogger
type watermillLoggerAdapter struct {
	appLogger application.AppLogger
	fields    watermill.LogFields
}

// NewWatermillLoggerAdapter cria uma nova instância de ZapLoggerAdapter
func NewWatermillLoggerAdapter(appLogger application.AppLogger) watermill.LoggerAdapter {
	return &watermillLoggerAdapter{
		appLogger: appLogger,
		fields:    watermill.LogFields{},
	}
}

// Error registra uma mensagem de nível ERROR com campos adicionais e erro
func (a *watermillLoggerAdapter) Error(msg string, err error, fields watermill.LogFields) {
	allFields := a.combineFields(fields)
	allFields["error"] = err.Error()
	a.appLogger.Error(context.TODO(), msg, allFields)
}

// Info registra uma mensagem de nível INFO com campos adicionais
func (a *watermillLoggerAdapter) Info(msg string, fields watermill.LogFields) {
	allFields := a.combineFields(fields)
	a.appLogger.Info(context.TODO(), msg, allFields)
}

// Debug registra uma mensagem de nível DEBUG com campos adicionais
func (a *watermillLoggerAdapter) Debug(msg string, fields watermill.LogFields) {
	allFields := a.combineFields(fields)
	a.appLogger.Debug(context.TODO(), msg, allFields)
}

// Trace registra uma mensagem de nível TRACE com campos adicionais
func (a *watermillLoggerAdapter) Trace(msg string, fields watermill.LogFields) {
	allFields := a.combineFields(fields)
	a.appLogger.Trace(context.TODO(), msg, allFields)
}

// With retorna uma nova instância de LoggerAdapter com campos adicionais
func (a *watermillLoggerAdapter) With(fields watermill.LogFields) watermill.LoggerAdapter {
	newFields := a.combineFields(fields)
	return &watermillLoggerAdapter{
		appLogger: a.appLogger,
		fields:    newFields,
	}
}

// combineFields combina os campos atuais com novos campos
func (a *watermillLoggerAdapter) combineFields(fields watermill.LogFields) watermill.LogFields {
	allFields := make(watermill.LogFields, len(a.fields)+len(fields))

	for k, v := range a.fields {
		allFields[k] = v
	}

	for k, v := range fields {
		allFields[k] = v
	}
	return allFields
}
