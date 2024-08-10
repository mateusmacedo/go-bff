package application

import (
	"context"

	"github.com/mateusmacedo/go-bff/pkg/domain"
)

// CommandHandler define a interface para manipuladores de comando.
type CommandHandler[C domain.Command[T], T any] interface {
	Handle(ctx context.Context, command C) error
}

// CommandBus define a interface para o barramento de comandos.
type CommandBus[C domain.Command[T], T any] interface {
	RegisterHandler(commandName string, handler CommandHandler[C, T]) // Registra um manipulador de comando
	Dispatch(ctx context.Context, command C) error                    // Despacha um comando
}
