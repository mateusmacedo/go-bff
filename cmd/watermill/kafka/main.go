package main

import (
	"context"
	"time"

	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/google/uuid"

	app "github.com/mateusmacedo/go-bff/internal/application"
	"github.com/mateusmacedo/go-bff/internal/domain"
	"github.com/mateusmacedo/go-bff/internal/infrastructure"
	pkgDomain "github.com/mateusmacedo/go-bff/pkg/domain"
	"github.com/mateusmacedo/go-bff/pkg/infrastructure/kafka/adapter"
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

	// Configuração do marshaler
	marshaler := kafka.DefaultMarshaler{}

	// Configuração do publisher para Kafka
	publisherConfig := kafka.PublisherConfig{
		Brokers:   []string{"localhost:9092"},
		Marshaler: marshaler,
	}
	// Criando um contexto com timeout mais longo
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	publisher, err := kafka.NewPublisher(publisherConfig, logger)
	if err != nil {
		appLogger.Error(ctx, "Erro ao criar publisher", map[string]interface{}{
			"error": err,
		})
	}
	defer publisher.Close()

	// Configuração do subscriber para Kafka
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

	// Inicialize os tópicos se ainda não existirem
	err = subscriber.SubscribeInitialize("ReservePassage")
	if err != nil {
		appLogger.Error(ctx, "Erro ao inicializar o tópico 'ReservePassage'", map[string]interface{}{
			"error": err,
		})
	}

	err = subscriber.SubscribeInitialize("FindPassage")
	if err != nil {
		appLogger.Error(ctx, "Erro ao inicializar o tópico 'FindPassage'", map[string]interface{}{
			"error": err,
		})
	}

	// Configuração do repositório
	repository := infrastructure.NewInMemoryPassageRepository(appLogger)

	// Gerador de ID
	idGenerator := func() string {
		return uuid.New().String()
	}

	// Criação dos handlers
	reserveHandler := app.NewReservePassageHandler(repository, idGenerator, appLogger)
	findHandler := app.NewFindPassageHandler(repository, appLogger)

	// Criação dos barramentos usando Kafka
	commandBus := adapter.NewKafkaCommandBus[pkgDomain.Command[app.ReservePassageData], app.ReservePassageData](publisher, subscriber, appLogger)
	queryBus := adapter.NewKafkaQueryBus[pkgDomain.Query[app.FindPassageData], app.FindPassageData, domain.Passage](publisher, subscriber, appLogger)
	eventBus := adapter.NewKafkaEventBus[pkgDomain.Event[string], string](publisher, subscriber, appLogger)

	// Registro dos handlers nos barramentos
	commandBus.RegisterHandler("ReservePassage", reserveHandler)
	queryBus.RegisterHandler("FindPassage", findHandler)

	// Criando um comando de reserva de passagem
	reserveData := app.ReservePassageData{
		PassengerName: "John Doe",
		DepartureTime: time.Now().Add(24 * time.Hour).Truncate(time.Second),
		SeatNumber:    12,
		Origin:        "City A",
		Destination:   "City B",
	}
	command := app.NewReservePassageCommand(reserveData)

	appLogger.Info(ctx, "Despachando o comando para reservar passagem...", map[string]interface{}{
		"command_name": command.CommandName(),
		"payload":      command.Payload(),
	})
	// Despachando o comando
	if err := commandBus.Dispatch(ctx, command); err != nil {
		appLogger.Error(ctx, "Erro ao reservar passagem", map[string]interface{}{
			"command_name": command.CommandName(),
			"payload":      command.Payload(),
			"error":        err,
		})
		return
	}
	appLogger.Info(ctx, "Comando de reserva de passagem despachado com sucesso", nil)
	// Espera mais longa para permitir o processamento das mensagens
	time.Sleep(15 * time.Second)

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

	appLogger.Info(ctx, "Despachando a consulta para encontrar passagem...", map[string]interface{}{
		"query_name": query.QueryName(),
		"payload":    query.Payload(),
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
	event := app.NewPassageBookedEvent("Passage successfully booked for John Doe")
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
