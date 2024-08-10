package application

import (
	"context"

	"github.com/mateusmacedo/go-bff/internal/domain"
	pkgApp "github.com/mateusmacedo/go-bff/pkg/application"
	pkgDomain "github.com/mateusmacedo/go-bff/pkg/domain"
)

// reservePassageHandler manipula o comando de reserva de passagem.
type reservePassageHandler struct {
	repository  domain.PassageRepository
	idGenerator pkgDomain.IDGenerator[string]
}

func (h *reservePassageHandler) Handle(ctx context.Context, command pkgDomain.Command[ReservePassageData]) error {
	select {
	case <-ctx.Done():
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

		if err := h.repository.Save(passage); err != nil {
			return err
		}

		return nil
	}
}

// NewReservePassageHandler cria um novo handler para o comando de reserva de passagem.
func NewReservePassageHandler(repo domain.PassageRepository, idGenerator pkgDomain.IDGenerator[string]) pkgApp.CommandHandler[pkgDomain.Command[ReservePassageData], ReservePassageData] {
	return &reservePassageHandler{
		repository:  repo,
		idGenerator: idGenerator,
	}
}

// findPassageHandler manipula a consulta para encontrar uma passagem.
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
