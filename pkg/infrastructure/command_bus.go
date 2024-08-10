package infrastructure

import (
	"context"
	"errors"
	"sync"

	"github.com/mateusmacedo/go-bff/pkg/application"
	"github.com/mateusmacedo/go-bff/pkg/domain"
)

type simpleCommandBus[C domain.Command[D], D any] struct {
	handlers map[string]application.CommandHandler[C, D]
	mu       sync.RWMutex
	logger   application.AppLogger
}

func NewSimpleCommandBus[C domain.Command[D], D any](logger application.AppLogger) application.CommandBus[C, D] {
	return &simpleCommandBus[C, D]{
		handlers: make(map[string]application.CommandHandler[C, D]),
		logger:   logger,
	}
}

func (bus *simpleCommandBus[C, D]) RegisterHandler(commandName string, handler application.CommandHandler[C, D]) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.handlers[commandName] = handler
}

func (bus *simpleCommandBus[C, D]) Dispatch(ctx context.Context, command C) error {
	bus.mu.RLock()
	handler, found := bus.handlers[command.CommandName()]
	bus.mu.RUnlock()

	if !found {
		bus.logger.Error(ctx, "no handler registered for command", map[string]interface{}{
			"command_name": command.CommandName(),
		})
		return errors.New("no handler registered for command")
	}

	bus.logger.Info(ctx, "dispatching command", map[string]interface{}{
		"command_name": command.CommandName(),
	})
	return handler.Handle(ctx, command)
}
