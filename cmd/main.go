package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	appLogger, err := zapAdapter.NewZapAppLogger()
	if err != nil {
		panic(err)
	}

	idGenerator := func() string {
		return uuid.New().String()
	}

	commandBus := pkgInfra.NewSimpleCommandBus[pkgDomain.Command[application.ReserveBusTicketData], application.ReserveBusTicketData](appLogger)
	queryBus := pkgInfra.NewSimpleQueryBus[pkgDomain.Query[application.FindBusTicketData], application.FindBusTicketData, []domain.BusTicket](appLogger)
	eventBus := pkgInfra.NewSimpleEventBus[pkgDomain.Event[string], string](appLogger)

	busTicketSlice := busticket.NewBusTicketSlice(commandBus, queryBus, idGenerator, appLogger, eventBus)

	router := chi.NewRouter()

	busTicketSlice.RegisterRoutes(router)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan
		appLogger.Info(ctx, "Sinal capturado", map[string]interface{}{"signal": sig})
		cancel()
	}()

	serverAddress := ":8080"
	server := &http.Server{
		Addr:    serverAddress,
		Handler: router,
	}

	go func() {
		appLogger.Info(ctx, "Server starting on:"+serverAddress, nil)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Error(ctx, "Erro ao iniciar o servidor", map[string]interface{}{
				"error": err,
			})
		}
	}()

	<-ctx.Done()
	appLogger.Info(ctx, "Encerrando servidor...", nil)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		appLogger.Error(context.Background(), "Erro ao encerrar servidor", map[string]interface{}{
			"error": err,
		})
	}

	appLogger.Info(context.Background(), "Servidor encerrado", nil)
}
