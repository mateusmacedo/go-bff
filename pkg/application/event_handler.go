package application

import (
	"context"

	"github.com/mateusmacedo/go-bff/pkg/domain"
)

// EventHandler define a interface para manipuladores de eventos.
type EventHandler[E domain.Event[T], T any] interface {
	Handle(ctx context.Context, event E) error
}
