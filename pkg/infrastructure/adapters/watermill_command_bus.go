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

type WatermillCommandBus[C domain.Command[T], T any] struct {
	publisher message.Publisher
	handlers  map[string]application.CommandHandler[C, T]
	mu        sync.RWMutex
}

func NewWatermillCommandBus[C domain.Command[T], T any](publisher message.Publisher) *WatermillCommandBus[C, T] {
	return &WatermillCommandBus[C, T]{
		publisher: publisher,
		handlers:  make(map[string]application.CommandHandler[C, T]),
	}
}

func (bus *WatermillCommandBus[C, T]) RegisterHandler(commandName string, handler application.CommandHandler[C, T]) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.handlers[commandName] = handler
}

func (bus *WatermillCommandBus[C, T]) Dispatch(ctx context.Context, command C) error {
	commandName := command.CommandName()

	bus.mu.RLock()
	handler, found := bus.handlers[commandName]
	bus.mu.RUnlock()

	if !found {
		return errors.New("no handler registered for command")
	}

	payload, err := json.Marshal(command.Payload())
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	if err := bus.publisher.Publish(commandName, msg); err != nil {
		return err
	}

	return handler.Handle(ctx, command)
}
