package orm

// 只拼接 where 中的 一组条件

type Column struct {
	table TableReference // 当使用 Join 进行关联的时候，会使用与实例化 struct 不同的 结构体，联表的时候，字段可能是不同表的字段，所以 column 级别也需要维护一个 table 信息
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

// 获取列别名
func (c Column) selectedAlias() string {
	return c.alias
}

// 获取字段名称
func (c Column) fieldName() string {
	return c.name
}

// 根据字段找到对应表
func (c Column) target() TableReference {
	return c.table
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

// In 有两种输入，一种是 IN 子查询
// 另外一种就是普通的值
// 这里我们可以定义两个方法，如 In  和 InQuery，也可以定义一个方法
// 这里我们使用一个方法
func (c Column) In(vals ...any) Predicate {
	return Predicate{
		left:  c,
		op:    opIN,
		right: valueOf(vals),
	}
}

func (c Column) InQuery(sub Subquery) Predicate {
	return Predicate{
		left:  c,
		op:    opIN,
		right: sub,
	}
}
