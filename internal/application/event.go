package application

import (
	"github.com/mateusmacedo/go-bff/pkg/domain"
)

// PassageBookedEvent representa um evento de passagem reservada.
type passageBookedEvent struct {
	data string
}

func (e passageBookedEvent) EventName() string {
	return "PassageBooked"
}

func (e passageBookedEvent) Payload() string {
	return e.data
}

// NewPassageBookedEvent cria um novo evento de passagem reservada.
func NewPassageBookedEvent(data string) domain.Event[string] {
	return passageBookedEvent{data: data}
}
