package application

import (
	"context"

	"github.com/mateusmacedo/go-bff/pkg/domain"
)

// QueryHandler define a interface para manipuladores de consulta.
type QueryHandler[Q domain.Query[T], T any, R any] interface {
	Handle(ctx context.Context, query Q) (R, error)
}

// QueryBus define a interface para o barramento de consultas.
type QueryBus[Q domain.Query[D], D any, R any] interface {
	RegisterHandler(queryName string, handler QueryHandler[Q, D, R]) // Registra um manipulador de consulta
	Dispatch(ctx context.Context, query Q) (R, error)                // Despacha uma consulta
}
