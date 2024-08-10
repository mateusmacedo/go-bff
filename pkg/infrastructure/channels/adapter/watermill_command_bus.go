package adapter

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
	logger     application.AppLogger
}

func NewWatermillCommandBus[C domain.Command[T], T any](publisher message.Publisher, subscriber message.Subscriber, logger application.AppLogger) *WatermillCommandBus[C, T] {
	return &WatermillCommandBus[C, T]{
		publisher:  publisher,
		subscriber: subscriber,
		handlers:   make(map[string]application.CommandHandler[C, T]),
		logger:     logger,
	}
}

func (bus *WatermillCommandBus[C, T]) RegisterHandler(commandName string, handler application.CommandHandler[C, T]) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.handlers[commandName] = handler

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		messages, err := bus.subscriber.Subscribe(ctx, commandName)
		if err != nil {
			bus.logger.Error(ctx, "error subscribing to command", map[string]interface{}{
				"command_name": commandName,
				"error":        err,
			})
			panic(err) // Handle error according to your needs
		}

		for msg := range messages {
			go func(msg *message.Message) {
				var payload T
				if err := json.Unmarshal(msg.Payload, &payload); err != nil {
					bus.logger.Error(ctx, "error unmarshalling command payload", map[string]interface{}{
						"command_name": commandName,
						"error":        err,
					})
					return
				}

				// Implement Command interface dynamically
				command := &dynamicCommand[T]{
					commandName: commandName,
					payload:     payload,
				}

				// Assert the command as type C
				if typedCommand, ok := interface{}(command).(C); ok {
					if err := handler.Handle(ctx, typedCommand); err != nil {
						bus.logger.Error(ctx, "error handling command", map[string]interface{}{
							"command_name": commandName,
							"error":        err,
						})
						return
					}
				} else {
					bus.logger.Error(ctx, "error asserting command type", map[string]interface{}{
						"command_name": commandName,
					})
					return
				}

				bus.logger.Info(ctx, "command handled", map[string]interface{}{
					"command_name": commandName,
				})
				msg.Ack()
			}(msg)
		}
	}()
}

func (bus *WatermillCommandBus[C, T]) Dispatch(ctx context.Context, command C) error {
	payload, err := json.Marshal(command.Payload())
	if err != nil {
		bus.logger.Error(ctx, "error marshalling command payload", map[string]interface{}{
			"command_name": command.CommandName(),
			"error":        err,
		})
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	if err := bus.publisher.Publish(command.CommandName(), msg); err != nil {
		bus.logger.Error(ctx, "error publishing command", map[string]interface{}{
			"command_name": command.CommandName(),
			"error":        err,
		})
		return err
	}

	bus.logger.Info(ctx, "command dispatched", map[string]interface{}{
		"command_name": command.CommandName(),
	})
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
