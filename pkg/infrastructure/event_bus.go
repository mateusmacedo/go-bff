package infrastructure

import (
	"context"
	"fmt"
	"sync"

	"github.com/mateusmacedo/go-bff/pkg/application"
	"github.com/mateusmacedo/go-bff/pkg/domain"
)

type simpleEventBus[E domain.Event[T], T any] struct {
	handlers map[string][]application.EventHandler[E, T]
	mu       sync.RWMutex
	logger   application.AppLogger
}

func NewSimpleEventBus[E domain.Event[T], T any](logger application.AppLogger) application.EventBus[E, T] {
	return &simpleEventBus[E, T]{
		handlers: make(map[string][]application.EventHandler[E, T]),
		logger:   logger,
	}
}

func (bus *simpleEventBus[E, T]) RegisterHandler(eventName string, handler application.EventHandler[E, T]) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.handlers[eventName] = append(bus.handlers[eventName], handler)
}

func (bus *simpleEventBus[E, T]) Publish(ctx context.Context, event E) error {
	bus.mu.RLock()
	handlers, found := bus.handlers[event.EventName()]
	bus.mu.RUnlock()

	if !found {
		LogInfo(ctx, bus.logger, "no handler registered for event", map[string]interface{}{
			"event_name": event.EventName(),
		})
		return nil
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(handlers))
	done := make(chan struct{})

	LogInfo(ctx, bus.logger, "publishing event", map[string]interface{}{
		"event_name": event.EventName(),
	})
	go func() {
		bus.logger.Info(ctx, "waiting for goroutines to finish", nil)
		wg.Wait()
		LogInfo(ctx, bus.logger, "all goroutines finished", nil)
		close(errChan)
		LogInfo(ctx, bus.logger, "error channel closed", nil)
		close(done)
		LogInfo(ctx, bus.logger, "done channel closed", nil)
	}()

	for _, handler := range handlers {
		wg.Add(1)
		go func(h application.EventHandler[E, T]) {
			defer wg.Done()
			if err := h.Handle(ctx, event); err != nil {
				select {
				case errChan <- err:
				case <-ctx.Done():
					LogError(ctx, bus.logger, "context done", ctx.Err(), map[string]interface{}{
						"event_name": event.EventName(),
						"error":      err,
					})
					return
				}
			}
			LogInfo(ctx, bus.logger, "event handled", map[string]interface{}{
				"event_name": event.EventName(),
			})
		}(handler)
	}

	select {
	case <-ctx.Done():
		LogError(ctx, bus.logger, "context done", ctx.Err(), nil)
		return ctx.Err()
	case <-done:
		LogInfo(ctx, bus.logger, "event published", map[string]interface{}{
			"event_name": event.EventName(),
		})
		return bus.collectErrors(ctx, errChan)
	}
}

func (bus *simpleEventBus[E, T]) collectErrors(ctx context.Context, errChan <-chan error) error {
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}
	if len(errors) > 0 {
		LogError(ctx, bus.logger, "errors publishing event", nil, map[string]interface{}{
			"errors": errors,
		})
		return fmt.Errorf("errors: %v", errors)
	}
	return nil
}
