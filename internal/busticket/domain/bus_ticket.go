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
	Save(ctx context.Context, passage BusTicket) error
	FindByID(ctx context.Context, id string) (BusTicket, error)
	Update(ctx context.Context, passage BusTicket) error
}
