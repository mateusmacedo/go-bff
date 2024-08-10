package busticket

import (
	"github.com/go-chi/chi/v5"

	"github.com/mateusmacedo/go-bff/internal/busticket/application"
	"github.com/mateusmacedo/go-bff/internal/busticket/domain"
	"github.com/mateusmacedo/go-bff/internal/busticket/infrastructure"
	pkgApp "github.com/mateusmacedo/go-bff/pkg/application"
	pkgDomain "github.com/mateusmacedo/go-bff/pkg/domain"
)

// BusTicketSlice estrutura que reúne todos os componentes do contexto de BusTicket.
type BusTicketSlice struct {
	httpHandler *infrastructure.BusTicketHTTPHandler
}

// NewBusTicketSlice cria uma nova instância de BusTicketSlice.
func NewBusTicketSlice(
	commandBus pkgApp.CommandBus[pkgDomain.Command[application.ReserveBusTicketData], application.ReserveBusTicketData],
	queryBus pkgApp.QueryBus[pkgDomain.Query[application.FindBusTicketData], application.FindBusTicketData, domain.BusTicket],
	idGenerator pkgDomain.IDGenerator[string],
	logger pkgApp.AppLogger,
	eventBus pkgApp.EventBus[pkgDomain.Event[string], string],
) *BusTicketSlice {
	// Inicializando o repositório
	repo := infrastructure.NewInMemoryBusTicketRepository(logger)

	// Criando os manipuladores de comando e consulta
	commandHandler := application.NewReserveBusTicketHandler(eventBus, repo, idGenerator, logger)
	queryHandler := application.NewFindBusTicketHandler(repo, logger)
	eventHandler := application.NewBusTicketBookedEventHandler(logger)

	// Registrando manipuladores no barramento
	commandBus.RegisterHandler("ReserveBusTicket", commandHandler)
	queryBus.RegisterHandler("FindBusTicket", queryHandler)
	eventBus.RegisterHandler("BusTicketBooked", eventHandler)

	// Criando o manipulador HTTP
	httpHandler := infrastructure.NewBusTicketHTTPHandler(commandBus, queryBus)

	return &BusTicketSlice{
		httpHandler: httpHandler,
	}
}

// RegisterRoutes registra as rotas HTTP do slice de BusTicket.
func (s *BusTicketSlice) RegisterRoutes(router *chi.Mux) {
	s.httpHandler.RegisterRoutes(router)
}
