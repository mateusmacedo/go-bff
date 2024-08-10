package application

import (
	"context"

	"github.com/mateusmacedo/go-bff/pkg/domain"
)

type EventHandler[E domain.Event[T], T any] interface {
	Handle(ctx context.Context, event E) error
}

type EventBus[E domain.Event[D], D any] interface {
	RegisterHandler(eventName string, handler EventHandler[E, D])
	Publish(ctx context.Context, event E) error
}
