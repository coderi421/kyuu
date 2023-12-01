package model

import (
	"github.com/coderi421/kyuu/orm/internal/errs"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

type Registry interface {
	Get(val any) (*Model, error)
	Register(val any, opts ...Option) (*Model, error)
}

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

func NewRegistry() Registry {
	return &registry{}
}

// Get fetches the model associated with a given value.
// If the model is not found in the registry, it is parsed and stored for future use.
// Get 查找元数据模型
func (r *registry) Get(val any) (*Model, error) {
	// Get the type of the value
	typ := reflect.TypeOf(val)

	// Check if the model is already present in the registry
	m, ok := r.models.Load(typ)
	if ok {
		return m.(*Model), nil
	}

	// Return the model
	return r.Register(val)
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

// Register registers a model in the registry with the given options.
// It parses the model if it is not found and applies the provided options.
// It stores the model in the registry and returns the registered model.
func (r *registry) Register(val any, opts ...Option) (*Model, error) {
	// If the model is not found, parse it
	m, err := r.parseModel(val)
	if err != nil {
		return nil, err
	}

	// Apply the provided options to the model
	for _, opt := range opts {
		err = opt(m)
		if err != nil {
			return nil, err
		}
	}

	typ := reflect.TypeOf(val)

	// Store the model in the registry
	r.models.Store(typ, m)

	return m, nil
}

// parseModel parses a given reflect.Type and returns a new model or an error.
// It checks if the type is a pointer to a struct and generates a map of Field names
// and their corresponding column names for the model.
// orm:"key1=value1,key2=value2"
func (r *registry) parseModel(val any) (*Model, error) {
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

	// Create a map to store the Struct Field names and their corresponding column names
	fds := make(map[string]*Field, numField)
	// Create a map to store the DB names and their corresponding column names
	colMap := make(map[string]*Field, numField)

	// Iterate over each Field in the struct
	for i := 0; i < numField; i++ {
		// Get the reflect.Struct Field of the current Field
		fdStruct := typ.Field(i)

		// Process the tag of the Field
		tags, err := r.parseTag(fdStruct.Tag)
		if err != nil {
			return nil, err
		}

		// Get the column name from the tag or use the default Field name
		colName := tags[tagKeyColumn]
		if colName == "" {
			// If the colName is "", user the default  ItemId -> item_id
			colName = underscoreName(fdStruct.Name)
		}

		f := &Field{
			ColName: colName,
			GoName:  fdStruct.Name,
			Type:    fdStruct.Type,
			Offset:  fdStruct.Offset,
		}
		// Store the Struct Field's column name in the map
		fds[fdStruct.Name] = f
		// Store the DB's column name in the map
		colMap[colName] = f
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
	return &Model{
		TableName: tableName,
		FieldMap:  fds,
		ColumnMap: colMap,
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

// WithTableName is a Option function that sets the table name for a Model.
func WithTableName(tableName string) Option {
	return func(model *Model) error {
		model.TableName = tableName
		return nil
	}
}

// ModelWithColumnName is a function that returns a Option function, which can be used to set the column name for a specific Field in a model.
func WithColumnName(field, columnName string) Option {
	return func(model *Model) error {
		// Check if the Field exists in the model's Field map
		fd, ok := model.FieldMap[field]
		if !ok {
			// Return an error if the Field is unknown
			return errs.NewErrUnknownField(field)
		}

		// Set the column name for the Field
		fd.ColName = columnName
		return nil
	}
}
