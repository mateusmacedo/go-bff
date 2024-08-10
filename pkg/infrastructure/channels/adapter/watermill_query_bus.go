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

type WatermillQueryBus[Q domain.Query[D], D any, R any] struct {
	publisher  message.Publisher
	subscriber message.Subscriber
	handlers   map[string]application.QueryHandler[Q, D, R]
	mu         sync.RWMutex
}

func NewWatermillQueryBus[Q domain.Query[D], D any, R any](publisher message.Publisher, subscriber message.Subscriber) *WatermillQueryBus[Q, D, R] {
	return &WatermillQueryBus[Q, D, R]{
		publisher:  publisher,
		subscriber: subscriber,
		handlers:   make(map[string]application.QueryHandler[Q, D, R]),
	}
}

func (bus *WatermillQueryBus[Q, D, R]) RegisterHandler(queryName string, handler application.QueryHandler[Q, D, R]) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.handlers[queryName] = handler

	go func() {
		messages, err := bus.subscriber.Subscribe(context.Background(), queryName)
		if err != nil {
			panic(err) // Handle error according to your needs
		}

		for msg := range messages {
			go func(msg *message.Message) {
				var payload D
				if err := json.Unmarshal(msg.Payload, &payload); err != nil {
					// Log the error
					return
				}

				// Implement Query interface dynamically
				query := &dynamicQuery[D]{
					queryName: queryName,
					payload:   payload,
				}

				// Assert the query as type Q
				if typedQuery, ok := interface{}(query).(Q); ok {
					result, err := handler.Handle(context.Background(), typedQuery)
					if err != nil {
						// Log the error
						return
					}

					responsePayload, err := json.Marshal(result)
					if err != nil {
						// Log the error
						return
					}

					responseMsg := message.NewMessage(watermill.NewUUID(), responsePayload)
					if err := bus.publisher.Publish(queryName+"_response", responseMsg); err != nil {
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

func (bus *WatermillQueryBus[Q, D, R]) Dispatch(ctx context.Context, query Q) (R, error) {
	payload, err := json.Marshal(query.Payload())
	if err != nil {
		var zero R
		return zero, err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
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
