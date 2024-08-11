package infrastructure

import (
	"context"
	"errors"
	"sync"

	"github.com/mateusmacedo/go-bff/internal/busticket/domain"
	"github.com/mateusmacedo/go-bff/pkg/application"
	pkgApp "github.com/mateusmacedo/go-bff/pkg/application"
)

type InMemoryBusTicketRepository struct {
	mu     sync.RWMutex
	data   map[string]domain.BusTicket
	logger pkgApp.AppLogger
}

func NewInMemoryBusTicketRepository(logger pkgApp.AppLogger) *InMemoryBusTicketRepository {
	return &InMemoryBusTicketRepository{
		data:   make(map[string]domain.BusTicket),
		logger: logger,
	}
}

func (r *InMemoryBusTicketRepository) Save(ctx context.Context, busTicket domain.BusTicket) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[busTicket.ID]; exists {
		application.LogInfo(ctx, r.logger, "busTicket already exists", map[string]interface{}{
			"busTicket": busTicket,
		})
		return errors.New("busTicket already exists")
	}

	application.LogInfo(ctx, r.logger, "busTicket saved", map[string]interface{}{
		"busTicket": busTicket,
	})
	r.data[busTicket.ID] = busTicket

	return nil
}

func (r *InMemoryBusTicketRepository) FindByPassengerName(ctx context.Context, passengerName string) ([]domain.BusTicket, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var busTickets []domain.BusTicket
	for _, busTicket := range r.data {
		if busTicket.PassengerName == passengerName {
			busTickets = append(busTickets, busTicket)
		}
	}

	application.LogInfo(ctx, r.logger, "busTickets found", map[string]interface{}{
		"passengerName": passengerName,
		"busTickets":    busTickets,
	})

	return busTickets, nil
}

func (r *InMemoryBusTicketRepository) Update(ctx context.Context, busTicket domain.BusTicket) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[busTicket.ID]; !exists {
		application.LogInfo(ctx, r.logger, "busTicket not found", map[string]interface{}{
			"busTicket": busTicket,
		})
		return errors.New("busTicket not found")
	}

	application.LogInfo(ctx, r.logger, "busTicket updated", map[string]interface{}{
		"busTicket": busTicket,
	})

	r.data[busTicket.ID] = busTicket

	return nil
}

func (r *InMemoryBusTicketRepository) GetData() map[string]domain.BusTicket {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.data
}
