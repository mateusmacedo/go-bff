package adapter

import (
	"context"
	"encoding/json"
	"errors"
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
	logger     application.AppLogger
}

func NewWatermillQueryBus[Q domain.Query[D], D any, R any](publisher message.Publisher, subscriber message.Subscriber, logger application.AppLogger) *WatermillQueryBus[Q, D, R] {
	return &WatermillQueryBus[Q, D, R]{
		publisher:  publisher,
		subscriber: subscriber,
		handlers:   make(map[string]application.QueryHandler[Q, D, R]),
		logger:     logger,
	}
}

func (bus *WatermillQueryBus[Q, D, R]) RegisterHandler(queryName string, handler application.QueryHandler[Q, D, R]) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.handlers[queryName] = handler

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		messages, err := bus.subscriber.Subscribe(ctx, queryName)
		if err != nil {
			application.LogError(ctx, bus.logger, "error subscribing to query", err, map[string]interface{}{
				"query_name": queryName,
			})
			return
		}

		for msg := range messages {
			go bus.processMessage(ctx, queryName, handler, msg)
		}
	}()
}

func (bus *WatermillQueryBus[Q, D, R]) Dispatch(ctx context.Context, query Q) (R, error) {
	payload, err := application.MarshalPayload(query.Payload())
	if err != nil {
		application.LogError(ctx, bus.logger, "error marshalling query payload", err, map[string]interface{}{
			"query_name": query.QueryName(),
		})
		var zero R
		return zero, err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
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
		application.LogError(ctx, bus.logger, "context done", ctx.Err(), map[string]interface{}{
			"query_name": query.QueryName(),
		})
		var zero R
		return zero, ctx.Err()
	}
}

func (bus *WatermillQueryBus[Q, D, R]) processMessage(ctx context.Context, queryName string, handler application.QueryHandler[Q, D, R], msg *message.Message) {
	var payload D
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		application.LogError(ctx, bus.logger, "error unmarshalling query payload", err, map[string]interface{}{
			"query_name": queryName,
		})
		return
	}

	query := &dynamicQuery[D]{
		queryName: queryName,
		payload:   payload,
	}

	if typedQuery, ok := interface{}(query).(Q); ok {
		result, err := handler.Handle(ctx, typedQuery)
		if err != nil {
			application.LogError(ctx, bus.logger, "error handling query", err, map[string]interface{}{
				"query_name": queryName,
			})
			return
		}

		responsePayload, err := application.MarshalPayload(result)
		if err != nil {
			application.LogError(ctx, bus.logger, "error marshalling query response", err, map[string]interface{}{
				"query_name": queryName,
			})
			return
		}

		responseMsg := message.NewMessage(watermill.NewUUID(), responsePayload)
		if err := bus.publisher.Publish(queryName+"_response", responseMsg); err != nil {
			application.LogError(ctx, bus.logger, "error publishing query response", err, map[string]interface{}{
				"query_name": queryName,
			})
			return
		}
	} else {
		err := errors.New("error handling query")
		application.LogError(ctx, bus.logger, "error handling query", err, map[string]interface{}{
			"query_name": queryName,
		})
		return
	}

	application.LogInfo(ctx, bus.logger, "query handled", map[string]interface{}{
		"query_name": queryName,
	})
	msg.Ack()
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
