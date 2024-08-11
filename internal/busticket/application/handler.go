package application

import (
	"context"

	"github.com/mateusmacedo/go-bff/internal/busticket/domain"
	pkgApp "github.com/mateusmacedo/go-bff/pkg/application"
	pkgDomain "github.com/mateusmacedo/go-bff/pkg/domain"
)

type reserveBusTicketHandler struct {
	eventBus    pkgApp.EventBus[pkgDomain.Event[string], string]
	repository  domain.BusTicketRepository
	idGenerator pkgDomain.IDGenerator[string]
	logger      pkgApp.AppLogger
}

func (h *reserveBusTicketHandler) Handle(ctx context.Context, command pkgDomain.Command[ReserveBusTicketData]) error {
	if ctx.Err() != nil {
		pkgApp.LogError(ctx, h.logger, "Contexto cancelado", ctx.Err(), nil)
		return ctx.Err()
	}

	data := command.Payload()
	busTicket := domain.BusTicket{
		ID:            h.idGenerator(),
		PassengerName: data.PassengerName,
		DepartureTime: data.DepartureTime,
		SeatNumber:    data.SeatNumber,
		Origin:        data.Origin,
		Destination:   data.Destination,
	}

	h.logger.Info(ctx, "Salvando passagem", map[string]interface{}{"id": busTicket.ID})
	if err := h.repository.Save(ctx, busTicket); err != nil {
		pkgApp.LogError(ctx, h.logger, "Erro ao salvar passagem", err, map[string]interface{}{"bus_ticket": busTicket})
		return err
	}

	event := NewBusTicketBookedEvent("BusTicket successfully booked for " + data.PassengerName)
	if err := h.eventBus.Publish(ctx, event); err != nil {
		pkgApp.LogError(ctx, h.logger, "Erro ao publicar evento", err, nil)
		return err
	}

	h.logger.Info(ctx, "BusTicket salva com sucesso", map[string]interface{}{"bus_ticket": busTicket})
	return nil
}

func NewReserveBusTicketHandler(eventBus pkgApp.EventBus[pkgDomain.Event[string], string], repo domain.BusTicketRepository, idGenerator pkgDomain.IDGenerator[string], logger pkgApp.AppLogger) pkgApp.CommandHandler[pkgDomain.Command[ReserveBusTicketData], ReserveBusTicketData] {
	return &reserveBusTicketHandler{
		eventBus:    eventBus,
		repository:  repo,
		idGenerator: idGenerator,
		logger:      logger,
	}
}

type findBusTicketHandler struct {
	repository domain.BusTicketRepository
	logger     pkgApp.AppLogger
}

func (h *findBusTicketHandler) Handle(ctx context.Context, query pkgDomain.Query[FindBusTicketData]) ([]domain.BusTicket, error) {
	if ctx.Err() != nil {
		pkgApp.LogError(ctx, h.logger, "Contexto cancelado", ctx.Err(), nil)
		return nil, ctx.Err()
	}

	data := query.Payload()
	busTicket, err := h.repository.FindByPassengerName(ctx, data.PassengerName)
	pkgApp.LogInfo(ctx, h.logger, "BusTicket encontrado", map[string]interface{}{"bus_ticket": busTicket})
	if err != nil {
		pkgApp.LogError(ctx, h.logger, "Erro ao encontrar passagem", err, map[string]interface{}{"passenger_name": data.PassengerName})
		return nil, err
	}

	pkgApp.LogInfo(ctx, h.logger, "Passagem encontrada", map[string]interface{}{"bus_ticket": busTicket})
	return busTicket, nil
}

func NewFindBusTicketHandler(repo domain.BusTicketRepository, logger pkgApp.AppLogger) pkgApp.QueryHandler[pkgDomain.Query[FindBusTicketData], FindBusTicketData, []domain.BusTicket] {
	return &findBusTicketHandler{
		repository: repo,
		logger:     logger,
	}
}

type busTicketBookedEventHandler struct {
	logger pkgApp.AppLogger
}

func (h *busTicketBookedEventHandler) Handle(ctx context.Context, event pkgDomain.Event[string]) error {
	if ctx.Err() != nil {
		pkgApp.LogError(ctx, h.logger, "Contexto cancelado", ctx.Err(), nil)
		return ctx.Err()
	}

	pkgApp.LogInfo(ctx, h.logger, "Evento recebido", map[string]interface{}{"event": event})
	return nil
}

func NewBusTicketBookedEventHandler(logger pkgApp.AppLogger) pkgApp.EventHandler[pkgDomain.Event[string], string] {
	return &busTicketBookedEventHandler{
		logger: logger,
	}
}
