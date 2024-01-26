package orm

// RawExpr 代表一个原生表达式
// 意味着 ORM 不会对它进行任何处理
type RawExpr struct {
	raw  string
	args []any
}

func (r RawExpr) selectable() {}

func (r RawExpr) expr() {}

func (r RawExpr) AsPredicate() Predicate {
	return Predicate{
		left: r,
	}
}

// Raw 创建一个 RawExpr
// 执行原生sql 语句
func Raw(expr string, args ...any) RawExpr {
	return RawExpr{
		raw:  expr,
		args: args,
	}
}

type binaryExpr struct {
	left  Expression
	op    op
	right Expression
}

// 实现功能接口 Expression
func (b binaryExpr) expr() {}

// MathExpr 为非导出 struct 创建类型
// update 过程中的计算方法
type MathExpr binaryExpr

func (m MathExpr) Add(val any) MathExpr {
	return MathExpr{
		left:  m,
		op:    opAdd,
		right: valueOf(val),
	}
}

func (m MathExpr) Multi(val any) MathExpr {
	return MathExpr{
		left:  m,
		op:    opMulti,
		right: valueOf(val),
	}
}

func (m MathExpr) expr() {}
