package application

import (
	"github.com/mateusmacedo/go-bff/pkg/domain"
)

type FindBusTicketData struct {
	PassengerName string
}

type findBusTicketQuery struct {
	data FindBusTicketData
}

func (q findBusTicketQuery) QueryName() string {
	return "FindBusTicket"
}

func (q findBusTicketQuery) Payload() FindBusTicketData {
	return q.data
}

func NewFindBusTicketQuery(data FindBusTicketData) domain.Query[FindBusTicketData] {
	return findBusTicketQuery{data: data}
}
