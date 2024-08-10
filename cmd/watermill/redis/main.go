package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/google/uuid"

	app "github.com/mateusmacedo/go-bff/internal/application"
	"github.com/mateusmacedo/go-bff/internal/domain"
	"github.com/mateusmacedo/go-bff/internal/infrastructure"
	pkgDomain "github.com/mateusmacedo/go-bff/pkg/domain"
	"github.com/mateusmacedo/go-bff/pkg/infrastructure/redis/adapter"
)

func main() {
	// Configuração do logger do Watermill
	logger := watermill.NewStdLogger(false, false)

	// Configuração do Redis
	redisClient := adapter.NewRedisClient()
	defer redisClient.Close()

	publisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: redisClient,
	}, logger)
	if err != nil {
		log.Fatalf("Erro ao criar publisher: %v", err)
	}
	defer publisher.Close()

	subscriber, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        redisClient,
		ConsumerGroup: "my_group",
		Consumer:      "my_consumer",
	}, logger)
	if err != nil {
		log.Fatalf("Erro ao criar subscriber: %v", err)
	}
	defer subscriber.Close()

	// Configuração do repositório
	repository := infrastructure.NewInMemoryPassageRepository()

	// Gerador de ID
	idGenerator := func() string {
		return uuid.New().String()
	}

	// Criação dos handlers
	reserveHandler := app.NewReservePassageHandler(repository, idGenerator)
	findHandler := app.NewFindPassageHandler(repository)

	// Criação dos barramentos usando Redis
	commandBus := adapter.NewRedisCommandBus[pkgDomain.Command[app.ReservePassageData], app.ReservePassageData](publisher, subscriber)
	queryBus := adapter.NewRedisQueryBus[pkgDomain.Query[app.FindPassageData], app.FindPassageData, domain.Passage](publisher, subscriber)
	eventBus := adapter.NewRedisEventBus[pkgDomain.Event[string], string](publisher, subscriber)

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
		fmt.Println("Erro ao reservar passagem:", err)
		return
	}
	fmt.Println("Passagem reservada com sucesso!")

	// Espera breve para permitir o processamento das mensagens
	time.Sleep(1 * time.Second)

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

	// Verifique se o ID foi encontrado
	if passageID == "" {
		fmt.Println("Erro: ID da passagem não encontrado após a reserva.")
		return
	}

	// Criando uma consulta para encontrar uma passagem
	query := app.NewFindPassageQuery(app.FindPassageData{
		PassageID: passageID,
	})

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
