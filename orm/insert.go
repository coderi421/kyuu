package orm

import (
	"context"
	"database/sql"
	"github.com/coderi421/kyuu/orm/internal/errs"
	"github.com/coderi421/kyuu/orm/model"
	"reflect"
)

type UpsertBuilder[T any] struct {
	i               *Inserter[T]
	conflictColumns []string // 由于不同 DB insertOrUpdate 语法不同，所以这里将 更新字段抽成共通
}

type Upsert struct {
	conflictColumns []string     // 为 sqlite3 ON CONFLICT (id) 这种语法准备的
	assigns         []Assignable // 插入失败后，更新的字段的切片
}

func (o *UpsertBuilder[T]) ConflictColumns(cols ...string) *UpsertBuilder[T] {
	o.conflictColumns = cols
	return o
}

// Update
//
//	@Description: 如果存在，则更新指定字段
//	@receiver o
//	@param assigns
//	@return *Inserter[T]
func (o *UpsertBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.onDuplicate = &Upsert{
		assigns:         assigns,
		conflictColumns: o.conflictColumns,
	}
	return o.i
}

type Inserter[T any] struct {
	builder
	values  []*T     // 缓存要插入的数据
	db      *DB      // 注册映射关系的实例，以及使用哪种映射方法的实例，以及 DB 实例
	columns []string // update 语句中，要更新哪些字段
	// 方案二
	onDuplicate *Upsert // 对应存在即更新语句： ON DUPLICATE KEY UPDATE

	// 方案一
	// onDuplicate []Assignable
}

func NewInserter[T any](db *DB) *Inserter[T] {
	return &Inserter[T]{
		db: db,
		builder: builder{
			quoter:  db.dialect.quoter(),
			dialect: db.dialect,
		},
	}
}

// Values
// Fields 指定要插入的列
// TODO 目前我们只支持指定具体的列，但是不支持复杂的表达式
// 例如不支持 VALUES(..., now(), now()) 这种在 MySQL 里面常用的
//
//	@Description: 将插入数据库中的数据
//	@receiver i
//	@param val
//	@return *Inserter[T]
func (i *Inserter[T]) Values(val ...*T) *Inserter[T] {
	i.values = val
	return i
}

// Columns
//
//	@Description: 更新指定的字段
//	@receiver i
//	@param cols
//	@return *Inserter[T]
func (i *Inserter[T]) Columns(cols ...string) *Inserter[T] {
	i.columns = cols
	return i
}

func (i *Inserter[T]) OnDuplicateKey() *UpsertBuilder[T] {
	return &UpsertBuilder[T]{
		i: i,
	}
}

func (i *Inserter[T]) Build() (*Query, error) {
	if len(i.values) == 0 {
		return nil, errs.ErrInsertZeroRow
	}
	// 由于多条数据都一样，同一个 struct 所以这里处理第一条就可以拿到 db field 和 struct 的映射关系
	m, err := i.db.r.Get(i.values[0])
	if err != nil {
		return nil, err
	}
	i.model = m

	i.sb.WriteString("INSERT INTO ")
	i.quote(m.TableName)
	i.sb.WriteString(" (")

	fields := m.Fields
	if len(i.columns) != 0 {
		// 如果只更新部分字段
		fields = make([]*model.Field, 0, len(i.columns))
		for _, c := range i.columns {
			field, ok := m.FieldMap[c]
			if !ok {
				return nil, errs.NewErrUnknownField(c)
			}
			fields = append(fields, field)
		}
	}

	// (len(i.values) + 1) 中 +1 是考虑到 UPSERT 语句会传递额外的参数
	i.args = make([]any, 0, len(fields)*len(i.values)+1)
	for idx, fd := range fields {
		if idx > 0 {
			i.sb.WriteByte(',')
		}
		i.quote(fd.ColName)
	}

	i.sb.WriteString(") VALUES ")
	for vIdx, val := range i.values {
		// 构建 VALUES (?,?,?), (?,?,?)
		if vIdx > 0 {
			i.sb.WriteByte(',')
		}
		// 由于是泛型，所以这里使用反射取值
		refVal := reflect.ValueOf(val).Elem()
		i.sb.WriteByte('(')
		for fIdx, filed := range fields {
			// 构建 (?,?,?)
			if fIdx > 0 {
				i.sb.WriteByte(',')
			}
			i.sb.WriteByte('?')
			// 由于 refVal 中的是所有的数据，所以需要确定第几个数据是我们需要的字段
			fdVal := refVal.Field(filed.Index)
			i.addArgs(fdVal.Interface())
		}
		i.sb.WriteByte(')')
	}

	if i.onDuplicate != nil {
		err = i.dialect.buildUpsert(&i.builder, i.onDuplicate)
		if err != nil {
			return nil, err
		}
	}

	i.sb.WriteByte(';')
	return &Query{
		SQL:  i.sb.String(),
		Args: i.args,
	}, nil
}

func (i *Inserter[T]) buildAssignment(a Assignable) error {
	switch assign := a.(type) {
	case Column:
		// 使用原本插入的值
		// "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?),(?,?,?,?) ON DUPLICATE KEY UPDATE `first_name`=VALUES(`first_name`),`last_name`=VALUES(`last_name`);"
		i.sb.WriteByte('`')
		fd, ok := i.model.FieldMap[assign.name]
		if !ok {
			return errs.NewErrUnknownField(assign.name)
		}
		i.sb.WriteString(fd.ColName)
		i.sb.WriteString("`=VALUES(`")
		i.sb.WriteString(fd.ColName)
		i.sb.WriteString("`)")
	case Assignment:
		// "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?) ON DUPLICATE KEY UPDATE `first_name`=?;"
		i.sb.WriteByte('`')
		fd, ok := i.model.FieldMap[assign.column]
		if !ok {
			return errs.NewErrUnknownField(assign.column)
		}
		i.sb.WriteString(fd.ColName)
		i.sb.WriteByte('`')
		i.sb.WriteString("=?")
		i.addArgs(assign.val)
	}
	return nil
}

func (i *Inserter[T]) Exec(ctx context.Context) sql.Result {
	q, err := i.Build()
	if err != nil {
		return Result{err: err}
	}
	res, err := i.db.db.ExecContext(ctx, q.SQL, q.Args...)
	return Result{
		err: err,
		res: res,
	}
}
