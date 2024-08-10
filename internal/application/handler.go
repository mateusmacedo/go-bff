package application

import (
	"context"
	"fmt"

	"github.com/mateusmacedo/go-bff/internal/domain"
	pkgApp "github.com/mateusmacedo/go-bff/pkg/application"
	pkgDomain "github.com/mateusmacedo/go-bff/pkg/domain"
)

type reservePassageHandler struct {
	repository  domain.PassageRepository
	idGenerator pkgDomain.IDGenerator[string]
}

func (h *reservePassageHandler) Handle(ctx context.Context, command pkgDomain.Command[ReservePassageData]) error {
	select {
	case <-ctx.Done():
		fmt.Println("Contexto cancelado antes do processamento do comando.")
		return ctx.Err()
	default:
		data := command.Payload()
		passage := domain.Passage{
			ID:            h.idGenerator(),
			PassengerName: data.PassengerName,
			DepartureTime: data.DepartureTime,
			SeatNumber:    data.SeatNumber,
			Origin:        data.Origin,
			Destination:   data.Destination,
		}

		fmt.Printf("Tentando salvar passagem: %+v\n", passage)
		if err := h.repository.Save(passage); err != nil {
			fmt.Println("Erro ao salvar passagem:", err)
			return err
		}

		fmt.Println("Passagem salva com sucesso!")
		return nil
	}
}

func NewReservePassageHandler(repo domain.PassageRepository, idGenerator pkgDomain.IDGenerator[string]) pkgApp.CommandHandler[pkgDomain.Command[ReservePassageData], ReservePassageData] {
	return &reservePassageHandler{
		repository:  repo,
		idGenerator: idGenerator,
	}
}

type findPassageHandler struct {
	repository domain.PassageRepository
}

func (h *findPassageHandler) Handle(ctx context.Context, query pkgDomain.Query[FindPassageData]) (domain.Passage, error) {
	select {
	case <-ctx.Done():
		return domain.Passage{}, ctx.Err()
	default:
		data := query.Payload()
		return h.repository.FindByID(data.PassageID)
	}
}

// NewFindPassageHandler cria um novo handler para a consulta de encontrar passagem.
func NewFindPassageHandler(repo domain.PassageRepository) pkgApp.QueryHandler[pkgDomain.Query[FindPassageData], FindPassageData, domain.Passage] {
	return &findPassageHandler{
		repository: repo,
	}
}
