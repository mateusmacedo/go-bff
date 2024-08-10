package infrastructure

import (
	"errors"

	"github.com/mateusmacedo/go-bff/internal/domain"
)

// InMemoryPassageRepository é uma implementação em memória do repositório de passagens.
type InMemoryPassageRepository struct {
	data map[string]domain.Passage
}

func NewInMemoryPassageRepository() *InMemoryPassageRepository {
	return &InMemoryPassageRepository{
		data: make(map[string]domain.Passage),
	}
}

func (r *InMemoryPassageRepository) Save(passage domain.Passage) error {
	if _, exists := r.data[passage.ID]; exists {
		return errors.New("passage already exists")
	}
	r.data[passage.ID] = passage
	return nil
}

func (r *InMemoryPassageRepository) FindByID(id string) (domain.Passage, error) {
	passage, exists := r.data[id]
	if !exists {
		return domain.Passage{}, errors.New("passage not found")
	}
	return passage, nil
}

func (r *InMemoryPassageRepository) Update(passage domain.Passage) error {
	if _, exists := r.data[passage.ID]; !exists {
		return errors.New("passage not found")
	}
	r.data[passage.ID] = passage
	return nil
}
