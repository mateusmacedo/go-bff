package application

import (
	"context"

	"github.com/mateusmacedo/go-bff/internal/domain"
	pkgApp "github.com/mateusmacedo/go-bff/pkg/application"
	pkgDomain "github.com/mateusmacedo/go-bff/pkg/domain"
)

type reservePassageHandler struct {
	repository  domain.PassageRepository
	idGenerator pkgDomain.IDGenerator[string]
	logger      pkgApp.AppLogger
}

func (h *reservePassageHandler) Handle(ctx context.Context, command pkgDomain.Command[ReservePassageData]) error {
	select {
	case <-ctx.Done():
		h.logger.Info(ctx, "Contexto cancelado", nil)
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

func NewReservePassageHandler(repo domain.PassageRepository, idGenerator pkgDomain.IDGenerator[string], logger pkgApp.AppLogger) pkgApp.CommandHandler[pkgDomain.Command[ReservePassageData], ReservePassageData] {
	return &reservePassageHandler{
		repository:  repo,
		idGenerator: idGenerator,
		logger:      logger,
	}
}

type findPassageHandler struct {
	repository domain.PassageRepository
	logger     pkgApp.AppLogger
}

func (h *findPassageHandler) Handle(ctx context.Context, query pkgDomain.Query[FindPassageData]) (domain.Passage, error) {
	select {
	case <-ctx.Done():
		h.logger.Info(ctx, "Contexto cancelado", nil)
		return domain.Passage{}, ctx.Err()
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
			return domain.Passage{}, err
		}
		h.logger.Info(ctx, "Passagem encontrada com sucesso", map[string]interface{}{
			"passage": passage,
		})
		return passage, nil
	}
}

// NewFindPassageHandler cria um novo handler para a consulta de encontrar passagem.
func NewFindPassageHandler(repo domain.PassageRepository, logger pkgApp.AppLogger) pkgApp.QueryHandler[pkgDomain.Query[FindPassageData], FindPassageData, domain.Passage] {
	return &findPassageHandler{
		repository: repo,
		logger:     logger,
	}
}
