package domain

type Command[T any] interface {
	CommandName() string
	Payload() T
}
