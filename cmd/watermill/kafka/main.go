package main

import (
	"context"
	"net/http"
	"time"

	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/mateusmacedo/go-bff/internal/busticket"
	"github.com/mateusmacedo/go-bff/internal/busticket/application"
	"github.com/mateusmacedo/go-bff/internal/busticket/domain"
	pkgDomain "github.com/mateusmacedo/go-bff/pkg/domain"
	"github.com/mateusmacedo/go-bff/pkg/infrastructure/kafka/adapter"
	watermillLogAdapter "github.com/mateusmacedo/go-bff/pkg/infrastructure/watermill/adapter"
	zapAdapter "github.com/mateusmacedo/go-bff/pkg/infrastructure/zaplogger/adapter"
)

func main() {

	appLogger, err := zapAdapter.NewZapAppLogger()
	if err != nil {
		panic(err)
	}

	logger := watermillLogAdapter.NewWatermillLoggerAdapter(appLogger)

	marshaler := kafka.DefaultMarshaler{}

	publisherConfig := kafka.PublisherConfig{
		Brokers:   []string{"localhost:9092"},
		Marshaler: marshaler,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	publisher, err := kafka.NewPublisher(publisherConfig, logger)
	if err != nil {
		appLogger.Error(ctx, "Erro ao criar publisher", map[string]interface{}{
			"error": err,
		})
	}
	defer publisher.Close()

	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V1_0_0_0
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.ClientID = "watermill"

	subscriberConfig := kafka.SubscriberConfig{
		Brokers:               []string{"localhost:9092"},
		Unmarshaler:           marshaler,
		ConsumerGroup:         "example_consumer_group",
		OverwriteSaramaConfig: saramaConfig,
		InitializeTopicDetails: &sarama.TopicDetail{
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	subscriber, err := kafka.NewSubscriber(subscriberConfig, logger)
	if err != nil {
		appLogger.Error(ctx, "Erro ao criar subscriber", map[string]interface{}{
			"error": err,
		})
	}
	defer subscriber.Close()

	commandBus := adapter.NewKafkaCommandBus[pkgDomain.Command[application.ReserveBusTicketData], application.ReserveBusTicketData](publisher, subscriber, appLogger)
	queryBus := adapter.NewKafkaQueryBus[pkgDomain.Query[application.FindBusTicketData], application.FindBusTicketData, []domain.BusTicket](publisher, subscriber, appLogger)
	eventBus := adapter.NewKafkaEventBus[pkgDomain.Event[string], string](publisher, subscriber, appLogger)

	idGenerator := func() string {
		return uuid.New().String()
	}

	busTicketSlice := busticket.NewBusTicketSlice(commandBus, queryBus, idGenerator, appLogger, eventBus)

	router := chi.NewRouter()

	busTicketSlice.RegisterRoutes(router)

	serverAddress := ":8080"
	appLogger.Info(context.Background(), "Starting HTTP server", map[string]interface{}{
		"address": serverAddress,
	})
	if err := http.ListenAndServe(serverAddress, router); err != nil {
		appLogger.Error(context.Background(), "Failed to start HTTP server", map[string]interface{}{
			"error": err,
		})
	}
}
