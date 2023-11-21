package orm

import (
	"github.com/coderi421/kyuu/orm/internal/errs"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

// 这种包变量对测试不友好，缺乏隔离
//
//	var defaultRegistry = &registry{
//		models: make(map[reflect.Type]*model, 16),
//	}
type registry struct {
	// models key 是类型名
	// 这种定义方式是不行的
	// 1. 类型名冲突，例如都是 User，但是一个映射过去 buyer_t
	// 一个映射过去 seller_t
	// 2. 并发不安全
	// models map[string]*model

	// lock sync.RWMutex
	// models map[reflect.Type]*model
	// reflect.Type 可以解决命名冲突的问题
	models sync.Map
}

// get fetches the model associated with a given value.
// If the model is not found in the registry, it is parsed and stored for future use.
func (r *registry) get(val any) (*model, error) {
	// Get the type of the value
	typ := reflect.TypeOf(val)

	// Check if the model is already present in the registry
	m, ok := r.models.Load(typ)
	if !ok {
		// If the model is not found, parse it
		var err error
		if m, err = r.parseModel(val); err != nil {
			return nil, err
		}
	}

	// Store the model in the registry
	r.models.Store(typ, m)

	// Return the model
	return m.(*model), nil
}

//var models = map[reflect.Type]*model{}
// 直接 map
// func (r *registry) get(val any) (*model, error) {
// 	typ := reflect.TypeOf(val)
// 	m, ok := r.models[typ]
// 	if !ok {
// 		var err error
// 		if m, err = r.parseModel(typ); err != nil {
// 			return nil, err
// 		}
// 	}
// 	r.models[typ] = m
// 	return m, nil
// }

// 使用读写锁的并发安全解决思路
// func (r *registry) get1(val any) (*model, error) {
// 	r.lock.RLock()
// 	typ := reflect.TypeOf(val)
// 	m, ok := r.models[typ]
// 	r.lock.RUnlock()
// 	if ok {
// 		return m, nil
// 	}
// 	r.lock.Lock()
// 	defer r.lock.Unlock()
// 	m, ok = r.models[typ]
// 	if ok {
// 		return m, nil
// 	}
// 	var err error
// 	if m, err = r.parseModel(typ); err != nil {
// 		return nil, err
// 	}
// 	r.models[typ] = m
// 	return m, nil
// }

// parseModel parses a given reflect.Type and returns a new model or an error.
// It checks if the type is a pointer to a struct and generates a map of field names
// and their corresponding column names for the model.
// orm:"key1=value1,key2=value2"
func (r *registry) parseModel(val any) (*model, error) {
	// Get the type of the input value
	typ := reflect.TypeOf(val)

	// Check if the type is a pointer to a struct
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		// Only support one-level pointer as input, e.g. *User does not support **User and User
		return nil, errs.ErrPointerOnly
	}

	// Dereference the pointer to get the struct type
	typ = typ.Elem()

	// Get the number of fields in the struct
	numField := typ.NumField()

	// Create a map to store the field names and their corresponding column names
	fds := make(map[string]*field, numField)

	// Iterate over each field in the struct
	for i := 0; i < numField; i++ {
		// Get the reflect.StructField of the current field
		fdStruct := typ.Field(i)

		// Process the tag of the field
		tags, err := r.parseTag(fdStruct.Tag)
		if err != nil {
			return nil, err
		}

		// Get the column name from the tag or use the default field name
		colName := tags[tagKeyColumn]
		if colName == "" {
			colName = underscoreName(fdStruct.Name)
		}

		// Store the field's column name in the map
		fds[fdStruct.Name] = &field{
			colName: colName,
		}
	}

	// Get the table name from the input value if it implements TableName interface
	var tableName string
	if tn, ok := val.(TableName); ok {
		tableName = tn.TableName()
	}
	// If the table name is not provided, generate it from the struct name
	if tableName == "" {
		tableName = underscoreName(typ.Name())
	}

	// Create and return the model
	return &model{
		tableName: tableName,
		fieldMap:  fds,
	}, nil
}

// parseTag parses the given struct tag and returns a map of key-value pairs.
// If the tag is empty, it returns an empty map and no error.
// If the tag contains an invalid key-value pair, it returns an error.
func (r *registry) parseTag(tag reflect.StructTag) (map[string]string, error) {
	ormTag := tag.Get(tagORMName)
	if ormTag == "" {
		// Return an empty map so that the caller doesn't need to check for nil
		return map[string]string{}, nil
	}

	// Initialize the result map with a capacity of 1, as we support only one key
	res := make(map[string]string, 1)

	// Split the tag string into individual key-value pairs
	pairs := strings.Split(ormTag, ",")
	for _, pair := range pairs {
		// Split each pair into key and value
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			return nil, errs.NewErrInvalidTagContent(pair)
		}
		// Add the key-value pair to the result map
		res[kv[0]] = kv[1]
	}

	return res, nil
}

// underscoreName converts a given table name to underscore case.
// It replaces any uppercase letter with an underscore followed by the lowercase letter.
// It returns the converted table name as a string.
// UserName -> user_name
func underscoreName(tableName string) string {
	var buf []byte
	for i, v := range tableName {
		// If the character is uppercase
		if unicode.IsUpper(v) {
			// Add an underscore before the lowercase letter
			if i != 0 {
				buf = append(buf, '_')
			}
			buf = append(buf, byte(unicode.ToLower(v)))
		} else {
			// Append the character as it is
			buf = append(buf, byte(v))
		}
	}
	return string(buf)
}
