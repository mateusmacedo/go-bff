package adapter

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/mateusmacedo/go-bff/pkg/application"
	"github.com/mateusmacedo/go-bff/pkg/domain"
)

type RedisCommandBus[C domain.Command[T], T any] struct {
	publisher  *redisstream.Publisher
	subscriber *redisstream.Subscriber
	handlers   map[string]application.CommandHandler[C, T]
}

func NewRedisCommandBus[C domain.Command[T], T any](publisher *redisstream.Publisher, subscriber *redisstream.Subscriber) *RedisCommandBus[C, T] {
	return &RedisCommandBus[C, T]{
		publisher:  publisher,
		subscriber: subscriber,
		handlers:   make(map[string]application.CommandHandler[C, T]),
	}
}

func (bus *RedisCommandBus[C, T]) RegisterHandler(commandName string, handler application.CommandHandler[C, T]) {
	bus.handlers[commandName] = handler

	go func() {
		messages, err := bus.subscriber.Subscribe(context.Background(), commandName)
		if err != nil {
			panic(err)
		}

		for msg := range messages {
			go func(msg *message.Message) {
				var payload T
				if err := json.Unmarshal(msg.Payload, &payload); err != nil {
					msg.Nack()
					return
				}

				command := &dynamicCommand[T]{
					commandName: commandName,
					payload:     payload,
				}

				if typedCommand, ok := interface{}(command).(C); ok {
					if err := handler.Handle(context.Background(), typedCommand); err != nil {
						msg.Nack()
						return
					}
				} else {
					msg.Nack()
					return
				}

				msg.Ack()
			}(msg)
		}
	}()
}

func (bus *RedisCommandBus[C, T]) Dispatch(ctx context.Context, command C) error {
	payload, err := json.Marshal(command.Payload())
	if err != nil {
		return err
	}

	msg := message.NewMessage(command.CommandName(), payload)
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
