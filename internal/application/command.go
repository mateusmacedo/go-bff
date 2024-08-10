package application

import (
	"time"

	"github.com/mateusmacedo/go-bff/pkg/domain"
)

// ReservePassageData contém os dados necessários para reservar uma passagem.
type ReservePassageData struct {
	PassengerName string
	DepartureTime time.Time
	SeatNumber    int
	Origin        string
	Destination   string
}

// reservePassageCommand é uma implementação privada de um comando para reservar uma passagem.
type reservePassageCommand struct {
	data ReservePassageData
}

func (c reservePassageCommand) CommandName() string {
	return "ReservePassage"
}

func (c reservePassageCommand) Payload() ReservePassageData {
	return c.data
}

// NewReservePassageCommand cria um novo comando para reservar uma passagem.
func NewReservePassageCommand(data ReservePassageData) domain.Command[ReservePassageData] {
	return reservePassageCommand{data: data}
}
