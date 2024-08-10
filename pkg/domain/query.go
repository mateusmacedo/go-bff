package domain

// Query representa uma consulta no sistema.
type Query[T any] interface {
	QueryName() string
	Payload() T
}
