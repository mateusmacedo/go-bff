package main

import (
	"context"
	"time"

	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/google/uuid"

	app "github.com/mateusmacedo/go-bff/internal/busticket/application"
	"github.com/mateusmacedo/go-bff/internal/busticket/domain"
	"github.com/mateusmacedo/go-bff/internal/busticket/infrastructure"
	pkgDomain "github.com/mateusmacedo/go-bff/pkg/domain"
	"github.com/mateusmacedo/go-bff/pkg/infrastructure/channels/adapter"
	watermillLogAdapter "github.com/mateusmacedo/go-bff/pkg/infrastructure/watermill/adapter"
	zapAdapter "github.com/mateusmacedo/go-bff/pkg/infrastructure/zaplogger/adapter"
)

func main() {
	// Criação de um novo logger
	appLogger, err := zapAdapter.NewZapAppLogger()
	if err != nil {
		panic(err)
	}

	// Configuração do adaptador de logger
	logger := watermillLogAdapter.NewWatermillLoggerAdapter(appLogger)

	// Configuração do publisher e subscriber em memória
	pubSub := gochannel.NewGoChannel(gochannel.Config{}, logger)

	// Configuração do repositório
	repository := infrastructure.NewInMemoryPassageRepository(appLogger)

	// Gerador de ID
	idGenerator := func() string {
		return uuid.New().String()
	}

	// Criação dos handlers
	reserveHandler := app.NewReserveBusTicketHandler(repository, idGenerator, appLogger)
	findHandler := app.NewFindBusTicketHandler(repository, appLogger)

	// Criação dos barramentos usando Watermill
	commandBus := adapter.NewWatermillCommandBus[pkgDomain.Command[app.ReserveBusTicketData], app.ReserveBusTicketData](pubSub, pubSub, appLogger)
	queryBus := adapter.NewWatermillQueryBus[pkgDomain.Query[app.FindBusTicketData], app.FindBusTicketData, domain.BusTicket](pubSub, pubSub, appLogger)
	eventBus := adapter.NewWatermillEventBus[pkgDomain.Event[string], string](pubSub, appLogger)

	// Registro dos handlers nos barramentos
	commandBus.RegisterHandler("ReservePassage", reserveHandler)
	queryBus.RegisterHandler("FindPassage", findHandler)

	// Criando um contexto com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Criando um comando de reserva de passagem
	reserveData := app.ReserveBusTicketData{
		PassengerName: "John Doe",
		DepartureTime: time.Now().Add(24 * time.Hour).Truncate(time.Second),
		SeatNumber:    12,
		Origin:        "City A",
		Destination:   "City B",
	}
	command := app.NewReserveBusTicketCommand(reserveData)

	// Despachando o comando
	if err := commandBus.Dispatch(ctx, command); err != nil {
		appLogger.Error(ctx, "Erro ao despachar comando de reserva de passagem", map[string]interface{}{
			"command_name": command.CommandName(),
			"payload":      command.Payload(),
			"error":        err,
		})
		return
	}
	appLogger.Info(ctx, "Comando de reserva de passagem despachado com sucesso", nil)

	// Espera breve para permitir o processamento das mensagens
	time.Sleep(1 * time.Second)

	// Obtendo o ID da passagem diretamente do repositório para evitar inconsistências
	var passageID string
	for id, passage := range repository.GetData() {
		if passage.PassengerName == reserveData.PassengerName && passage.DepartureTime.Equal(reserveData.DepartureTime) {
			passageID = id
			break
		}
	}

	// Criando uma consulta para encontrar uma passagem
	query := app.NewFindBusTicketQuery(app.FindBusTicketData{
		PassageID: passageID,
	})

	// Despachando a consulta
	passage, err := queryBus.Dispatch(ctx, query)
	if err != nil {
		appLogger.Error(ctx, "Erro ao despachar consulta para encontrar passagem", map[string]interface{}{
			"query_name": query.QueryName(),
			"payload":    query.Payload(),
			"error":      err,
		})
	} else {
		appLogger.Info(ctx, "Consulta para encontrar passagem despachada com sucesso", map[string]interface{}{
			"passage": passage,
		})
	}

	// Exemplo de publicação de um evento
	event := app.NewBusTicketBookedEvent("Passage successfully booked for John Doe")
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
