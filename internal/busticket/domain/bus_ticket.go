package domain

import (
	"context"
	"time"
)

type BusTicket struct {
	ID            string
	PassengerName string
	DepartureTime time.Time
	SeatNumber    int
	Origin        string
	Destination   string
}

type BusTicketRepository interface {
	Save(ctx context.Context, busTicket BusTicket) error

	FindByPassengerName(ctx context.Context, passengerName string) ([]BusTicket, error)
	Update(ctx context.Context, busTicket BusTicket) error
}
