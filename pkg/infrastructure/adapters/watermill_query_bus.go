package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/mateusmacedo/go-bff/pkg/application"
	"github.com/mateusmacedo/go-bff/pkg/domain"
)

type WatermillQueryBus[Q domain.Query[D], D any, R any] struct {
	publisher message.Publisher
	handlers  map[string]application.QueryHandler[Q, D, R]
	mu        sync.RWMutex
}

func NewWatermillQueryBus[Q domain.Query[D], D any, R any](publisher message.Publisher) *WatermillQueryBus[Q, D, R] {
	return &WatermillQueryBus[Q, D, R]{
		publisher: publisher,
		handlers:  make(map[string]application.QueryHandler[Q, D, R]),
	}
}

func (bus *WatermillQueryBus[Q, D, R]) RegisterHandler(queryName string, handler application.QueryHandler[Q, D, R]) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.handlers[queryName] = handler
}

func (bus *WatermillQueryBus[Q, D, R]) Dispatch(ctx context.Context, query Q) (R, error) {
	queryName := query.QueryName()

	bus.mu.RLock()
	handler, found := bus.handlers[queryName]
	bus.mu.RUnlock()

	var zero R
	if !found {
		return zero, errors.New("no handler registered for query")
	}

	payload, err := json.Marshal(query.Payload())
	if err != nil {
		return zero, err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	if err := bus.publisher.Publish(queryName, msg); err != nil {
		return zero, err
	}

	return handler.Handle(ctx, query)
}
