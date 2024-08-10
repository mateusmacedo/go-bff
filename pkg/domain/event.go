package domain

// Event representa um evento no sistema.
type Event[T any] interface {
	EventName() string
	Payload() T
}
