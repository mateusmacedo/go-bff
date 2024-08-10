package infrastructure

import (
	"context"
	"fmt"
	"sync"

	"github.com/mateusmacedo/go-bff/pkg/application"
	"github.com/mateusmacedo/go-bff/pkg/domain"
)

// simpleEventBus é uma implementação simples de um barramento de eventos que utiliza goroutines.
type simpleEventBus[E domain.Event[T], T any] struct {
	handlers map[string][]application.EventHandler[E, T]
	mu       sync.RWMutex
	logger   application.AppLogger
}

// NewSimpleEventBus cria uma nova instância do SimpleEventBus.
func NewSimpleEventBus[E domain.Event[T], T any](logger application.AppLogger) application.EventBus[E, T] {
	return &simpleEventBus[E, T]{
		handlers: make(map[string][]application.EventHandler[E, T]),
		logger:   logger,
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
		bus.logger.Info(ctx, "no handler registered for event", map[string]interface{}{
			"event_name": event.EventName(),
		})
		return nil // Nenhum manipulador registrado, consideramos um sucesso silencioso
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(handlers))
	done := make(chan struct{})

	bus.logger.Info(ctx, "anonymous goroutine started", nil)
	go func() {
		bus.logger.Info(ctx, "waiting for goroutines to finish", nil)
		wg.Wait()
		bus.logger.Info(ctx, "all goroutines finished", nil)
		bus.logger.Info(ctx, "closing error channel", nil)
		close(errChan)
		bus.logger.Info(ctx, "closing done channel", nil)
		close(done)
	}()

	for _, handler := range handlers {
		wg.Add(1)
		go func(h application.EventHandler[E, T]) {
			defer wg.Done()
			if err := h.Handle(ctx, event); err != nil {
				select {
				case errChan <- err:
				case <-ctx.Done():
					bus.logger.Error(ctx, "error publishing event", map[string]interface{}{
						"event_name": event.EventName(),
						"error":      err,
					})
					return
				}
			}
			bus.logger.Info(ctx, "event published", map[string]interface{}{
				"event_name": event.EventName(),
			})
		}(handler)
	}

	select {
	case <-ctx.Done():
		bus.logger.Error(ctx, "error publishing event", map[string]interface{}{
			"event_name": event.EventName(),
			"error":      ctx.Err(),
		})
		return ctx.Err()
	case <-done:
		bus.logger.Info(ctx, "event published", map[string]interface{}{
			"event_name": event.EventName(),
		})
		return bus.collectErrors(ctx, errChan)
	}
}

// collectErrors coleta todos os erros de um canal e retorna um erro agregando todos eles.
func (bus *simpleEventBus[E, T]) collectErrors(ctx context.Context, errChan <-chan error) error {
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}
	if len(errors) > 0 {
		bus.logger.Error(ctx, "error publishing event", map[string]interface{}{
			"errors": errors,
		})
		return fmt.Errorf("errors: %v", errors)
	}
	return nil
}
