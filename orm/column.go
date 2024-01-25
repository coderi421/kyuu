package orm

// 只拼接 where 中的 一组条件

type Column struct {
	name  string
	alias string // as 别名
}

// 处理插入操作指定字段的接口
func (c Column) assign() {}

func (c Column) expr()       {}
func (c Column) selectable() {}

func (c Column) As(alias string) Column {
	return Column{
		name:  c.name,
		alias: alias,
	}
}

type value struct {
	val any
}

func (v value) expr() {}

// valueOf creates a new value object with the given value.
// It takes in a generic value and returns a value object.
func valueOf(val any) value {
	return value{val: val}
}

func C(name string) Column {
	return Column{name: name}
}

// Add ...Set(Assign("Age", C("Age").Add(1))), -> SET `age`=`age` + ?;
func (c Column) Add(delta int) MathExpr {
	return MathExpr{
		left:  c,
		op:    opAdd,
		right: value{val: delta},
	}
}

// Multi ...Set(Assign("Age", C("Age").Multi(1))), -> SET `age`=`age` * ?;
func (c Column) Multi(delta int) MathExpr {
	return MathExpr{
		left:  c,
		op:    opMulti,
		right: value{val: delta},
	}
}

// EQ 例如 C("id").Eq(12)
func (c Column) EQ(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opEQ,
		right: exprOf(arg), // 如果 arg 不是 Expression 类型 就让他变成这个类型
	}
}

// LT 例如 C("id").LT(12)
func (c Column) LT(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opLT,
		right: exprOf(arg), // 如果 arg 不是 Expression 类型 就让他变成这个类型
	}
}

func (c Column) GT(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opGT,
		right: exprOf(arg), // 如果 arg 不是 Expression 类型 就让他变成这个类型
	}
}
