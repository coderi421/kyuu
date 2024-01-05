package orm

import (
	"database/sql"
	"github.com/coderi421/kyuu/orm/internal/valuer"
	"github.com/coderi421/kyuu/orm/internal/valuer/reflect"
	"github.com/coderi421/kyuu/orm/internal/valuer/unsafe"
	"github.com/coderi421/kyuu/orm/model"
)

type DBOption func(*DB)

type DB struct {
	r          model.Registry // 存储数据库表和 struct 映射关系的实例
	db         *sql.DB
	valCreator valuer.Creator // 与DB交互映射的实现
}

// Open opens a database connection using the specified driver and DSN.
// It also accepts optional DBOptions.
// 可以传入不同的 driver 连接不同的 db
func Open(driver string, dsn string, opts ...DBOption) (*DB, error) {
	// Open the database connection
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	// Return the opened database with the provided options
	return OpenDB(db, opts...)
}

// OpenDB 根据用户传递进来的 db 直接初始化 db 实例
func OpenDB(db *sql.DB, opts ...DBOption) (*DB, error) {
	// Initialize a new DB instance with an empty registry.
	res := &DB{
		r:          model.NewRegistry(),   // 构造 sql 的方法
		db:         db,                    // 数据库
		valCreator: unsafe.NewUnsafeValue, // 映射 DB 查询结果的实现
	}

	// Apply each option to the DB instance.
	for _, opt := range opts {
		opt(res)
	}

	// Return the DB instance and no error.
	return res, nil
}

// DBWithRegistry 这里可以替换成不同的映射实例
func DBWithRegistry(r model.Registry) DBOption {
	return func(db *DB) {
		db.r = r
	}
}

// DBUseReflectValuer 如果想使用反射进行获取相关信息，则调用这个 option
func DBUseReflectValuer() DBOption {
	return func(db *DB) {
		db.valCreator = reflect.NewReflectValue
	}
}

// MustNewDB creates a new DB with the provided options.
// If the creation fails, it panics.
func MustNewDB(driver string, dsn string, opts ...DBOption) *DB {
	// Attempt to create a new DB using the provided options.
	db, err := Open(driver, dsn, opts...)
	if err != nil {
		// If an error occurs, panic with the error message.
		panic(err)
	}
	// Return the created DB.
	return db
}
