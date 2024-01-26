package orm

import (
	"context"
)

type Querier[T any] interface {
	// Get retrieves a T object from the database.
	// It takes a context as input and returns a pointer to T and an error.
	Get(ctx context.Context) (*T, error)
	GetMulti(ctx context.Context) ([]*T, error)
}

type Executor interface {
	Exec(ctx context.Context) Result
}

type Query struct {
	SQL  string
	Args []any
}

type QueryBuilder interface {
	Build() (*Query, error)
}
