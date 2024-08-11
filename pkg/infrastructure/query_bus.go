package infrastructure

import (
	"context"
	"errors"
	"sync"

	"github.com/mateusmacedo/go-bff/pkg/application"
	"github.com/mateusmacedo/go-bff/pkg/domain"
)

type simpleQueryBus[Q domain.Query[D], D any, R any] struct {
	handlers map[string]application.QueryHandler[Q, D, R]
	mu       sync.RWMutex
	logger   application.AppLogger
}

func NewSimpleQueryBus[Q domain.Query[D], D any, R any](logger application.AppLogger) application.QueryBus[Q, D, R] {
	return &simpleQueryBus[Q, D, R]{
		handlers: make(map[string]application.QueryHandler[Q, D, R]),
		logger:   logger,
	}
}

func (bus *simpleQueryBus[Q, D, R]) RegisterHandler(queryName string, handler application.QueryHandler[Q, D, R]) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.handlers[queryName] = handler
}

func (bus *simpleQueryBus[Q, D, R]) Dispatch(ctx context.Context, query Q) (R, error) {
	bus.mu.RLock()
	handler, found := bus.handlers[query.QueryName()]
	bus.mu.RUnlock()

	var zero R
	if !found {
		err := errors.New("no handler registered for query")
		application.LogError(ctx, bus.logger, "no handler registered for query", err, map[string]interface{}{
			"query_name": query.QueryName(),
		})
		return zero, err
	}

	resultChan := make(chan R, 1)
	errChan := make(chan error, 1)

	go func() {
		result, err := handler.Handle(ctx, query)
		if err != nil {
			application.LogError(ctx, bus.logger, "error handling query", err, map[string]interface{}{
				"query_name": query.QueryName(),
			})
			errChan <- err
			return
		}

		application.LogInfo(ctx, bus.logger, "query handled", map[string]interface{}{
			"query_name": query.QueryName(),
		})
		resultChan <- result
	}()

	select {
	case <-ctx.Done():
		application.LogError(ctx, bus.logger, "context done", ctx.Err(), nil)
		return zero, ctx.Err()
	case result := <-resultChan:
		application.LogInfo(ctx, bus.logger, "query handled", map[string]interface{}{
			"query_name": query.QueryName(),
		})
		return result, nil
	case err := <-errChan:
		application.LogError(ctx, bus.logger, "error handling query", err, map[string]interface{}{
			"query_name": query.QueryName(),
		})
		return zero, err
	}
}
