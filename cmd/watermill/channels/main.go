package main

import (
	"context"
	"net/http"

	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/mateusmacedo/go-bff/internal/busticket"
	"github.com/mateusmacedo/go-bff/internal/busticket/application"
	"github.com/mateusmacedo/go-bff/internal/busticket/domain"
	pkgDomain "github.com/mateusmacedo/go-bff/pkg/domain"
	"github.com/mateusmacedo/go-bff/pkg/infrastructure/channels/adapter"
	watermillLogAdapter "github.com/mateusmacedo/go-bff/pkg/infrastructure/watermill/adapter"
	zapAdapter "github.com/mateusmacedo/go-bff/pkg/infrastructure/zaplogger/adapter"
)

func main() {

	appLogger, err := zapAdapter.NewZapAppLogger()
	if err != nil {
		panic(err)
	}

	logger := watermillLogAdapter.NewWatermillLoggerAdapter(appLogger)

	pubSub := gochannel.NewGoChannel(gochannel.Config{}, logger)

	commandBus := adapter.NewWatermillCommandBus[pkgDomain.Command[application.ReserveBusTicketData], application.ReserveBusTicketData](pubSub, pubSub, appLogger)
	queryBus := adapter.NewWatermillQueryBus[pkgDomain.Query[application.FindBusTicketData], application.FindBusTicketData, []domain.BusTicket](pubSub, pubSub, appLogger)
	eventBus := adapter.NewWatermillEventBus[pkgDomain.Event[string], string](pubSub, appLogger)

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
