package application

import (
	"context"

	"github.com/mateusmacedo/go-bff/internal/busticket/domain"
	pkgApp "github.com/mateusmacedo/go-bff/pkg/application"
	pkgDomain "github.com/mateusmacedo/go-bff/pkg/domain"
)

type reserveBusTicketHandler struct {
	repository  domain.BusTicketRepository
	idGenerator pkgDomain.IDGenerator[string]
	logger      pkgApp.AppLogger
}

func (h *reserveBusTicketHandler) Handle(ctx context.Context, command pkgDomain.Command[ReserveBusTicketData]) error {
	select {
	case <-ctx.Done():
		h.logger.Info(ctx, "Contexto cancelado", nil)
		return ctx.Err()
	default:
		data := command.Payload()
		passage := domain.BusTicket{
			ID:            h.idGenerator(),
			PassengerName: data.PassengerName,
			DepartureTime: data.DepartureTime,
			SeatNumber:    data.SeatNumber,
			Origin:        data.Origin,
			Destination:   data.Destination,
		}

		h.logger.Info(ctx, "Salvando passagem", map[string]interface{}{
			"passage": passage,
		})
		if err := h.repository.Save(ctx, passage); err != nil {
			h.logger.Error(ctx, "Erro ao salvar passagem", map[string]interface{}{
				"passage": passage,
				"error":   err,
			})
			return err
		}

		h.logger.Info(ctx, "Passagem salva com sucesso", map[string]interface{}{
			"passage": passage,
		})
		return nil
	}
}

func NewReserveBusTicketHandler(repo domain.BusTicketRepository, idGenerator pkgDomain.IDGenerator[string], logger pkgApp.AppLogger) pkgApp.CommandHandler[pkgDomain.Command[ReserveBusTicketData], ReserveBusTicketData] {
	return &reserveBusTicketHandler{
		repository:  repo,
		idGenerator: idGenerator,
		logger:      logger,
	}
}

type findPassageHandler struct {
	repository domain.BusTicketRepository
	logger     pkgApp.AppLogger
}

func (h *findPassageHandler) Handle(ctx context.Context, query pkgDomain.Query[FindBusTicketData]) (domain.BusTicket, error) {
	select {
	case <-ctx.Done():
		h.logger.Info(ctx, "Contexto cancelado", nil)
		return domain.BusTicket{}, ctx.Err()
	default:
		data := query.Payload()
		passage, err := h.repository.FindByID(ctx, data.PassageID)
		h.logger.Info(ctx, "Encontrando passagem", map[string]interface{}{
			"passage_id": data.PassageID,
		})
		if err != nil {
			h.logger.Error(ctx, "Erro ao encontrar passagem", map[string]interface{}{
				"passage_id": data.PassageID,
				"error":      err,
			})
			return domain.BusTicket{}, err
		}
		h.logger.Info(ctx, "Passagem encontrada com sucesso", map[string]interface{}{
			"passage": passage,
		})
		return passage, nil
	}
}

// NewFindBusTicketHandler cria um novo handler para a consulta de encontrar passagem.
func NewFindBusTicketHandler(repo domain.BusTicketRepository, logger pkgApp.AppLogger) pkgApp.QueryHandler[pkgDomain.Query[FindBusTicketData], FindBusTicketData, domain.BusTicket] {
	return &findPassageHandler{
		repository: repo,
		logger:     logger,
	}
}
