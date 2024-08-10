package application

import (
	"github.com/mateusmacedo/go-bff/pkg/domain"
)

type busTicketBookedEvent struct {
	data string
}

func (e busTicketBookedEvent) EventName() string {
	return "BusTicketBooked"
}

func (e busTicketBookedEvent) Payload() string {
	return e.data
}

func NewBusTicketBookedEvent(data string) domain.Event[string] {
	return busTicketBookedEvent{data: data}
}
