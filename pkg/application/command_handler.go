package application

import (
	"context"

	"github.com/mateusmacedo/go-bff/pkg/domain"
)

type CommandHandler[C domain.Command[T], T any] interface {
	Handle(ctx context.Context, command C) error
}

type CommandBus[C domain.Command[T], T any] interface {
	RegisterHandler(commandName string, handler CommandHandler[C, T])
	Dispatch(ctx context.Context, command C) error
}
