package domain

import (
	"context"
	"time"
)

// Passage representa uma passagem rodoviária.
type Passage struct {
	ID            string
	PassengerName string
	DepartureTime time.Time
	SeatNumber    int
	Origin        string
	Destination   string
}

// PassageRepository define a interface para o repositório de passagens.
type PassageRepository interface {
	Save(ctx context.Context, passage Passage) error
	FindByID(ctx context.Context, id string) (Passage, error)
	Update(ctx context.Context, passage Passage) error
}
