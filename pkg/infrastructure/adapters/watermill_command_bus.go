// internal/adapters/watermill_command_bus.go
package adapters

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/mateusmacedo/go-bff/pkg/application"
	"github.com/mateusmacedo/go-bff/pkg/domain"
)

type WatermillCommandBus[C domain.Command[T], T any] struct {
	publisher  message.Publisher
	subscriber message.Subscriber
	handlers   map[string]application.CommandHandler[C, T]
	mu         sync.RWMutex
}

func NewWatermillCommandBus[C domain.Command[T], T any](publisher message.Publisher, subscriber message.Subscriber) *WatermillCommandBus[C, T] {
	return &WatermillCommandBus[C, T]{
		publisher:  publisher,
		subscriber: subscriber,
		handlers:   make(map[string]application.CommandHandler[C, T]),
	}
}

func (bus *WatermillCommandBus[C, T]) RegisterHandler(commandName string, handler application.CommandHandler[C, T]) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.handlers[commandName] = handler

	go func() {
		messages, err := bus.subscriber.Subscribe(context.Background(), commandName)
		if err != nil {
			panic(err) // Handle error according to your needs
		}

		for msg := range messages {
			go func(msg *message.Message) {
				var payload T
				if err := json.Unmarshal(msg.Payload, &payload); err != nil {
					// Log the error
					return
				}

				// Implement Command interface dynamically
				command := &dynamicCommand[T]{
					commandName: commandName,
					payload:     payload,
				}

				// Assert the command as type C
				if typedCommand, ok := interface{}(command).(C); ok {
					if err := handler.Handle(context.Background(), typedCommand); err != nil {
						// Log the error
						return
					}
				} else {
					// Handle error if the type assertion fails
					return
				}

				msg.Ack()
			}(msg)
		}
	}()
}

func (bus *WatermillCommandBus[C, T]) Dispatch(ctx context.Context, command C) error {
	payload, err := json.Marshal(command.Payload())
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	if err := bus.publisher.Publish(command.CommandName(), msg); err != nil {
		return err
	}

	return nil
}

type dynamicCommand[T any] struct {
	commandName string
	payload     T
}

func (c *dynamicCommand[T]) CommandName() string {
	return c.commandName
}

func (c *dynamicCommand[T]) Payload() T {
	return c.payload
}
