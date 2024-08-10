package infrastructure

import (
	"context"
	"errors"
	"sync"

	"github.com/mateusmacedo/go-bff/pkg/application"
	"github.com/mateusmacedo/go-bff/pkg/domain"
)

// simpleCommandBus é uma implementação simples de um barramento de comandos.
type simpleCommandBus[C domain.Command[T], T any] struct {
	handlers map[string]application.CommandHandler[C, T]
	mu       sync.RWMutex
}

// NewSimpleCommandBus cria uma nova instância de SimpleCommandBus.
func NewSimpleCommandBus[C domain.Command[T], T any]() *simpleCommandBus[C, T] {
	return &simpleCommandBus[C, T]{
		handlers: make(map[string]application.CommandHandler[C, T]),
	}
}

// RegisterHandler registra um manipulador para um comando específico.
func (bus *simpleCommandBus[C, T]) RegisterHandler(commandName string, handler application.CommandHandler[C, T]) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.handlers[commandName] = handler
}

// Dispatch despacha um comando para o manipulador registrado usando goroutines.
func (bus *simpleCommandBus[C, T]) Dispatch(ctx context.Context, command C) error {
	bus.mu.RLock()
	handler, found := bus.handlers[command.CommandName()]
	bus.mu.RUnlock()

	if !found {
		return errors.New("no handler registered for command")
	}

	done := make(chan error, 1)

	go func() {
		done <- handler.Handle(ctx, command)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}
