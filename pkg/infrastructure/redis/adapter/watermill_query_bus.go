// internal/adapters/redis_query_bus.go
package adapter

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/mateusmacedo/go-bff/pkg/application"
	"github.com/mateusmacedo/go-bff/pkg/domain"
)

type RedisQueryBus[Q domain.Query[D], D any, R any] struct {
	publisher  *redisstream.Publisher
	subscriber *redisstream.Subscriber
	handlers   map[string]application.QueryHandler[Q, D, R]
	logger     application.AppLogger
}

func NewRedisQueryBus[Q domain.Query[D], D any, R any](publisher *redisstream.Publisher, subscriber *redisstream.Subscriber, logger application.AppLogger) *RedisQueryBus[Q, D, R] {
	return &RedisQueryBus[Q, D, R]{
		publisher:  publisher,
		subscriber: subscriber,
		handlers:   make(map[string]application.QueryHandler[Q, D, R]),
		logger:     logger,
	}
}

func (bus *RedisQueryBus[Q, D, R]) RegisterHandler(queryName string, handler application.QueryHandler[Q, D, R]) {
	bus.handlers[queryName] = handler

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		messages, err := bus.subscriber.Subscribe(ctx, queryName)
		if err != nil {
			bus.logger.Error(ctx, "error subscribing to query", map[string]interface{}{
				"query_name": queryName,
				"error":      err,
			})
			panic(err)
		}

		for msg := range messages {
			go func(msg *message.Message) {
				var payload D
				if err := json.Unmarshal(msg.Payload, &payload); err != nil {
					bus.logger.Error(ctx, "error unmarshalling query payload", map[string]interface{}{
						"query_name": queryName,
						"error":      err,
					})
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
						bus.logger.Error(ctx, "error handling query", map[string]interface{}{
							"query_name": queryName,
							"error":      err,
						})
						msg.Nack()
						return
					}

					responsePayload, err := json.Marshal(result)
					if err != nil {
						bus.logger.Error(ctx, "error marshalling query result", map[string]interface{}{
							"query_name": queryName,
							"error":      err,
						})
						msg.Nack()
						return
					}

					responseMsg := message.NewMessage(queryName+"_response", responsePayload)
					if err := bus.publisher.Publish(queryName+"_response", responseMsg); err != nil {
						bus.logger.Error(ctx, "error publishing query response", map[string]interface{}{
							"query_name": queryName,
							"error":      err,
						})
						msg.Nack()
						return
					}
				} else {
					bus.logger.Error(ctx, "error asserting query type", map[string]interface{}{
						"query_name": queryName,
					})
					msg.Nack()
					return
				}

				bus.logger.Info(ctx, "query handled", map[string]interface{}{
					"query_name": queryName,
				})
				msg.Ack()
			}(msg)
		}
	}()
}

func (bus *RedisQueryBus[Q, D, R]) Dispatch(ctx context.Context, query Q) (R, error) {
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
