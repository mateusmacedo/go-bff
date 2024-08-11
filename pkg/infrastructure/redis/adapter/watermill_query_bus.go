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
			application.LogError(ctx, bus.logger, "error subscribing to query", err, map[string]interface{}{
				"query_name": queryName,
			})
			panic(err)
		}

		for msg := range messages {
			go func(msg *message.Message) {
				var payload D
				if err := json.Unmarshal(msg.Payload, &payload); err != nil {
					application.LogError(ctx, bus.logger, "error unmarshalling query payload", err, map[string]interface{}{
						"query_name": queryName,
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
						application.LogError(ctx, bus.logger, "error handling query", err, map[string]interface{}{
							"query_name": queryName,
						})
						msg.Nack()
						return
					}

					responsePayload, err := json.Marshal(result)
					if err != nil {
						application.LogError(ctx, bus.logger, "error marshalling query response", err, map[string]interface{}{
							"query_name": queryName,
						})
						msg.Nack()
						return
					}

					responseMsg := message.NewMessage(queryName+"_response", responsePayload)
					if err := bus.publisher.Publish(queryName+"_response", responseMsg); err != nil {
						application.LogError(ctx, bus.logger, "error publishing query response", err, map[string]interface{}{
							"query_name": queryName,
						})
						msg.Nack()
						return
					}
				} else {
					application.LogError(ctx, bus.logger, "error casting query", err, map[string]interface{}{
						"query_name": queryName,
					})
					msg.Nack()
					return
				}

				application.LogInfo(ctx, bus.logger, "query handled", map[string]interface{}{
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
		application.LogError(ctx, bus.logger, "error marshalling query payload", err, map[string]interface{}{
			"query_name": query.QueryName(),
		})
		var zero R
		return zero, err
	}

	msg := message.NewMessage(query.QueryName(), payload)
	if err := bus.publisher.Publish(query.QueryName(), msg); err != nil {
		application.LogError(ctx, bus.logger, "error publishing query", err, map[string]interface{}{
			"query_name": query.QueryName(),
		})
		var zero R
		return zero, err
	}

	responseMessages, err := bus.subscriber.Subscribe(ctx, query.QueryName()+"_response")
	if err != nil {
		application.LogError(ctx, bus.logger, "error subscribing to query response", err, map[string]interface{}{
			"query_name": query.QueryName(),
		})
		var zero R
		return zero, err
	}

	select {
	case responseMsg := <-responseMessages:
		var result R
		if err := json.Unmarshal(responseMsg.Payload, &result); err != nil {
			application.LogError(ctx, bus.logger, "error unmarshalling query response", err, map[string]interface{}{
				"query_name": query.QueryName(),
			})
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
