package busticket

import (
	"github.com/go-chi/chi/v5"

	"github.com/mateusmacedo/go-bff/internal/busticket/application"
	"github.com/mateusmacedo/go-bff/internal/busticket/domain"
	"github.com/mateusmacedo/go-bff/internal/busticket/infrastructure"
	pkgApp "github.com/mateusmacedo/go-bff/pkg/application"
	pkgDomain "github.com/mateusmacedo/go-bff/pkg/domain"
)

type BusTicketSlice struct {
	httpHandler *infrastructure.BusTicketHTTPHandler
}

func NewBusTicketSlice(
	commandBus pkgApp.CommandBus[pkgDomain.Command[application.ReserveBusTicketData], application.ReserveBusTicketData],
	queryBus pkgApp.QueryBus[pkgDomain.Query[application.FindBusTicketData], application.FindBusTicketData, []domain.BusTicket],
	idGenerator pkgDomain.IDGenerator[string],
	logger pkgApp.AppLogger,
	eventBus pkgApp.EventBus[pkgDomain.Event[string], string],
	repository domain.BusTicketRepository,
) *BusTicketSlice {
	commandHandler := application.NewReserveBusTicketHandler(eventBus, repository, idGenerator, logger)
	queryHandler := application.NewFindBusTicketHandler(repository, logger)
	eventHandler := application.NewBusTicketBookedEventHandler(logger)

	commandBus.RegisterHandler("ReserveBusTicket", commandHandler)
	queryBus.RegisterHandler("FindBusTicket", queryHandler)
	eventBus.RegisterHandler("BusTicketBooked", eventHandler)

	httpHandler := infrastructure.NewBusTicketHTTPHandler(commandBus, queryBus)

	return &BusTicketSlice{
		httpHandler: httpHandler,
	}
}

func (s *BusTicketSlice) RegisterRoutes(router *chi.Mux) {
	s.httpHandler.RegisterRoutes(router)
}
