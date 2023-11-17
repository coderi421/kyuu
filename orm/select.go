package orm

import (
	"github.com/coderi421/kyuu/orm/internal/errs"
	"strings"
)

// Selector represents a query selector that allows building SQL SELECT statements.
// It holds the necessary information to construct the query.
type Selector[T any] struct {
	sb    strings.Builder // sb is used to build the SQL query string.
	args  []any           // args holds the arguments for the query.
	table string          // table is the name of the table to select from.
	where []Predicate     // where holds the WHERE predicates for the query.
	model *model          // model is the model associated with the selector.

	db *DB // db is the DB instance used for executing the query.
}

// NewSelector creates a new instance of Selector.
func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		db: db,
	}
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
	var (
		t   T
		err error
	)

	s.model, err = s.db.r.get(&t)
	if err != nil {
		return nil, err
	}

	s.sb.WriteString("SELECT * FROM ")

	if s.table == "" {
		s.sb.WriteByte('`')
		// Get the name of the struct using reflection
		s.sb.WriteString(s.model.tableName)
		s.sb.WriteByte('`')
	} else {
		// 这里没有处理 添加`符号，让用户自己应该名字自己在做什么
		s.sb.WriteString(s.table)
	}

	// construct where
	if len(s.where) > 0 {
		// 类似这种可有可无的部分，都要在前面加一个空格
		s.sb.WriteString(" WHERE ")
		// 取出第一个作为开始的节点
		p := s.where[0]
		for i := 1; i < len(s.where); i++ {
			// 合并多个 predicate
			p = p.And(s.where[i])
		}

		// 递归处理 where 语句
		if err = s.buildExpression(p); err != nil {
			return nil, err
		}
	}

	s.sb.WriteString(";")

	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}

// Column 代表是列名，直接拼接列名
// value 代表参数，加入参数列表
// Predicate 代表一个查询条件：
// 如果左边是一个 Predicate，那么加上括号
// 递归构造左边
// 构造操作符
// 如果右边是一个 Predicate，那么加上括号
func (s *Selector[T]) buildExpression(e Expression) error {
	if e == nil {
		return nil
	}

	switch expr := e.(type) {
	case Column:
		fd, ok := s.model.fieldMap[expr.name]
		if !ok {
			return errs.NewErrUnknownField(expr.name)
		}
		s.sb.WriteByte('`')
		s.sb.WriteString(fd.colName)
		s.sb.WriteByte('`')
	case value:
		s.sb.WriteByte('?')
		s.args = append(s.args, expr.val)
	case Predicate:
		// 如果左边有复杂结构，则在最外边套一层括号
		_, lp := expr.left.(Predicate)
		if lp {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(expr.left); err != nil {
			return err
		}
		if lp {
			s.sb.WriteByte(')')
		}

		//处理运算符号
		s.sb.WriteByte(' ')
		s.sb.WriteString(expr.op.String())
		s.sb.WriteByte(' ')

		// 处理右边的逻辑
		_, rp := expr.right.(Predicate)
		if rp {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(expr.right); err != nil {
			return err
		}
		if rp {
			s.sb.WriteByte(')')
		}
	default:
		return errs.NewErrUnsupportedExpressionType(expr)
	}

	return nil
}

// Where 用于构造 WHERE 查询条件。如果 ps 长度为 0，那么不会构造 WHERE 部分
func (s *Selector[T]) Where(ps ...Predicate) *Selector[T] {
	// Set the WHERE conditions
	s.where = ps
	return s
}

// cols 是用于 WHERE 的列，难以解决 And Or 和 Not 等问题
// func (s *Selector[T]) Where(cols []string, args...any) *Selector[T] {
// 	s.whereCols = cols
// 	s.args = append(s.args, args...)
// }

// 最为灵活的设计
// func (s *Selector[T]) Where(where string, args...any) *Selector[T] {
// 	s.where = where
// 	s.args = append(s.args, args...)
// }
