package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/google/uuid"

	app "github.com/mateusmacedo/go-bff/internal/application"
	"github.com/mateusmacedo/go-bff/internal/domain"
	"github.com/mateusmacedo/go-bff/internal/infrastructure"
	pkgDomain "github.com/mateusmacedo/go-bff/pkg/domain"
	"github.com/mateusmacedo/go-bff/pkg/infrastructure/channels/adapter"
)

func main() {
	// Configuração do logger do Watermill
	logger := watermill.NewStdLogger(false, false)

	// Configuração do publisher e subscriber em memória
	pubSub := gochannel.NewGoChannel(gochannel.Config{}, logger)

	// Configuração do repositório
	repository := infrastructure.NewInMemoryPassageRepository()

	// Gerador de ID
	idGenerator := func() string {
		return uuid.New().String()
	}

	// Criação dos handlers
	reserveHandler := app.NewReservePassageHandler(repository, idGenerator)
	findHandler := app.NewFindPassageHandler(repository)

	// Criação dos barramentos usando Watermill
	commandBus := adapter.NewWatermillCommandBus[pkgDomain.Command[app.ReservePassageData], app.ReservePassageData](pubSub, pubSub)
	queryBus := adapter.NewWatermillQueryBus[pkgDomain.Query[app.FindPassageData], app.FindPassageData, domain.Passage](pubSub, pubSub)
	eventBus := adapter.NewWatermillEventBus[pkgDomain.Event[string], string](pubSub)

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
