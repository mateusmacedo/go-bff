package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/google/uuid"

	app "github.com/mateusmacedo/go-bff/internal/application"
	"github.com/mateusmacedo/go-bff/internal/domain"
	"github.com/mateusmacedo/go-bff/internal/infrastructure"
	pkgDomain "github.com/mateusmacedo/go-bff/pkg/domain"
	"github.com/mateusmacedo/go-bff/pkg/infrastructure/kafka/adapter"
)

func main() {
	// Configuração do logger do Watermill
	logger := watermill.NewStdLogger(false, false)

	// Configuração do marshaler
	marshaler := kafka.DefaultMarshaler{}

	// Configuração do publisher para Kafka
	publisherConfig := kafka.PublisherConfig{
		Brokers:   []string{"localhost:9092"},
		Marshaler: marshaler,
	}

	publisher, err := kafka.NewPublisher(publisherConfig, logger)
	if err != nil {
		log.Fatalf("failed to create Kafka publisher: %v", err)
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
		log.Fatalf("failed to create Kafka subscriber: %v", err)
	}
	defer subscriber.Close()

	// Inicialize os tópicos se ainda não existirem
	err = subscriber.SubscribeInitialize("ReservePassage")
	if err != nil {
		log.Fatalf("failed to initialize Kafka topic 'ReservePassage': %v", err)
	}

	err = subscriber.SubscribeInitialize("FindPassage")
	if err != nil {
		log.Fatalf("failed to initialize Kafka topic 'FindPassage': %v", err)
	}

	// Configuração do repositório
	repository := infrastructure.NewInMemoryPassageRepository()

	// Gerador de ID
	idGenerator := func() string {
		return uuid.New().String()
	}

	// Criação dos handlers
	reserveHandler := app.NewReservePassageHandler(repository, idGenerator)
	findHandler := app.NewFindPassageHandler(repository)

	// Criação dos barramentos usando Kafka
	commandBus := adapter.NewKafkaCommandBus[pkgDomain.Command[app.ReservePassageData], app.ReservePassageData](publisher, subscriber)
	queryBus := adapter.NewKafkaQueryBus[pkgDomain.Query[app.FindPassageData], app.FindPassageData, domain.Passage](publisher, subscriber)
	eventBus := adapter.NewKafkaEventBus[pkgDomain.Event[string], string](publisher, subscriber)

	// Registro dos handlers nos barramentos
	commandBus.RegisterHandler("ReservePassage", reserveHandler)
	queryBus.RegisterHandler("FindPassage", findHandler)

	// Criando um contexto com timeout mais longo
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
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

	fmt.Println("Despachando o comando de reserva de passagem...")
	// Despachando o comando
	if err := commandBus.Dispatch(ctx, command); err != nil {
		fmt.Println("Erro ao reservar passagem:", err)
		return
	}
	fmt.Println("Passagem reservada com sucesso!")

	// Espera mais longa para permitir o processamento das mensagens
	time.Sleep(15 * time.Second)

	// Imprimir todos os dados do repositório
	fmt.Println("Dados do repositório após a reserva:", repository.GetData())

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

	fmt.Println("Despachando a consulta para encontrar a passagem...")
	// Despachando a consulta
	passage, err := queryBus.Dispatch(ctx, query)
	if err != nil {
		fmt.Println("Erro ao encontrar passagem:", err)
	} else {
		fmt.Printf("Passagem encontrada: %+v\n", passage)
	}

	// Exemplo de publicação de um evento
	event := app.NewPassageBookedEvent("Passage successfully booked for John Doe")
	if err := eventBus.Publish(ctx, event); err != nil {
		fmt.Println("Erro ao publicar evento:", err)
	}
}
