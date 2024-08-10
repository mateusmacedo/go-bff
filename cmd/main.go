package main

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/mateusmacedo/go-bff/internal/busticket"
	"github.com/mateusmacedo/go-bff/internal/busticket/application"
	"github.com/mateusmacedo/go-bff/internal/busticket/domain"
	pkgDomain "github.com/mateusmacedo/go-bff/pkg/domain"
	pkgInfra "github.com/mateusmacedo/go-bff/pkg/infrastructure"
	zapAdapter "github.com/mateusmacedo/go-bff/pkg/infrastructure/zaplogger/adapter"
)

func main() {
	// Criação de um novo logger
	appLogger, err := zapAdapter.NewZapAppLogger()
	if err != nil {
		panic(err)
	}

	// Gerador de ID
	idGenerator := func() string {
		return uuid.New().String()
	}

	// Criação dos barramentos
	commandBus := pkgInfra.NewSimpleCommandBus[pkgDomain.Command[application.ReserveBusTicketData], application.ReserveBusTicketData](appLogger)
	queryBus := pkgInfra.NewSimpleQueryBus[pkgDomain.Query[application.FindBusTicketData], application.FindBusTicketData, domain.BusTicket](appLogger)
	eventBus := pkgInfra.NewSimpleEventBus[pkgDomain.Event[string], string](appLogger)

	// Criar um contexto de aplicação de ticket de ônibus (BusTicket) utilizando o slice
	busTicketSlice := busticket.NewBusTicketSlice(commandBus, queryBus, idGenerator, appLogger, eventBus)

	// Configuração do roteador HTTP
	router := chi.NewRouter()

	// Registro das rotas HTTP
	busTicketSlice.RegisterRoutes(router)

	// Iniciando o servidor HTTP
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
