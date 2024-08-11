package adapter

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/mateusmacedo/go-bff/pkg/application"
	"github.com/mateusmacedo/go-bff/pkg/domain"
	"github.com/mateusmacedo/go-bff/pkg/infrastructure"
)

type RedisCommandBus[C domain.Command[T], T any] struct {
	publisher  *redisstream.Publisher
	subscriber *redisstream.Subscriber
	handlers   map[string]application.CommandHandler[C, T]
	logger     application.AppLogger
}

func NewRedisCommandBus[C domain.Command[T], T any](publisher *redisstream.Publisher, subscriber *redisstream.Subscriber, logger application.AppLogger) *RedisCommandBus[C, T] {
	return &RedisCommandBus[C, T]{
		publisher:  publisher,
		subscriber: subscriber,
		handlers:   make(map[string]application.CommandHandler[C, T]),
		logger:     logger,
	}
}

func (bus *RedisCommandBus[C, T]) RegisterHandler(commandName string, handler application.CommandHandler[C, T]) {
	bus.handlers[commandName] = handler

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		messages, err := bus.subscriber.Subscribe(ctx, commandName)
		if err != nil {
			infrastructure.LogError(ctx, bus.logger, "error subscribing to command", err, map[string]interface{}{
				"command_name": commandName,
			})
		}

		for msg := range messages {
			go func(msg *message.Message) {
				var payload T
				if err := json.Unmarshal(msg.Payload, &payload); err != nil {
					infrastructure.LogError(ctx, bus.logger, "error unmarshalling command payload", err, map[string]interface{}{
						"command_name": commandName,
					})
					msg.Nack()
					return
				}

				command := &dynamicCommand[T]{
					commandName: commandName,
					payload:     payload,
				}

				if typedCommand, ok := interface{}(command).(C); ok {
					if err := handler.Handle(ctx, typedCommand); err != nil {
						infrastructure.LogError(ctx, bus.logger, "error handling command", err, map[string]interface{}{
							"command_name": commandName,
						})
						msg.Nack()
						return
					}
				} else {
					infrastructure.LogError(ctx, bus.logger, "error casting command", err, map[string]interface{}{
						"command_name": commandName,
					})
					msg.Nack()
					return
				}

				infrastructure.LogInfo(ctx, bus.logger, "command handled", map[string]interface{}{
					"command_name": commandName,
				})
				msg.Ack()
			}(msg)
		}
	}()
}

func (bus *RedisCommandBus[C, T]) Dispatch(ctx context.Context, command C) error {
	payload, err := json.Marshal(command.Payload())
	if err != nil {
		infrastructure.LogError(ctx, bus.logger, "error marshalling command payload", err, map[string]interface{}{
			"command_name": command.CommandName(),
		})
		return err
	}

	msg := message.NewMessage(command.CommandName(), payload)
	if err := bus.publisher.Publish(command.CommandName(), msg); err != nil {
		infrastructure.LogError(ctx, bus.logger, "error publishing command", err, map[string]interface{}{
			"command_name": command.CommandName(),
		})
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
