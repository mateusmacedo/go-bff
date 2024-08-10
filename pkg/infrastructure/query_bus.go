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
}

func NewSimpleQueryBus[Q domain.Query[D], D any, R any]() application.QueryBus[Q, D, R] {
	return &simpleQueryBus[Q, D, R]{
		handlers: make(map[string]application.QueryHandler[Q, D, R]),
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
		return zero, errors.New("no handler registered for query")
	}

	resultChan := make(chan R, 1)
	errChan := make(chan error, 1)

	go func() {
		result, err := handler.Handle(ctx, query)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- result
	}()

	select {
	case <-ctx.Done():
		return zero, ctx.Err()
	case result := <-resultChan:
		return result, nil
	case err := <-errChan:
		return zero, err
	}
}
