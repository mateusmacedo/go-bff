package domain

import "time"

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
	Save(passage Passage) error
	FindByID(id string) (Passage, error)
	Update(passage Passage) error
}
