package orm

import (
	"reflect"
	"strings"
)

type Selector[T any] struct {
	table string
}

// NewSelector creates a new instance of Selector.
func NewSelector[T any]() *Selector[T] {
	return &Selector[T]{}
}

// From sets the table name for the selector.
// It returns the updated selector.
func (s *Selector[T]) From(tbl string) *Selector[T] {
	s.table = tbl
	return s
}

// Build generates a SQL query for selecting all columns from a table.
// It returns the generated query as a *Query struct or an error if there was any.
func (s *Selector[T]) Build() (*Query, error) {
	var sb strings.Builder
	sb.WriteString("SELECT * FROM ")

	if s.table == "" {
		var t T
		sb.WriteByte('`')
		// Get the name of the struct using reflection
		sb.WriteString(reflect.TypeOf(t).Name())
		sb.WriteByte('`')
	} else {
		// 这里没有处理 添加`符号，让用户自己应该名字自己在做什么
		sb.WriteString(s.table)
	}

	sb.WriteString(";")

	return &Query{
		SQL: sb.String(),
	}, nil
}
