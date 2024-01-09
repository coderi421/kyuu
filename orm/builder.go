package orm

import (
	"github.com/coderi421/kyuu/orm/internal/errs"
	"github.com/coderi421/kyuu/orm/model"
	"strings"
)

type builder struct {
	sb    strings.Builder // sb is used to build the SQL query string.
	args  []any           // args holds the arguments for the query.
	model *model.Model    // model is the model associated with the selector.
}

// type Predicates []Predicate
//
// func (ps Predicates) build(s *strings.Builder) error {
// 	// 写在这里
// }

// type predicates struct {
// 	// WHERE 或者 HAVING
// 	prefix string
// 	ps []Predicate
// }

// func (ps predicates) build(s *strings.Builder) error {
//  包含拼接 WHERE 或者 HAVING 的部分
// 	// 写在这里
// }

// buildPredicates builds the predicates for the given list of predicates.
func (b *builder) buildPredicates(ps []Predicate) error {
	// Take the first predicate as the starting node.
	p := ps[0]

	// Iterate through the remaining predicates.
	for i := 1; i < len(ps); i++ {
		// Merge multiple predicates using the `And` method.
		p = p.And(ps[i])
	}

	// Recursively process the where statement.
	if err := b.buildExpression(p); err != nil {
		return err
	}
	return nil
}

// buildExpression builds the SQL query for the given expression.
// It takes an expression as input and recursively constructs the SQL query.
// The SQL query is stored in the builder's string buffer (b.sb).
// The argument values are stored in the builder's argument list (b.args).
func (b *builder) buildExpression(e Expression) error {
	// Column 代表是列名，直接拼接列名
	// value 代表参数，加入参数列表
	// Predicate 代表一个查询条件：
	// 如果左边是一个 Predicate，那么加上括号
	// 递归构造左边
	// 构造操作符
	// 如果右边是一个 Predicate，那么加上括号
	if e == nil {
		return nil
	}

	switch expr := e.(type) {
	case Column:
		// Append column name to the SQL query
		fd, ok := b.model.FieldMap[expr.name]
		if !ok {
			return errs.NewErrUnknownField(expr.name)
		}
		b.sb.WriteByte('`')
		b.sb.WriteString(fd.ColName)
		b.sb.WriteByte('`')
	case value:
		// Append placeholder to the SQL query and add value to the argument list
		b.sb.WriteByte('?')
		//b.args = append(b.args, expr.val)
		b.addArgs(expr.val)
	case RawExpr:
		// 执行原生 sql 语句
		b.sb.WriteString(expr.raw)
		if len(expr.args) != 0 {
			b.addArgs(expr.args...)
			//if b.args == nil {
			//	b.args = make([]any, 0, 8)
			//}
			//b.args = append(b.args, expr.args...)
		}
	case Predicate:
		// Build left expression
		// 如果左边有复杂结构，则在最外边套一层括号
		_, lp := expr.left.(Predicate)
		if lp {
			b.sb.WriteByte('(')
		}
		if err := b.buildExpression(expr.left); err != nil {
			return err
		}
		if lp {
			b.sb.WriteByte(')')
		}

		if expr.op == "" {
			// 如果只有左边（op 符号为空，就不需要连接），例如执行原生 sql raw 的时候，就只有左边
			return nil
		}

		//处理运算符号
		// Append operator to the SQL query
		b.sb.WriteByte(' ')
		b.sb.WriteString(expr.op.String())
		b.sb.WriteByte(' ')

		// 处理右边的逻辑
		// Build right expression
		_, rp := expr.right.(Predicate)
		if rp {
			b.sb.WriteByte('(')
		}
		if err := b.buildExpression(expr.right); err != nil {
			return err
		}
		if rp {
			b.sb.WriteByte(')')
		}
	default:
		return errs.NewErrUnsupportedExpressionType(expr)
	}

	return nil
}

func (b *builder) addArgs(args ...any) {
	if b.args == nil {
		b.args = make([]any, 0, 8)
	}
	b.args = append(b.args, args...)
}
