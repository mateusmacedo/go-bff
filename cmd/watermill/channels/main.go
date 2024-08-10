package main

import (
	"context"
	"time"

	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/google/uuid"

	app "github.com/mateusmacedo/go-bff/internal/application"
	"github.com/mateusmacedo/go-bff/internal/domain"
	"github.com/mateusmacedo/go-bff/internal/infrastructure"
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
	repository := infrastructure.NewInMemoryPassageRepository()

	// Gerador de ID
	idGenerator := func() string {
		return uuid.New().String()
	}

	// Criação dos handlers
	reserveHandler := app.NewReservePassageHandler(repository, idGenerator, appLogger)
	findHandler := app.NewFindPassageHandler(repository, appLogger)

	// Criação dos barramentos usando Watermill
	commandBus := adapter.NewWatermillCommandBus[pkgDomain.Command[app.ReservePassageData], app.ReservePassageData](pubSub, pubSub, appLogger)
	queryBus := adapter.NewWatermillQueryBus[pkgDomain.Query[app.FindPassageData], app.FindPassageData, domain.Passage](pubSub, pubSub, appLogger)
	eventBus := adapter.NewWatermillEventBus[pkgDomain.Event[string], string](pubSub, appLogger)

	// Registro dos handlers nos barramentos
	commandBus.RegisterHandler("ReservePassage", reserveHandler)
	queryBus.RegisterHandler("FindPassage", findHandler)

	// Criando um contexto com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Criando um comando de reserva de passagem
	reserveData := app.ReservePassageData{
		PassengerName: "John Doe",
		DepartureTime: time.Now().Add(24 * time.Hour).Truncate(time.Second),
		SeatNumber:    12,
		Origin:        "City A",
		Destination:   "City B",
	}
	command := app.NewReservePassageCommand(reserveData)

	// Despachando o comando
	if err := commandBus.Dispatch(ctx, command); err != nil {
		appLogger.Error(ctx, "Erro ao despachar comando de reserva de passagem", map[string]interface{}{
			"error": err,
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
	query := app.NewFindPassageQuery(app.FindPassageData{
		PassageID: passageID,
	})

	// Despachando a consulta
	passage, err := queryBus.Dispatch(ctx, query)
	if err != nil {
		appLogger.Error(ctx, "Erro ao despachar consulta para encontrar passagem", map[string]interface{}{
			"error": err,
		})
	} else {
		appLogger.Info(ctx, "Consulta para encontrar passagem despachada com sucesso", map[string]interface{}{
			"passage": passage,
		})
	}

	// Exemplo de publicação de um evento
	event := app.NewPassageBookedEvent("Passage successfully booked for John Doe")
	if err := eventBus.Publish(ctx, event); err != nil {
		appLogger.Error(ctx, "Erro ao publicar evento", map[string]interface{}{
			"error": err,
		})
	} else {
		appLogger.Info(ctx, "Evento publicado com sucesso", nil)
	}
}
