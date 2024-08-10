// internal/infrastructure/repository.go
package infrastructure

import (
	"context"
	"errors"
	"sync"

	"github.com/mateusmacedo/go-bff/internal/domain"
	pkgApp "github.com/mateusmacedo/go-bff/pkg/application"
)

// InMemoryPassageRepository é uma implementação em memória do repositório de passagens.
type InMemoryPassageRepository struct {
	mu     sync.RWMutex
	data   map[string]domain.Passage
	logger pkgApp.AppLogger
}

func NewInMemoryPassageRepository(logger pkgApp.AppLogger) *InMemoryPassageRepository {
	return &InMemoryPassageRepository{
		data:   make(map[string]domain.Passage),
		logger: logger,
	}
}

func (r *InMemoryPassageRepository) Save(ctx context.Context, passage domain.Passage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[passage.ID]; exists {
		r.logger.Error(ctx, "passage already exists", map[string]interface{}{
			"passage": passage,
		})
		return errors.New("passage already exists")
	}

	r.logger.Info(ctx, "passage saved", map[string]interface{}{
		"passage": passage,
	})
	r.data[passage.ID] = passage

	return nil
}

func (r *InMemoryPassageRepository) FindByID(ctx context.Context, id string) (domain.Passage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	passage, exists := r.data[id]
	if !exists {
		r.logger.Error(ctx, "passage not found", map[string]interface{}{
			"id": id,
		})
		return domain.Passage{}, errors.New("passage not found")
	}

	r.logger.Info(ctx, "passage found", map[string]interface{}{
		"passage": passage,
	})

	return passage, nil
}

func (r *InMemoryPassageRepository) Update(ctx context.Context, passage domain.Passage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[passage.ID]; !exists {
		r.logger.Error(ctx, "passage not found", map[string]interface{}{
			"passage": passage,
		})
		return errors.New("passage not found")
	}

	r.logger.Info(ctx, "passage updated", map[string]interface{}{
		"passage": passage,
	})
	r.data[passage.ID] = passage

	return nil
}

// Método auxiliar para obter todos os dados (apenas para depuração).
func (r *InMemoryPassageRepository) GetData() map[string]domain.Passage {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.data
}
