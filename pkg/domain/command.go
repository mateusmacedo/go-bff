package domain

// Command representa um comando no sistema.
type Command[T any] interface {
	CommandName() string
	Payload() T
}
