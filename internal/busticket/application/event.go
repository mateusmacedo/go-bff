package application

import (
	"github.com/mateusmacedo/go-bff/pkg/domain"
)

// PassageBookedEvent representa um evento de passagem reservada.
type busTicketBookedEvent struct {
	data string
}

func (e busTicketBookedEvent) EventName() string {
	return "BusTicketBooked"
}

func (e busTicketBookedEvent) Payload() string {
	return e.data
}

// NewBusTicketBookedEvent cria um novo evento de passagem reservada.
func NewBusTicketBookedEvent(data string) domain.Event[string] {
	return busTicketBookedEvent{data: data}
}
