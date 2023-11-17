package reflect

import (
	"errors"
	"reflect"
)

// iterateFields 返回所有的字段名字
// input 只能是结构体，或者结构体指针，可以是多重指针
func interateFileds(input any) (map[string]any, error) {
	typ := reflect.TypeOf(input)
	val := reflect.ValueOf(input)

	// 处理指针，要拿到指针指向的东西
	// 这里我们综合考虑了多重指针的效果
	// 使用 for 是因为可能是多重指针
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}

	// 如果不是结构体，就返回 error
	if typ.Kind() != reflect.Struct {
		return nil, errors.New("非法类型")
	}

	num := typ.NumField()
	res := make(map[string]any, num)
	for i := 0; i < num; i++ {
		fd := typ.Field(i)
		fdVal := val.Field(i)
		if fd.IsExported() {
			res[fd.Name] = fdVal.Interface()
		} else {
			// 为了演示效果，不公开字段我们用零值来填充
			res[fd.Name] = reflect.Zero(fd.Type).Interface()
		}
	}

	return res, nil
}
