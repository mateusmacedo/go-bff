package adapter

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/mateusmacedo/go-bff/pkg/application"
	"github.com/mateusmacedo/go-bff/pkg/domain"
	"github.com/mateusmacedo/go-bff/pkg/infrastructure"
)

type KafkaEventBus[E domain.Event[D], D any] struct {
	publisher  *kafka.Publisher
	subscriber *kafka.Subscriber
	handlers   map[string][]application.EventHandler[E, D]
	logger     application.AppLogger
}

func NewKafkaEventBus[E domain.Event[D], D any](publisher *kafka.Publisher, subscriber *kafka.Subscriber, logger application.AppLogger) *KafkaEventBus[E, D] {
	return &KafkaEventBus[E, D]{
		publisher:  publisher,
		subscriber: subscriber,
		handlers:   make(map[string][]application.EventHandler[E, D]),
		logger:     logger,
	}
}

func (bus *KafkaEventBus[E, D]) RegisterHandler(eventName string, handler application.EventHandler[E, D]) {
	bus.handlers[eventName] = append(bus.handlers[eventName], handler)

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		messages, err := bus.subscriber.Subscribe(ctx, eventName)
		if err != nil {
			infrastructure.LogError(ctx, bus.logger, "error subscribing to event", err, map[string]interface{}{
				"event_name": eventName,
			})
			return
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

func (bus *KafkaEventBus[E, D]) Publish(ctx context.Context, event E) error {
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
