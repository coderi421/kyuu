package orm

import (
	"github.com/coderi421/kyuu/orm/internal/errs"
	"strings"
)

type builder struct {
	// 核心的共通部分
	core
	sb   strings.Builder // sb is used to build the SQL query string.
	args []any           // args holds the arguments for the query.

	quoter  byte    // 不同数据库的标点不同，mysql `id_my` postgresql 'id_my'
	dialect Dialect // db 初始化的时候，确定的方言  mysql postgresql
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
		// 这种写法很隐晦
		//expr.alias = ""
		// Append column name to the SQL query
		// WHERE 中的条件 不允许用别名
		return b.buildColumn(expr.table, expr.name)
	case Aggregate:
		//ex. HAVING AVG(`age`) = ?;
		return b.buildAggregate(expr, false)
	case value:
		// Append placeholder to the SQL query and add value to the argument list
		b.sb.WriteByte('?')
		//b.args = append(b.args, expr.val)
		b.addArgs(expr.val)
	case RawExpr:
		// 执行原生 sql 语句
		//b.sb.WriteString(expr.raw)
		//if len(expr.args) != 0 {
		//	b.addArgs(expr.args...)
		//	//if b.args == nil {
		//	//	b.args = make([]any, 0, 8)
		//	//}
		//	//b.args = append(b.args, expr.args...)
		//}
		b.raw(expr)
	case MathExpr:
		return b.buildBinaryExpr(binaryExpr(expr))
	case Predicate:
		return b.buildBinaryExpr(binaryExpr(expr))
		//// Build left expression
		//// 如果左边有复杂结构，则在最外边套一层括号
		//_, lp := expr.left.(Predicate)
		//if lp {
		//	b.sb.WriteByte('(')
		//}
		//if err := b.buildExpression(expr.left); err != nil {
		//	return err
		//}
		//if lp {
		//	b.sb.WriteByte(')')
		//}
		//
		//if expr.op == "" {
		//	// 如果只有左边（op 符号为空，就不需要连接），例如执行原生 sql raw 的时候，就只有左边
		//	return nil
		//}
		//
		////处理运算符号
		//// Append operator to the SQL query
		//b.sb.WriteByte(' ')
		//b.sb.WriteString(expr.op.String())
		//b.sb.WriteByte(' ')
		//
		//// 处理右边的逻辑
		//// Build right expression
		//_, rp := expr.right.(Predicate)
		//if rp {
		//	b.sb.WriteByte('(')
		//}
		//if err := b.buildExpression(expr.right); err != nil {
		//	return err
		//}
		//if rp {
		//	b.sb.WriteByte(')')
		//}
	case SubqueryExpr:
		b.sb.WriteString(expr.pred)
		b.sb.WriteByte(' ')
		return b.buildSubquery(expr.s, false)
	case Subquery:
		return b.buildSubquery(expr, false)
	case binaryExpr:
		return b.buildBinaryExpr(expr)
	default:
		return errs.NewErrUnsupportedExpressionType(expr)
	}

	return nil
}

func (b *builder) buildColumn(table TableReference, fd string) error {
	// Join 的时候可能使用的 struct 不是 初始化 builder 时候的 struct
	// 找到表中对应名字
	//meta, ok := b.model.FieldMap[fd]
	//if !ok {
	//	return errs.NewErrUnknownField(fd)
	//}
	var alias string
	if table != nil {
		alias = table.tableAlias()
	}
	if alias != "" {
		b.quote(alias)
		b.sb.WriteByte('.')
	}
	colName, err := b.colName(table, fd)
	if err != nil {
		return err
	}
	b.quote(colName)
	// from 后的部分不需要 使用 as 别名
	return nil
}

func (b *builder) colName(table TableReference, fd string) (string, error) {
	switch tab := table.(type) {
	case nil:
		fdMeta, ok := b.model.FieldMap[fd]
		if !ok {
			return "", errs.NewErrUnknownField(fd)
		}
		return fdMeta.ColName, nil
	case Table:
		// 这里处理的是 join 中其他表的字段名称
		/*
			t1 := TableOf(&Order{}).As("t1")
							t2 := TableOf(&OrderDetail{}).As("t2")
							t3 := t1.Join(t2).On(t1.C("Id").EQ(t2.C("OrderId")))
							t4 := TableOf(&Item{}).As("t4")
							t5 := t3.Join(t4).On(t2.C("ItemId").EQ(t4.C("Id")))
							return NewSelector[Order](db).From(t5)
		*/
		m, err := b.r.Get(tab.entity)
		if err != nil {
			return "", err
		}
		fdMeta, ok := m.FieldMap[fd]
		if !ok {
			return "", errs.NewErrUnknownField(fd)
		}
		return fdMeta.ColName, nil
	default:
		return "", errs.NewErrUnsupportedExpressionType(tab)
	}
}

func (b *builder) buildAggregate(a Aggregate, useAlias bool) error {
	// 找到表中对应名字
	// 这里使用 ORM 的时候，默认使用 struct 的名字作为 column 检索字段
	// for example: Id, Age 然后到内存中储存的 map 中找对应的表中的字段名称
	//fd, ok := b.model.FieldMap[a.arg]
	//if !ok {
	//	return errs.NewErrUnknownField(a.arg)
	//}

	b.sb.WriteString(a.fn)
	b.sb.WriteByte('(')
	err := b.buildColumn(a.table, a.arg)
	if err != nil {
		return err
	}
	b.sb.WriteByte(')')
	if useAlias {
		b.buildAs(a.alias)
	}
	return nil
}

func (b *builder) buildSubquery(tab Subquery, useAlias bool) error {
	q, err := tab.s.Build()
	if err != nil {
		return err
	}
	b.sb.WriteByte('(')
	b.sb.WriteString(q.SQL[:len(q.SQL)-1])
	if len(q.Args) > 0 {
		b.addArgs(q.Args...)
	}
	b.sb.WriteByte(')')
	if useAlias {
		b.sb.WriteString(" AS ")
		b.quote(tab.alias)
	}
	return nil
}

func (b *builder) buildBinaryExpr(e binaryExpr) error {
	err := b.buildSubExpr(e.left)
	if err != nil {
		return err
	}
	if e.op != "" {
		b.sb.WriteByte(' ')
		b.sb.WriteString(e.op.String())
	}
	if e.right != nil {
		b.sb.WriteByte(' ')
		return b.buildSubExpr(e.right)
	}
	return nil
}

// 处理运算逻辑
func (b *builder) buildSubExpr(subExpr Expression) error {
	switch sub := subExpr.(type) {
	case MathExpr:
		_ = b.sb.WriteByte('(')
		if err := b.buildBinaryExpr(binaryExpr(sub)); err != nil {
			return err
		}
		_ = b.sb.WriteByte(')')
	case binaryExpr:
		_ = b.sb.WriteByte('(')
		if err := b.buildBinaryExpr(sub); err != nil {
			return err
		}
		_ = b.sb.WriteByte(')')
	case Predicate:
		_ = b.sb.WriteByte('(')
		if err := b.buildBinaryExpr(binaryExpr(sub)); err != nil {
			return err
		}
		_ = b.sb.WriteByte(')')
	default:
		if err := b.buildExpression(sub); err != nil {
			return err
		}
	}
	return nil
}

func (b *builder) addArgs(args ...any) {
	if b.args == nil {
		b.args = make([]any, 0, 8)
	}
	b.args = append(b.args, args...)
}

func (b *builder) buildAs(alias string) {
	if alias != "" {
		b.sb.WriteString(" AS ")
		b.quote(alias)
	}
}

func (b *builder) quote(name string) {
	b.sb.WriteByte(b.quoter)
	b.sb.WriteString(name)
	b.sb.WriteByte(b.quoter)
}

func (b *builder) raw(r RawExpr) {
	b.sb.WriteString(r.raw)
	if len(r.args) != 0 {
		b.addArgs(r.args...)
	}
}
