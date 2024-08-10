// internal/adapters/redis_event_bus.go
package adapter

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/mateusmacedo/go-bff/pkg/application"
	"github.com/mateusmacedo/go-bff/pkg/domain"
)

type RedisEventBus[E domain.Event[D], D any] struct {
	publisher  *redisstream.Publisher
	subscriber *redisstream.Subscriber
	handlers   map[string][]application.EventHandler[E, D]
}

func NewRedisEventBus[E domain.Event[D], D any](publisher *redisstream.Publisher, subscriber *redisstream.Subscriber) *RedisEventBus[E, D] {
	return &RedisEventBus[E, D]{
		publisher:  publisher,
		subscriber: subscriber,
		handlers:   make(map[string][]application.EventHandler[E, D]),
	}
}

func (bus *RedisEventBus[E, D]) RegisterHandler(eventName string, handler application.EventHandler[E, D]) {
	bus.handlers[eventName] = append(bus.handlers[eventName], handler)

	go func() {
		messages, err := bus.subscriber.Subscribe(context.Background(), eventName)
		if err != nil {
			panic(err)
		}

		for msg := range messages {
			go func(msg *message.Message) {
				var payload D
				if err := json.Unmarshal(msg.Payload, &payload); err != nil {
					msg.Nack()
					return
				}

				event := &dynamicEvent[D]{
					eventName: eventName,
					payload:   payload,
				}

				if typedEvent, ok := interface{}(event).(E); ok {
					for _, handler := range bus.handlers[eventName] {
						if err := handler.Handle(context.Background(), typedEvent); err != nil {
							msg.Nack()
							return
						}
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

func (bus *RedisEventBus[E, D]) Publish(ctx context.Context, event E) error {
	payload, err := json.Marshal(event.Payload())
	if err != nil {
		return err
	}

	msg := message.NewMessage(event.EventName(), payload)
	return bus.publisher.Publish(event.EventName(), msg)
}

type dynamicEvent[D any] struct {
	eventName string
	payload   D
}

func (e *dynamicEvent[D]) EventName() string {
	return e.eventName
}

func (e *dynamicEvent[D]) Payload() D {
	return e.payload
}
