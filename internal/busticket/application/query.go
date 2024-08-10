package application

import (
	"github.com/mateusmacedo/go-bff/pkg/domain"
)

// FindBusTicketData contém os dados necessários para encontrar uma passagem.
type FindBusTicketData struct {
	PassageID string
}

// findBusTicketQuery é uma implementação privada de uma consulta para encontrar uma passagem.
type findBusTicketQuery struct {
	data FindBusTicketData
}

func (q findBusTicketQuery) QueryName() string {
	return "FindPassage"
}

func (q findBusTicketQuery) Payload() FindBusTicketData {
	return q.data
}

// NewFindBusTicketQuery cria uma nova consulta para encontrar uma passagem.
func NewFindBusTicketQuery(data FindBusTicketData) domain.Query[FindBusTicketData] {
	return findBusTicketQuery{data: data}
}
