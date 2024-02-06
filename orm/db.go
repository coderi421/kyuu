package orm

import (
	"context"
	"database/sql"
	"github.com/coderi421/kyuu/orm/internal/errs"
	"github.com/coderi421/kyuu/orm/internal/valuer/reflect"
	"github.com/coderi421/kyuu/orm/internal/valuer/unsafe"
	"github.com/coderi421/kyuu/orm/model"
)

type DBOption func(*DB)

type DB struct {
	core
	//dialect    Dialect
	//r          model.Registry // 存储数据库表和 struct 映射关系的实例
	//valCreator valuer.Creator // 与DB交互映射的实现
	db *sql.DB
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
		core: core{
			dialect:    MySQL,                 // 默认方言为 mysql
			r:          model.NewRegistry(),   // 构造 sql 的方法
			valCreator: unsafe.NewUnsafeValue, // 映射 DB 查询结果的实现
		},
		db: db, // 数据库
	}

	// Apply each option to the DB instance.
	for _, opt := range opts {
		opt(res)
	}

	// Return the DB instance and no error.
	return res, nil
}

// DBWithRegistry 这里可以替换成不同的映射实例
func DBWithDialect(d Dialect) DBOption {
	return func(db *DB) {
		db.dialect = d
	}
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

// BeginTx 开启事务
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx, db: db}, nil
}

// type txKey struct {
//
// }

// BeginTxV2 事务扩散
// 个人不太喜欢
// func (db *DB) BeginTxV2(ctx context.Context,
// 	opts *sql.TxOptions) (context.Context, *Tx, error) {
// 	val := ctx.Value(txKey{})
// 	if val != nil {
// 		tx := val.(*Tx)
// 		if !tx.done {
// 			return ctx, tx, nil
// 		}
// 	}
// 	tx, err := db.BeginTx(ctx, opts)
// 	if err != nil {
// 		return ctx, nil, err
// 	}
// 	ctx = context.WithValue(ctx, txKey{}, tx)
// 	return ctx, tx, nil
// }

// DoTx 将会开启事务执行 fn。如果 fn 返回错误或者发生 panic，事务将会回滚，
// 否则提交事务
func (db *DB) DoTx(ctx context.Context,
	fn func(ctx context.Context, tx *Tx) error,
	opts *sql.TxOptions) (err error) {
	var tx *Tx
	tx, err = db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	panicked := true
	defer func() {
		if panicked || err != nil {
			e := tx.Rollback()
			if e != nil {
				err = errs.NewErrFailToRollbackTx(err, e, panicked)
			}
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(ctx, tx)
	panicked = false
	return err
}

func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) getCore() core {
	return db.core
}

func (db *DB) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.db.QueryContext(ctx, query, args...)
}

func (db *DB) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.db.ExecContext(ctx, query, args...)
}
