package main

import (
	"context"
	"net/http"

	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/mateusmacedo/go-bff/internal/busticket"
	"github.com/mateusmacedo/go-bff/internal/busticket/application"
	"github.com/mateusmacedo/go-bff/internal/busticket/domain"
	pkgDomain "github.com/mateusmacedo/go-bff/pkg/domain"
	"github.com/mateusmacedo/go-bff/pkg/infrastructure/redis/adapter"
	watermillLogAdapter "github.com/mateusmacedo/go-bff/pkg/infrastructure/watermill/adapter"
	zapAdapter "github.com/mateusmacedo/go-bff/pkg/infrastructure/zaplogger/adapter"
)

func main() {

	appLogger, err := zapAdapter.NewZapAppLogger()
	if err != nil {
		panic(err)
	}

	logger := watermillLogAdapter.NewWatermillLoggerAdapter(appLogger)

	redisClient := adapter.NewRedisClient()
	defer redisClient.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	publisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: redisClient,
	}, logger)
	if err != nil {
		appLogger.Error(ctx, "Erro ao criar publisher", map[string]interface{}{
			"error": err,
		})
		return
	}
	defer publisher.Close()

	subscriber, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        redisClient,
		ConsumerGroup: "my_group",
		Consumer:      "my_consumer",
	}, logger)
	if err != nil {
		appLogger.Error(ctx, "Erro ao criar subscriber", map[string]interface{}{
			"error": err,
		})
		return
	}
	defer subscriber.Close()

	commandBus := adapter.NewRedisCommandBus[pkgDomain.Command[application.ReserveBusTicketData], application.ReserveBusTicketData](publisher, subscriber, appLogger)
	queryBus := adapter.NewRedisQueryBus[pkgDomain.Query[application.FindBusTicketData], application.FindBusTicketData, []domain.BusTicket](publisher, subscriber, appLogger)
	eventBus := adapter.NewRedisEventBus[pkgDomain.Event[string], string](publisher, subscriber, appLogger)

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
