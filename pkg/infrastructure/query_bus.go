package infrastructure

import (
	"context"
	"errors"
	"sync"

	"github.com/mateusmacedo/go-bff/pkg/application"
	"github.com/mateusmacedo/go-bff/pkg/domain"
)

// simpleQueryBus é uma implementação simples de um barramento de consultas.
type simpleQueryBus[Q domain.Query[T], T any, R any] struct {
	handlers map[string]application.QueryHandler[Q, T, R]
	mu       sync.RWMutex
}

// NewSimpleQueryBus cria uma nova instância de SimpleQueryBus.
func NewSimpleQueryBus[Q domain.Query[T], T any, R any]() *simpleQueryBus[Q, T, R] {
	return &simpleQueryBus[Q, T, R]{
		handlers: make(map[string]application.QueryHandler[Q, T, R]),
	}
}

// RegisterHandler registra um manipulador para uma consulta específica.
func (bus *simpleQueryBus[Q, T, R]) RegisterHandler(queryName string, handler application.QueryHandler[Q, T, R]) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.handlers[queryName] = handler
}

// Dispatch despacha uma consulta para o manipulador registrado usando goroutines.
func (bus *simpleQueryBus[Q, T, R]) Dispatch(ctx context.Context, query Q) (R, error) {
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
