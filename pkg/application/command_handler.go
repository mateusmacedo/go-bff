package application

import (
	"context"

	"github.com/mateusmacedo/go-bff/pkg/domain"
)

// CommandHandler define a interface para manipuladores de comando.
type CommandHandler[C domain.Command[T], T any] interface {
	Handle(ctx context.Context, command C) error
}
