package orm

import (
	"context"
	"github.com/coderi421/kyuu/orm/internal/errs"
)

// Selector represents a query selector that allows building SQL SELECT statements.
// It holds the necessary information to construct the query.
type Selector[T any] struct {
	// select delete update insert 都需要使用
	builder

	table string      // table is the name of the table to select from.
	where []Predicate // where holds the WHERE predicates for the query.

	db      *DB // db is the DB instance used for executing the query.
	columns []Selectable
}

// NewSelector creates a new instance of Selector.
func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		db: db,
	}
}

// Select 检索指定 column
func (s *Selector[T]) Select(cols ...Selectable) *Selector[T] {
	s.columns = cols
	return s
}

// From sets the table name for the selector.
// It returns the updated selector.
func (s *Selector[T]) From(tbl string) *Selector[T] {
	s.table = tbl
	return s
}

// Build generates a SQL query for selecting all columns from a table.
// It returns the generated query as a *Query struct or an error if there was any.
func (s *Selector[T]) Build() (*Query, error) {
	var (
		t   T
		err error
	)

	s.model, err = s.db.r.Get(&t)
	if err != nil {
		return nil, err
	}

	s.sb.WriteString("SELECT ")
	if err = s.buildColumns(); err != nil {
		return nil, err
	}
	s.sb.WriteString(" FROM ")

	if s.table == "" {
		s.sb.WriteByte('`')
		// Get the name of the struct using reflection
		s.sb.WriteString(s.model.TableName)
		s.sb.WriteByte('`')
	} else {
		// 这里没有处理 添加`符号，让用户自己应该名字自己在做什么
		s.sb.WriteString(s.table)
	}

	// construct where
	if len(s.where) > 0 {
		// 类似这种可有可无的部分，都要在前面加一个空格
		// 没有将 s.sb.WriteString(" WHERE ") 也放到 buildPredicates 中 是应为可能有 HAVING 的情况
		s.sb.WriteString(" WHERE ")
		// 取出第一个作为开始的节点
		// 构造 谓语相关逻辑
		if err = s.buildPredicates(s.where); err != nil {
			return nil, err
		}
	}

	s.sb.WriteString(";")

	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) buildColumns() error {
	if len(s.columns) == 0 {
		s.sb.WriteByte('*')
		return nil
	}

	for i, c := range s.columns {
		if i > 0 {
			s.sb.WriteByte(',')
		}

		switch val := c.(type) {
		case Column:
			s.sb.WriteByte('`')
			// 找到表中对应名字
			fd, ok := s.model.FieldMap[val.name]
			if !ok {
				return errs.NewErrUnknownField(val.name)
			}
			s.sb.WriteString(fd.ColName)
			s.sb.WriteByte('`')
			s.buildAs(val.alias)
		case Aggregate:
			s.sb.WriteString(val.fn)
			s.sb.WriteString("(`")
			// 找到表中对应名字
			// 这里使用 ORM 的时候，默认使用 struct 的名字作为 column 检索字段
			// for example: Id, Age 然后到内存中储存的 map 中找对应的表中的字段名称
			fd, ok := s.model.FieldMap[val.arg]
			if !ok {
				return errs.NewErrUnknownField(val.arg)
			}
			s.sb.WriteString(fd.ColName)
			s.sb.WriteString("`)")
			s.buildAs(val.alias)
		case RawExpr:
			s.sb.WriteString(val.raw)
			if len(val.args) != 0 {
				s.addArgs(val.args...)
			}
		default:
			return errs.NewErrUnsupportedSelectable(c)
		}
	}

	return nil
}

// Where 用于构造 WHERE 查询条件。如果 ps 长度为 0，那么不会构造 WHERE 部分
func (s *Selector[T]) Where(ps ...Predicate) *Selector[T] {
	// Set the WHERE conditions
	s.where = ps
	return s
}

// cols 是用于 WHERE 的列，难以解决 And Or 和 Not 等问题
// func (s *Selector[T]) Where(cols []string, args...any) *Selector[T] {
// 	s.whereCols = cols
// 	s.args = append(s.args, args...)
// }

// 最为灵活的设计
// func (s *Selector[T]) Where(where string, args...any) *Selector[T] {
// 	s.where = where
// 	s.args = append(s.args, args...)
// }

// Get 根据拼接成的 sql 文，到 db 中获取数据
func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	q, err := s.Build()
	if err != nil {
		return nil, err
	}

	// s.db 是我们定义的 DB
	// s.db.db 则是 sql.DB
	// 使用 QueryContext，从而和 GetMulti 能够复用处理结果集的代码
	rows, err := s.db.db.QueryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, ErrNoRows
	}

	// 创建与 db table 对应的 *struct
	tp := new(T)
	meta, err := s.db.r.Get(tp)
	if err != nil {
		return nil, err
	}

	// 开始进行映射 db table 和 struct 的关系
	val := s.db.valCreator(tp, meta)
	// 使用存在映射关系的实体 val， 将 rows 中的数据 映射到 *struct[T] 中
	err = val.SetColumns(rows)

	return tp, err
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	q, err := s.Build()
	if err != nil {
		return nil, err
	}

	rows, err := s.db.db.QueryContext(ctx, q.SQL, q.Args)
	if err != nil {
		return nil, err
	}
	//var tps []*T
	// 初始化 tp

	for rows.Next() {
		//tp := new(T)
		//meta, e := s.db.r.Get(tp)
		//if e != nil {
		//	return nil, e
		//}
		//// 开始进行映射 db table 和 struct 的关系
		//val := s.db.valCreator(tp, meta)
		//tps = append(tps, new(T))
		// 在这里构造 []*T
	}

	return nil, nil
}

//func (s *Selector[T]) addArgs(args ...any) {
//	if s.args == nil {
//		s.args = make([]any, 0, 8)
//	}
//	s.args = append(s.args, args...)
//}

func (s *Selector[T]) buildAs(alias string) {
	if alias != "" {
		s.sb.WriteString(" AS ")
		s.sb.WriteByte('`')
		s.sb.WriteString(alias)
		s.sb.WriteByte('`')
	}
}

// Selectable 暂时没什么作用只是用作标记，可检索指定字段的标记
// 让结构体实现这个接口，就可以传入
// 使用接口为的是：让 聚合函数， columns， 以及 RawExpr（原生sql） 都能作为参数传入统一个函数，做统一处理
type Selectable interface {
	selectable()
}
