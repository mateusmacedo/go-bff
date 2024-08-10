package main

import (
	"context"
	"time"

	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/google/uuid"

	"github.com/mateusmacedo/go-bff/internal/application"
	"github.com/mateusmacedo/go-bff/internal/domain"
	"github.com/mateusmacedo/go-bff/internal/infrastructure"
	pkgDomain "github.com/mateusmacedo/go-bff/pkg/domain"
	"github.com/mateusmacedo/go-bff/pkg/infrastructure/redis/adapter"
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

	// Configuração do Redis
	redisClient := adapter.NewRedisClient()
	defer redisClient.Close()

	// Configurar o contexto com cancelamento
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	publisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: redisClient,
	}, logger)
	if err != nil {
		appLogger.Error(ctx, "Erro ao criar publisher", map[string]interface{}{
			"error": err,
		})
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
	}
	defer subscriber.Close()

	// Configuração do repositório
	repository := infrastructure.NewInMemoryPassageRepository(appLogger)

	// Gerador de ID
	idGenerator := func() string {
		return uuid.New().String()
	}

	// Criação dos handlers
	reserveHandler := application.NewReservePassageHandler(repository, idGenerator, appLogger)
	findHandler := application.NewFindPassageHandler(repository, appLogger)

	// Criação dos barramentos usando Redis
	commandBus := adapter.NewRedisCommandBus[pkgDomain.Command[application.ReservePassageData], application.ReservePassageData](publisher, subscriber, appLogger)
	queryBus := adapter.NewRedisQueryBus[pkgDomain.Query[application.FindPassageData], application.FindPassageData, domain.Passage](publisher, subscriber, appLogger)
	eventBus := adapter.NewRedisEventBus[pkgDomain.Event[string], string](publisher, subscriber, appLogger)

	// Registro dos handlers nos barramentos
	commandBus.RegisterHandler("ReservePassage", reserveHandler)
	queryBus.RegisterHandler("FindPassage", findHandler)

	// Criando um comando de reserva de passagem
	reserveData := application.ReservePassageData{
		PassengerName: "John Doe",
		DepartureTime: time.Now().Add(24 * time.Hour).Truncate(time.Second),
		SeatNumber:    12,
		Origin:        "City A",
		Destination:   "City B",
	}

	// Despachando o comando
	command := application.NewReservePassageCommand(reserveData)
	appLogger.Info(ctx, "Despachando o comando para reservar passagem...", map[string]interface{}{
		"command_name": command.CommandName(),
	})
	if err := commandBus.Dispatch(ctx, command); err != nil {
		appLogger.Error(ctx, "Erro ao reservar passagem", map[string]interface{}{
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
	query := application.NewFindPassageQuery(application.FindPassageData{
		PassageID: passageID,
	})

	// Despachando a consulta
	passage, err := queryBus.Dispatch(ctx, query)
	if err != nil {
		appLogger.Error(ctx, "Erro ao despachar consulta para encontrar passagem", map[string]interface{}{
			"query_name": query,
			"payload":    query.Payload(),
			"error":      err,
		})
		return
	}
	appLogger.Info(ctx, "Consulta para encontrar passagem despachada com sucesso", map[string]interface{}{
		"passage": passage,
	})

	// Exemplo de publicação de um evento
	event := application.NewPassageBookedEvent("Passage successfully booked for John Doe")
	if err := eventBus.Publish(ctx, event); err != nil {
		appLogger.Error(ctx, "Erro ao publicar evento", map[string]interface{}{
			"event_name": event.EventName(),
			"payload":    event.Payload(),
			"error":      err,
		})
		return
	} else {
		appLogger.Info(ctx, "Evento publicado com sucesso", map[string]interface{}{
			"event_name": event.EventName(),
			"payload":    event.Payload(),
		})
	}

}
