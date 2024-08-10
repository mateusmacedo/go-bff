package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mateusmacedo/go-bff/internal/application"
	"github.com/mateusmacedo/go-bff/internal/domain"
	"github.com/mateusmacedo/go-bff/internal/infrastructure"
	pkgDomain "github.com/mateusmacedo/go-bff/pkg/domain"
	pkgInfra "github.com/mateusmacedo/go-bff/pkg/infrastructure"
)

func main() {
	// Configuração do repositório
	repository := infrastructure.NewInMemoryPassageRepository()

	// Gerador de ID
	idGenerator := func() string {
		return uuid.New().String()
	}

	// Criação dos handlers
	reserveHandler := application.NewReservePassageHandler(repository, idGenerator)
	findHandler := application.NewFindPassageHandler(repository)

	// Criação dos barramentos
	commandBus := pkgInfra.NewSimpleCommandBus[pkgDomain.Command[application.ReservePassageData], application.ReservePassageData]()
	queryBus := pkgInfra.NewSimpleQueryBus[pkgDomain.Query[application.FindPassageData], application.FindPassageData, domain.Passage]()
	eventBus := pkgInfra.NewSimpleEventBus[pkgDomain.Event[string], string]()

	// Registro dos handlers nos barramentos
	commandBus.RegisterHandler("ReservePassage", reserveHandler)
	queryBus.RegisterHandler("FindPassage", findHandler)

	// Criando um contexto com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Criando um comando de reserva de passagem
	reserveData := application.ReservePassageData{
		PassengerName: "John Doe",
		DepartureTime: time.Now().Add(24 * time.Hour),
		SeatNumber:    12,
		Origin:        "City A",
		Destination:   "City B",
	}
	command := application.NewReservePassageCommand(reserveData)

	// Despachando o comando
	if err := commandBus.Dispatch(ctx, command); err != nil {
		fmt.Println("Erro ao reservar passagem:", err)
		return
	}
	fmt.Println("Passagem reservada com sucesso!")

	// Obtendo o ID da passagem diretamente do repositório para evitar inconsistências
	var passageID string
	for id, passage := range repository.GetData() { // Utilize GetData() se disponível
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
	query := application.NewFindPassageQuery(application.FindPassageData{
		PassageID: passageID, // Use o ID gerado corretamente
	})

	// Despachando a consulta
	passage, err := queryBus.Dispatch(ctx, query)
	if err != nil {
		fmt.Println("Erro ao encontrar passagem:", err)
	} else {
		fmt.Printf("Passagem encontrada: %+v\n", passage)
	}

	// Exemplo de publicação de um evento
	event := application.NewPassageBookedEvent("Passage successfully booked for John Doe")
	if err := eventBus.Publish(ctx, event); err != nil {
		fmt.Println("Erro ao publicar evento:", err)
	}
}
