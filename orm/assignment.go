package orm

// Assignable 标记接口，
// 实现该接口意味着可以用于赋值语句，
// 用于在 UPDATE 和 UPSERT 中 Assign("FirstName", "DaMing") -> SET`first_name`=?
type Assignable interface {
	assign()
}

type Assignment struct {
	column string
	val    Expression
}

func Assign(column string, val any) Assignment {
	v, ok := val.(Expression)
	if !ok {
		v = value{val: val}
	}
	return Assignment{
		column: column,
		val:    v,
	}
}

// 实现标记接口
func (a Assignment) assign() {}
