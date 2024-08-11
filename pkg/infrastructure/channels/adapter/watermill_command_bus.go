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
			application.LogError(ctx, bus.logger, "error subscribing to command", err, map[string]interface{}{
				"command_name": commandName,
			})
			panic(err)
		}

		for msg := range messages {
			go bus.processMessage(ctx, commandName, handler, msg)
		}
	}()
}

func (bus *WatermillCommandBus[C, T]) Dispatch(ctx context.Context, command C) error {
	payload, err := application.MarshalPayload(command.Payload())
	if err != nil {
		application.LogError(ctx, bus.logger, "error marshalling command payload", err, map[string]interface{}{
			"command_name": command.CommandName(),
		})
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	if err := bus.publisher.Publish(command.CommandName(), msg); err != nil {
		application.LogError(ctx, bus.logger, "error publishing command", err, map[string]interface{}{
			"command_name": command.CommandName(),
		})
		return err
	}

	application.LogInfo(ctx, bus.logger, "command dispatched", map[string]interface{}{
		"command_name": command.CommandName(),
	})
	return nil
}

func (bus *WatermillCommandBus[C, T]) processMessage(ctx context.Context, commandName string, handler application.CommandHandler[C, T], msg *message.Message) {
	var payload T
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		application.LogError(ctx, bus.logger, "error unmarshalling command payload", err, map[string]interface{}{
			"command_name": commandName,
		})
		return
	}

	command := &dynamicCommand[T]{
		commandName: commandName,
		payload:     payload,
	}

	if typedCommand, ok := interface{}(command).(C); ok {
		if err := handler.Handle(ctx, typedCommand); err != nil {
			application.LogError(ctx, bus.logger, "error handling command", err, map[string]interface{}{
				"command_name": commandName,
			})
			return
		}
	} else {
		application.LogError(ctx, bus.logger, "error asserting command type", nil, map[string]interface{}{
			"command_name": commandName,
		})
		return
	}

	application.LogInfo(ctx, bus.logger, "command handled", map[string]interface{}{
		"command_name": commandName,
	})
	msg.Ack()
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
