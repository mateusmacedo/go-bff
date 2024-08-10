package infrastructure

import (
	"context"
	"sync"

	"github.com/mateusmacedo/go-bff/pkg/application"
	"github.com/mateusmacedo/go-bff/pkg/domain"
)

// simpleEventBus é uma implementação simples de um barramento de eventos que utiliza goroutines.
type simpleEventBus[E domain.Event[T], T any] struct {
	handlers map[string][]application.EventHandler[E, T]
	mu       sync.RWMutex
}

// NewSimpleEventBus cria uma nova instância do SimpleEventBus.
func NewSimpleEventBus[E domain.Event[T], T any]() *simpleEventBus[E, T] {
	return &simpleEventBus[E, T]{
		handlers: make(map[string][]application.EventHandler[E, T]),
	}
}

// RegisterHandler registra um manipulador para um evento específico.
func (bus *simpleEventBus[E, T]) RegisterHandler(eventName string, handler application.EventHandler[E, T]) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.handlers[eventName] = append(bus.handlers[eventName], handler)
}

// Publish publica um evento para os manipuladores registrados usando goroutines.
func (bus *simpleEventBus[E, T]) Publish(ctx context.Context, event E) error {
	bus.mu.RLock()
	handlers, found := bus.handlers[event.EventName()]
	bus.mu.RUnlock()

	if !found {
		return nil // Nenhum manipulador registrado, consideramos um sucesso silencioso
	}

	done := make(chan error, len(handlers))
	var wg sync.WaitGroup

	for _, handler := range handlers {
		wg.Add(1)
		go func(h application.EventHandler[E, T]) {
			defer wg.Done()
			if err := h.Handle(ctx, event); err != nil {
				done <- err
			}
		}(handler)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err, ok := <-done:
			if !ok {
				return nil // Todos os handlers foram processados
			}
			if err != nil {
				return err // Retorna o primeiro erro encontrado
			}
		}
	}
}
