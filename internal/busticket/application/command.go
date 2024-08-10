package application

import (
	"time"

	"github.com/mateusmacedo/go-bff/pkg/domain"
)

type ReserveBusTicketData struct {
	PassengerName string
	DepartureTime time.Time
	SeatNumber    int
	Origin        string
	Destination   string
}

type reserveBusTicketCommand struct {
	data ReserveBusTicketData
}

func (c reserveBusTicketCommand) CommandName() string {
	return "ReserveBusTicket"
}

func (c reserveBusTicketCommand) Payload() ReserveBusTicketData {
	return c.data
}

func NewReserveBusTicketCommand(data ReserveBusTicketData) domain.Command[ReserveBusTicketData] {
	return reserveBusTicketCommand{data: data}
}
