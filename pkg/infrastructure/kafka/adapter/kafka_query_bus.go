package adapter

import (
	"context"
	"encoding/json"
	"log"

	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/mateusmacedo/go-bff/pkg/application"
	"github.com/mateusmacedo/go-bff/pkg/domain"
)

type KafkaQueryBus[Q domain.Query[D], D any, R any] struct {
	publisher  *kafka.Publisher
	subscriber *kafka.Subscriber
	handlers   map[string]application.QueryHandler[Q, D, R]
}

func NewKafkaQueryBus[Q domain.Query[D], D any, R any](publisher *kafka.Publisher, subscriber *kafka.Subscriber) *KafkaQueryBus[Q, D, R] {
	return &KafkaQueryBus[Q, D, R]{
		publisher:  publisher,
		subscriber: subscriber,
		handlers:   make(map[string]application.QueryHandler[Q, D, R]),
	}
}

func (bus *KafkaQueryBus[Q, D, R]) RegisterHandler(queryName string, handler application.QueryHandler[Q, D, R]) {
	bus.handlers[queryName] = handler

	go func() {
		messages, err := bus.subscriber.Subscribe(context.Background(), queryName)
		if err != nil {
			log.Fatalf("could not subscribe to query: %v", err)
		}

		for msg := range messages {
			go func(msg *message.Message) {
				var payload D
				if err := json.Unmarshal(msg.Payload, &payload); err != nil {
					msg.Nack()
					return
				}

				query := &dynamicQuery[D]{
					queryName: queryName,
					payload:   payload,
				}

				if typedQuery, ok := interface{}(query).(Q); ok {
					result, err := handler.Handle(context.Background(), typedQuery)
					if err != nil {
						msg.Nack()
						return
					}

					responsePayload, err := json.Marshal(result)
					if err != nil {
						msg.Nack()
						return
					}

					responseMsg := message.NewMessage(queryName+"_response", responsePayload)
					if err := bus.publisher.Publish(queryName+"_response", responseMsg); err != nil {
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

func (bus *KafkaQueryBus[Q, D, R]) Dispatch(ctx context.Context, query Q) (R, error) {
	payload, err := json.Marshal(query.Payload())
	if err != nil {
		var zero R
		return zero, err
	}

	msg := message.NewMessage(query.QueryName(), payload)
	if err := bus.publisher.Publish(query.QueryName(), msg); err != nil {
		var zero R
		return zero, err
	}

	responseMessages, err := bus.subscriber.Subscribe(ctx, query.QueryName()+"_response")
	if err != nil {
		var zero R
		return zero, err
	}

	select {
	case responseMsg := <-responseMessages:
		var result R
		if err := json.Unmarshal(responseMsg.Payload, &result); err != nil {
			var zero R
			return zero, err
		}
		responseMsg.Ack()
		return result, nil
	case <-ctx.Done():
		var zero R
		return zero, ctx.Err()
	}
}

type dynamicQuery[D any] struct {
	queryName string
	payload   D
}

func (q *dynamicQuery[D]) QueryName() string {
	return q.queryName
}

func (q *dynamicQuery[D]) Payload() D {
	return q.payload
}
