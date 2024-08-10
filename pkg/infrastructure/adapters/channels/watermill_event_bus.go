package channels

import (
	"context"
	"encoding/json"
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
}

func NewWatermillEventBus[E domain.Event[D], D any](publisher message.Publisher) *WatermillEventBus[E, D] {
	return &WatermillEventBus[E, D]{
		publisher: publisher,
		handlers:  make(map[string][]application.EventHandler[E, D]),
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
		return nil // Nenhum manipulador registrado, consideramos um sucesso silencioso
	}

	payload, err := json.Marshal(event.Payload())
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	if err := bus.publisher.Publish(eventName, msg); err != nil {
		return err
	}

	for _, handler := range handlers {
		if err := handler.Handle(ctx, event); err != nil {
			return err
		}
	}

	return nil
}
