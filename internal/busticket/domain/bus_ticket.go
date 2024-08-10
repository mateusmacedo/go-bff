package domain

import (
	"context"
	"time"
)

type BusTicket struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	PassengerName string    `json:"passengerName" gorm:"index"`
	DepartureTime time.Time `json:"departureTime"`
	SeatNumber    int       `json:"seatNumber"`
	Origin        string    `json:"origin"`
	Destination   string    `json:"destination"`
}

type BusTicketRepository interface {
	Save(ctx context.Context, busTicket BusTicket) error

	FindByPassengerName(ctx context.Context, passengerName string) ([]BusTicket, error)
	Update(ctx context.Context, busTicket BusTicket) error
}
