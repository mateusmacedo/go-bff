package application

import (
	"time"

	"github.com/mateusmacedo/go-bff/pkg/domain"
)

// ReserveBusTicketData contém os dados necessários para reservar uma passagem.
type ReserveBusTicketData struct {
	PassengerName string
	DepartureTime time.Time
	SeatNumber    int
	Origin        string
	Destination   string
}

// reserveBusTicketCommand é uma implementação privada de um comando para reservar uma passagem.
type reserveBusTicketCommand struct {
	data ReserveBusTicketData
}

func (c reserveBusTicketCommand) CommandName() string {
	return "ReserveBusTicket"
}

func (c reserveBusTicketCommand) Payload() ReserveBusTicketData {
	return c.data
}

// NewReserveBusTicketCommand cria um novo comando para reservar uma passagem.
func NewReserveBusTicketCommand(data ReserveBusTicketData) domain.Command[ReserveBusTicketData] {
	return reserveBusTicketCommand{data: data}
}
