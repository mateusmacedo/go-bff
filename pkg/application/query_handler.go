package application

import (
	"context"

	"github.com/mateusmacedo/go-bff/pkg/domain"
)

type QueryHandler[Q domain.Query[T], T any, R any] interface {
	Handle(ctx context.Context, query Q) (R, error)
}

type QueryBus[Q domain.Query[D], D any, R any] interface {
	RegisterHandler(queryName string, handler QueryHandler[Q, D, R])
	Dispatch(ctx context.Context, query Q) (R, error)
}
