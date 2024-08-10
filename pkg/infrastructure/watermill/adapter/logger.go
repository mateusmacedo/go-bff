package adapter

import (
	"context"

	"github.com/ThreeDotsLabs/watermill"

	"github.com/mateusmacedo/go-bff/pkg/application"
)

type watermillLoggerAdapter struct {
	appLogger application.AppLogger
	fields    watermill.LogFields
}

func NewWatermillLoggerAdapter(appLogger application.AppLogger) watermill.LoggerAdapter {
	return &watermillLoggerAdapter{
		appLogger: appLogger,
		fields:    watermill.LogFields{},
	}
}

func (a *watermillLoggerAdapter) Error(msg string, err error, fields watermill.LogFields) {
	allFields := a.combineFields(fields)
	allFields["error"] = err.Error()
	a.appLogger.Error(context.TODO(), msg, allFields)
}

func (a *watermillLoggerAdapter) Info(msg string, fields watermill.LogFields) {
	allFields := a.combineFields(fields)
	a.appLogger.Info(context.TODO(), msg, allFields)
}

func (a *watermillLoggerAdapter) Debug(msg string, fields watermill.LogFields) {
	allFields := a.combineFields(fields)
	a.appLogger.Debug(context.TODO(), msg, allFields)
}

func (a *watermillLoggerAdapter) Trace(msg string, fields watermill.LogFields) {
	allFields := a.combineFields(fields)
	a.appLogger.Trace(context.TODO(), msg, allFields)
}

func (a *watermillLoggerAdapter) With(fields watermill.LogFields) watermill.LoggerAdapter {
	newFields := a.combineFields(fields)
	return &watermillLoggerAdapter{
		appLogger: a.appLogger,
		fields:    newFields,
	}
}

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
