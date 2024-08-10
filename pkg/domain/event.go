package domain

type Event[T any] interface {
	EventName() string
	Payload() T
}
