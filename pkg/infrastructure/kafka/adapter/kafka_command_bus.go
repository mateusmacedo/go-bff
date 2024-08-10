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

type KafkaCommandBus[C domain.Command[T], T any] struct {
	publisher  *kafka.Publisher
	subscriber *kafka.Subscriber
	handlers   map[string]application.CommandHandler[C, T]
	logger     application.AppLogger
}

func NewKafkaCommandBus[C domain.Command[T], T any](publisher *kafka.Publisher, subscriber *kafka.Subscriber, logger application.AppLogger) *KafkaCommandBus[C, T] {
	return &KafkaCommandBus[C, T]{
		publisher:  publisher,
		subscriber: subscriber,
		handlers:   make(map[string]application.CommandHandler[C, T]),
		logger:     logger,
	}
}

func (bus *KafkaCommandBus[C, T]) RegisterHandler(commandName string, handler application.CommandHandler[C, T]) {
	bus.handlers[commandName] = handler

	go bus.subscribeAndHandle(commandName)
}

func (bus *KafkaCommandBus[C, T]) subscribeAndHandle(commandName string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messages, err := bus.subscriber.Subscribe(ctx, commandName)
	if err != nil {
		bus.logger.Error(ctx, "error subscribing to command", map[string]interface{}{
			"command_name": commandName,
			"error":        err,
		})
		return
	}

	for msg := range messages {
		go bus.handleMessage(ctx, commandName, msg)
	}
}

func (bus *KafkaCommandBus[C, T]) handleMessage(ctx context.Context, commandName string, msg *message.Message) {
	var payload T
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		bus.logger.Error(ctx, "error unmarshalling command payload", map[string]interface{}{
			"command_name": commandName,
			"error":        err,
		})
		msg.Nack()
		return
	}

	command := &dynamicCommand[T]{commandName: commandName, payload: payload}
	if typedCommand, ok := interface{}(command).(C); ok {
		if err := bus.handlers[commandName].Handle(ctx, typedCommand); err != nil {
			bus.logger.Error(ctx, "error handling command", map[string]interface{}{
				"command_name": commandName,
				"error":        err,
			})
			msg.Nack()
			return
		}
	} else {
		bus.logger.Error(ctx, "error asserting command type", map[string]interface{}{
			"command_name": commandName,
		})
		msg.Nack()
		return
	}

	bus.logger.Info(ctx, "command handled", map[string]interface{}{
		"command_name": commandName,
	})
	msg.Ack()
}

func (bus *KafkaCommandBus[C, T]) Dispatch(ctx context.Context, command C) error {
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
