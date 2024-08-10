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

type BusTicketHTTPHandler struct {
	commandBus pkgApp.CommandBus[pkgDomain.Command[application.ReserveBusTicketData], application.ReserveBusTicketData]
	queryBus   pkgApp.QueryBus[pkgDomain.Query[application.FindBusTicketData], application.FindBusTicketData, []domain.BusTicket]
}

func NewBusTicketHTTPHandler(
	commandBus pkgApp.CommandBus[pkgDomain.Command[application.ReserveBusTicketData], application.ReserveBusTicketData],
	queryBus pkgApp.QueryBus[pkgDomain.Query[application.FindBusTicketData], application.FindBusTicketData, []domain.BusTicket],
) *BusTicketHTTPHandler {
	return &BusTicketHTTPHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

func (h *BusTicketHTTPHandler) HandleReserveBusTicket(w http.ResponseWriter, r *http.Request) {
	var cmd application.ReserveBusTicketData
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		handleError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	command := application.NewReserveBusTicketCommand(cmd)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := h.commandBus.Dispatch(ctx, command); err != nil {
		handleError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"message": "Bus ticket reserved", "data": cmd}); err != nil {
		handleError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *BusTicketHTTPHandler) HandleFindBusTicket(w http.ResponseWriter, r *http.Request) {
	busTicketID := chi.URLParam(r, "busTicketID")
	findBusTicketData := application.FindBusTicketData{
		PassengerName: busTicketID,
	}
	query := application.NewFindBusTicketQuery(findBusTicketData)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	busTicket, err := h.queryBus.Dispatch(ctx, query)
	if err != nil {
		handleError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(busTicket); err != nil {
		handleError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *BusTicketHTTPHandler) RegisterRoutes(router chi.Router) {
	router.Post("/bustickets", h.HandleReserveBusTicket)
	router.Get("/bustickets/{busTicketID}", h.HandleFindBusTicket)
}

func handleError(w http.ResponseWriter, message string, statusCode int) {
	http.Error(w, message, statusCode)
}
