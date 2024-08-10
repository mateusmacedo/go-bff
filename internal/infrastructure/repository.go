// internal/infrastructure/repository.go
package infrastructure

import (
	"errors"
	"sync"

	"github.com/mateusmacedo/go-bff/internal/domain"
)

// InMemoryPassageRepository é uma implementação em memória do repositório de passagens.
type InMemoryPassageRepository struct {
	mu   sync.RWMutex
	data map[string]domain.Passage
}

func NewInMemoryPassageRepository() *InMemoryPassageRepository {
	return &InMemoryPassageRepository{
		data: make(map[string]domain.Passage),
	}
}

func (r *InMemoryPassageRepository) Save(passage domain.Passage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[passage.ID]; exists {
		return errors.New("passage already exists")
	}
	r.data[passage.ID] = passage
	return nil
}

func (r *InMemoryPassageRepository) FindByID(id string) (domain.Passage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	passage, exists := r.data[id]
	if !exists {
		return domain.Passage{}, errors.New("passage not found")
	}
	return passage, nil
}

func (r *InMemoryPassageRepository) Update(passage domain.Passage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[passage.ID]; !exists {
		return errors.New("passage not found")
	}
	r.data[passage.ID] = passage
	return nil
}

// Método auxiliar para obter todos os dados (apenas para depuração).
func (r *InMemoryPassageRepository) GetData() map[string]domain.Passage {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.data
}
