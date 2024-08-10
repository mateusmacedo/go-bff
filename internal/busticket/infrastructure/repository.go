// internal/infrastructure/repository.go
package infrastructure

import (
	"context"
	"errors"
	"sync"

	"github.com/mateusmacedo/go-bff/internal/busticket/domain"
	pkgApp "github.com/mateusmacedo/go-bff/pkg/application"
)

// InMemoryBusTicketRepository é uma implementação em memória do repositório de passagens.
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
		r.logger.Error(ctx, "busTicket already exists", map[string]interface{}{
			"busTicket": busTicket,
		})
		return errors.New("busTicket already exists")
	}

	r.logger.Info(ctx, "busTicket saved", map[string]interface{}{
		"busTicket": busTicket,
	})
	r.data[busTicket.ID] = busTicket

	return nil
}

func (r *InMemoryBusTicketRepository) FindByID(ctx context.Context, id string) (domain.BusTicket, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	busTicket, exists := r.data[id]
	if !exists {
		r.logger.Error(ctx, "busTicket not found", map[string]interface{}{
			"id": id,
		})
		return domain.BusTicket{}, errors.New("busTicket not found")
	}

	r.logger.Info(ctx, "busTicket found", map[string]interface{}{
		"busTicket": busTicket,
	})

	return busTicket, nil
}

func (r *InMemoryBusTicketRepository) Update(ctx context.Context, busTicket domain.BusTicket) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[busTicket.ID]; !exists {
		r.logger.Error(ctx, "busTicket not found", map[string]interface{}{
			"busTicket": busTicket,
		})
		return errors.New("busTicket not found")
	}

	r.logger.Info(ctx, "busTicket updated", map[string]interface{}{
		"busTicket": busTicket,
	})
	r.data[busTicket.ID] = busTicket

	return nil
}

// Método auxiliar para obter todos os dados (apenas para depuração).
func (r *InMemoryBusTicketRepository) GetData() map[string]domain.BusTicket {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.data
}
