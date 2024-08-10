package application

import (
	"github.com/mateusmacedo/go-bff/pkg/domain"
)

// FindPassageData contém os dados necessários para encontrar uma passagem.
type FindPassageData struct {
	PassageID string
}

// findPassageQuery é uma implementação privada de uma consulta para encontrar uma passagem.
type findPassageQuery struct {
	data FindPassageData
}

func (q findPassageQuery) QueryName() string {
	return "FindPassage"
}

func (q findPassageQuery) Payload() FindPassageData {
	return q.data
}

// NewFindPassageQuery cria uma nova consulta para encontrar uma passagem.
func NewFindPassageQuery(data FindPassageData) domain.Query[FindPassageData] {
	return findPassageQuery{data: data}
}
