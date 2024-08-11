package adapter

import (
	"context"
	"sync"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/mateusmacedo/go-bff/pkg/application"
	"github.com/mateusmacedo/go-bff/pkg/domain"
)

type WatermillEventBus[E domain.Event[D], D any] struct {
	publisher message.Publisher
	handlers  map[string][]application.EventHandler[E, D]
	mu        sync.RWMutex
	logger    application.AppLogger
}

func NewWatermillEventBus[E domain.Event[D], D any](publisher message.Publisher, logger application.AppLogger) *WatermillEventBus[E, D] {
	return &WatermillEventBus[E, D]{
		publisher: publisher,
		handlers:  make(map[string][]application.EventHandler[E, D]),
		logger:    logger,
	}
}

func (bus *WatermillEventBus[E, D]) RegisterHandler(eventName string, handler application.EventHandler[E, D]) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.handlers[eventName] = append(bus.handlers[eventName], handler)
}

func (bus *WatermillEventBus[E, D]) Publish(ctx context.Context, event E) error {
	eventName := event.EventName()

	bus.mu.RLock()
	handlers, found := bus.handlers[eventName]
	bus.mu.RUnlock()

	if !found {
		bus.logger.Info(ctx, "no handlers registered", map[string]interface{}{
			"event_name": eventName,
		})
		return nil
	}

	payload, err := application.MarshalPayload(event.Payload())
	if err != nil {
		application.LogError(ctx, bus.logger, "error marshalling event payload", err, map[string]interface{}{
			"event_name": eventName,
		})
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	if err := bus.publisher.Publish(eventName, msg); err != nil {
		application.LogError(ctx, bus.logger, "error publishing event", err, map[string]interface{}{
			"event_name": eventName,
		})
		return err
	}

	for _, handler := range handlers {
		if err := handler.Handle(ctx, event); err != nil {
			application.LogError(ctx, bus.logger, "error handling event", err, map[string]interface{}{
				"event_name": eventName,
			})
			return err
		}
	}

	application.LogInfo(ctx, bus.logger, "event published", map[string]interface{}{
		"event_name": eventName,
	})
	return nil
}
