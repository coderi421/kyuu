package orm

// Aggregate 代表聚合函数， 例如 AVG, MAX, MIN 等 以及别名
type Aggregate struct {
	table TableReference
	fn    string
	arg   string
	alias string
}

func (a Aggregate) selectedAlias() string {
	return a.alias
}

func (a Aggregate) fieldName() string {
	return a.arg
}

func (a Aggregate) target() TableReference {
	return a.table
}

func (a Aggregate) expr() {}

// As 这里使用 值 作为接收者，可以防止并发问题，每次都返回一个新的；也有小利于垃圾回收，局部之后，变量就会被回收
func (a Aggregate) As(alias string) Aggregate {
	return Aggregate{
		fn:    a.fn,
		arg:   a.arg,
		alias: alias,
	}
}

// EQ 例如 AVG("id").EQ(12)
func (a Aggregate) EQ(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opEQ,
		right: exprOf(arg),
	}
}

func (a Aggregate) LT(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opLT,
		right: exprOf(arg),
	}
}

func (a Aggregate) GT(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opGT,
		right: exprOf(arg),
	}
}

// Avg
//
//	@Description: 求平均值
//	@param c column，聚合函数中填写的字段
//	@return Aggregate
func Avg(c string) Aggregate {
	return Aggregate{
		fn:  "AVG",
		arg: c,
	}
}

// Max
//
//	@Description: 求最大值
//	@param c column，聚合函数中填写的字段
//	@return Aggregate
func Max(c string) Aggregate {
	return Aggregate{
		fn:  "MAX",
		arg: c,
	}
}

// Min
//
//	@Description: 求聚合函数中填写的最小值
//	@param c column，聚合函数中填写的字段
//	@return Aggregate
func Min(c string) Aggregate {
	return Aggregate{
		fn:  "MIN",
		arg: c,
	}
}

// Count
//
//	@Description: 获取数量
//	@param c column，聚合函数中填写的字段
//	@return Aggregate
func Count(c string) Aggregate {
	return Aggregate{
		fn:  "COUNT",
		arg: c,
	}
}

// Sum
//
//	@Description: 求和
//	@param c
//	@return Aggregate
func Sum(c string) Aggregate {
	return Aggregate{
		fn:  "SUM",
		arg: c,
	}
}
