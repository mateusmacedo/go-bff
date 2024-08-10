package domain

type Query[T any] interface {
	QueryName() string
	Payload() T
}
