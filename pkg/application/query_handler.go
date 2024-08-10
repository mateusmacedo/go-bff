package application

import (
	"context"

	"github.com/mateusmacedo/go-bff/pkg/domain"
)

// QueryHandler define a interface para manipuladores de consulta.
type QueryHandler[Q domain.Query[T], T any, R any] interface {
	Handle(ctx context.Context, query Q) (R, error)
}
