package orm

import "reflect"

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

// AssignNotNilColumns
//
//	@Description: 判断是否是 Nil， 如果是 nil 那么调用 assign 的时候，就不处理这个字段
//	@param entity
//	@return []Assignable
func AssignNotNilColumns(entity interface{}) []Assignable {
	return AssignColumns(entity, func(typ reflect.StructField, val reflect.Value) bool {
		switch val.Kind() {
		case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
			return !val.IsNil()
		}
		return true
	})
}

// AssignNotZeroColumns
//
//	@Description: 判断是否是零值， 如果是零值 那么调用 assign 的时候，就不处理这个字段
//	@param entity
//	@return []Assignable
func AssignNotZeroColumns(entity interface{}) []Assignable {
	return AssignColumns(entity, func(typ reflect.StructField, val reflect.Value) bool {
		return !val.IsZero()
	})
}

// AssignColumns returns a list of Assignable values for the given entity,
// filtered by the provided filter function.
func AssignColumns(entity interface{}, filter func(typ reflect.StructField, val reflect.Value) bool) []Assignable {
	// Get the value and type of the entity
	val := reflect.ValueOf(entity).Elem()
	typ := reflect.TypeOf(entity).Elem()

	// Get the number of fields in the entity
	numField := val.NumField()

	// Create a slice to store the Assignable values
	res := make([]Assignable, 0, numField)

	// Iterate over each field in the entity
	for i := 0; i < numField; i++ {
		// Get the value and type of the field
		fieldVal := val.Field(i)
		fieldTyp := typ.Field(i)

		// Check if the field passes the filter
		if filter(fieldTyp, fieldVal) {
			// Create an Assignable value and add it to the result slice
			res = append(res, Assign(fieldTyp.Name, fieldVal.Interface()))
		}
	}

	// Return the list of Assignable values
	return res
}

// 实现标记接口
func (a Assignment) assign() {}
