package infrastructure

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/mateusmacedo/go-bff/internal/busticket/application"
	"github.com/mateusmacedo/go-bff/internal/busticket/domain"
	pkgApp "github.com/mateusmacedo/go-bff/pkg/application"
	pkgDomain "github.com/mateusmacedo/go-bff/pkg/domain"
)

// BusTicketHTTPHandler manipula as rotas HTTP para comandos e consultas de passagens.
type BusTicketHTTPHandler struct {
	commandBus pkgApp.CommandBus[pkgDomain.Command[application.ReserveBusTicketData], application.ReserveBusTicketData]
	queryBus   pkgApp.QueryBus[pkgDomain.Query[application.FindBusTicketData], application.FindBusTicketData, domain.BusTicket]
}

// NewBusTicketHTTPHandler cria uma nova instância de BusTicketHTTPHandler.
func NewBusTicketHTTPHandler(
	commandBus pkgApp.CommandBus[pkgDomain.Command[application.ReserveBusTicketData], application.ReserveBusTicketData],
	queryBus pkgApp.QueryBus[pkgDomain.Query[application.FindBusTicketData], application.FindBusTicketData, domain.BusTicket],
) *BusTicketHTTPHandler {
	return &BusTicketHTTPHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

// HandleReserveBusTicket manipula o comando de reserva de passagem.
func (h *BusTicketHTTPHandler) HandleReserveBusTicket(w http.ResponseWriter, r *http.Request) {
	var cmd application.ReserveBusTicketData
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	command := application.NewReserveBusTicketCommand(cmd)

	// Criar um contexto com timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := h.commandBus.Dispatch(ctx, command); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "BusTicket reserved successfully!"})
}

// HandleFindBusTicket manipula a consulta para encontrar uma passagem.
func (h *BusTicketHTTPHandler) HandleFindBusTicket(w http.ResponseWriter, r *http.Request) {
	busTicketID := chi.URLParam(r, "busTicketID")
	findBusTicketData := application.FindBusTicketData{
		BusTicketID: busTicketID,
	}
	query := application.NewFindBusTicketQuery(findBusTicketData)

	// Criar um contexto com timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	busTicket, err := h.queryBus.Dispatch(ctx, query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(busTicket)
}

// RegisterRoutes registra as rotas HTTP para o manipulador de passagens.
func (h *BusTicketHTTPHandler) RegisterRoutes(router chi.Router) {
	router.Post("/bustickets", h.HandleReserveBusTicket)
	router.Get("/bustickets/{busTicketID}", h.HandleFindBusTicket)
}
