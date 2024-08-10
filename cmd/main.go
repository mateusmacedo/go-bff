package main

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/mateusmacedo/go-bff/internal/busticket/application"
	"github.com/mateusmacedo/go-bff/internal/busticket/domain"
	"github.com/mateusmacedo/go-bff/internal/busticket/infrastructure"
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

	// Configuração do repositório
	repository := infrastructure.NewInMemoryPassageRepository(appLogger)

	// Gerador de ID
	idGenerator := func() string {
		return uuid.New().String()
	}

	// Criação dos handlers
	reserveHandler := application.NewReserveBusTicketHandler(repository, idGenerator, appLogger)
	findHandler := application.NewFindBusTicketHandler(repository, appLogger)

	// Criação dos barramentos
	commandBus := pkgInfra.NewSimpleCommandBus[pkgDomain.Command[application.ReserveBusTicketData], application.ReserveBusTicketData](appLogger)
	queryBus := pkgInfra.NewSimpleQueryBus[pkgDomain.Query[application.FindBusTicketData], application.FindBusTicketData, domain.BusTicket](appLogger)
	eventBus := pkgInfra.NewSimpleEventBus[pkgDomain.Event[string], string](appLogger)

	// Registro dos handlers nos barramentos
	commandBus.RegisterHandler("ReservePassage", reserveHandler)
	queryBus.RegisterHandler("FindPassage", findHandler)

	// Criando um contexto com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Criando um comando de reserva de passagem
	reserveData := application.ReserveBusTicketData{
		PassengerName: "John Doe",
		DepartureTime: time.Now().Add(24 * time.Hour),
		SeatNumber:    12,
		Origin:        "City A",
		Destination:   "City B",
	}
	command := application.NewReserveBusTicketCommand(reserveData)

	// Despachando o comando
	if err := commandBus.Dispatch(ctx, command); err != nil {
		appLogger.Error(ctx, "Erro ao reservar passagem", map[string]interface{}{
			"command_name": command.CommandName(),
			"payload":      command.Payload(),
			"error":        err,
		})
		return
	}
	appLogger.Info(ctx, "Passagem reservada com sucesso", map[string]interface{}{
		"passengerName": reserveData.PassengerName,
		"departureTime": reserveData.DepartureTime,
	})

	// Obtendo o ID da passagem diretamente do repositório para evitar inconsistências
	var passageID string
	for id, passage := range repository.GetData() { // Utilize GetData() se disponível
		if passage.PassengerName == reserveData.PassengerName && passage.DepartureTime.Equal(reserveData.DepartureTime) {
			passageID = id
			break
		}
	}

	// Criando uma consulta para encontrar uma passagem
	query := application.NewFindBusTicketQuery(application.FindBusTicketData{
		PassageID: passageID, // Use o ID gerado corretamente
	})

	// Despachando a consulta
	passage, err := queryBus.Dispatch(ctx, query)
	if err != nil {
		appLogger.Error(ctx, "Erro ao encontrar passagem", map[string]interface{}{
			"query_name": query.QueryName(),
			"payload":    query.Payload(),
			"error":      err,
		})
	} else {
		appLogger.Info(ctx, "Passagem encontrada", map[string]interface{}{
			"passengerName": passage.PassengerName,
			"departureTime": passage.DepartureTime,
		})
	}

	// Exemplo de publicação de um evento
	event := application.NewBusTicketBookedEvent("Passage successfully booked for John Doe")
	if err := eventBus.Publish(ctx, event); err != nil {
		appLogger.Error(ctx, "Erro ao publicar evento", map[string]interface{}{
			"event_name": event.EventName(),
			"payload":    event.Payload(),
			"error":      err,
		})
	} else {
		appLogger.Info(ctx, "Evento publicado com sucesso", map[string]interface{}{
			"event_name": event.EventName(),
			"payload":    event.Payload(),
		})
	}
}
