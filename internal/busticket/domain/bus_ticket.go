package domain

import (
	"context"
	"time"
)

// BusTicket representa uma passagem rodoviária.
type BusTicket struct {
	ID            string
	PassengerName string
	DepartureTime time.Time
	SeatNumber    int
	Origin        string
	Destination   string
}

// BusTicketRepository define a interface para o repositório de passagens.
type BusTicketRepository interface {
	Save(ctx context.Context, busTicket BusTicket) error
	// FindByID(ctx context.Context, id string) (BusTicket, error)
	FindByPassengerName(ctx context.Context, passengerName string) ([]BusTicket, error)
	Update(ctx context.Context, busTicket BusTicket) error
}
