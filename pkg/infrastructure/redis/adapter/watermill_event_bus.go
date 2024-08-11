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

type RedisEventBus[E domain.Event[D], D any] struct {
	publisher  *redisstream.Publisher
	subscriber *redisstream.Subscriber
	handlers   map[string][]application.EventHandler[E, D]
	logger     application.AppLogger
}

func NewRedisEventBus[E domain.Event[D], D any](publisher *redisstream.Publisher, subscriber *redisstream.Subscriber, logger application.AppLogger) *RedisEventBus[E, D] {
	return &RedisEventBus[E, D]{
		publisher:  publisher,
		subscriber: subscriber,
		handlers:   make(map[string][]application.EventHandler[E, D]),
		logger:     logger,
	}
}

func (bus *RedisEventBus[E, D]) RegisterHandler(eventName string, handler application.EventHandler[E, D]) {
	bus.handlers[eventName] = append(bus.handlers[eventName], handler)

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		messages, err := bus.subscriber.Subscribe(ctx, eventName)
		if err != nil {
			infrastructure.LogError(ctx, bus.logger, "error subscribing to event", err, map[string]interface{}{
				"event_name": eventName,
			})
		}

		for msg := range messages {
			go func(msg *message.Message) {
				var payload D
				if err := json.Unmarshal(msg.Payload, &payload); err != nil {
					infrastructure.LogError(ctx, bus.logger, "error unmarshalling event payload", err, map[string]interface{}{
						"event_name": eventName,
					})
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
							infrastructure.LogError(ctx, bus.logger, "error handling event", err, map[string]interface{}{
								"event_name": eventName,
							})
							msg.Nack()
							return
						}
					}
				} else {
					infrastructure.LogError(ctx, bus.logger, "error casting event", err, map[string]interface{}{
						"event_name": eventName,
					})
					msg.Nack()
					return
				}

				infrastructure.LogInfo(ctx, bus.logger, "event handled", map[string]interface{}{
					"event_name": eventName,
				})
				msg.Ack()
			}(msg)
		}
	}()
}

func (bus *RedisEventBus[E, D]) Publish(ctx context.Context, event E) error {
	payload, err := json.Marshal(event.Payload())
	if err != nil {
		infrastructure.LogError(ctx, bus.logger, "error marshalling event payload", err, map[string]interface{}{
			"event_name": event.EventName(),
		})
		return err
	}

	infrastructure.LogInfo(ctx, bus.logger, "publishing event", map[string]interface{}{
		"event_name": event.EventName(),
	})
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
