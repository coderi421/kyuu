package orm

import "github.com/coderi421/kyuu/orm/internal/errs"

var (
	MySQL   Dialect = &mysqlDialect{}
	SQLite3 Dialect = &sqlite3Dialect{}
)

type Dialect interface {
	quoter() byte
	buildUpsert(b *builder, u *Upsert) error
}

type standardSQL struct {
}

func (s *standardSQL) quoter() byte {
	//TODO implement me
	panic("implement me")
}

func (s *standardSQL) buildUpsert(b *builder, u *Upsert) error {
	//TODO implement me
	panic("implement me")
}

type mysqlDialect struct {
	standardSQL
}

func (m *mysqlDialect) quoter() byte {
	return '`'
}

func (m *mysqlDialect) buildUpsert(b *builder, u *Upsert) error {
	b.sb.WriteString(" ON DUPLICATE KEY UPDATE ")
	for idx, a := range u.assigns {
		if idx > 0 {
			b.sb.WriteByte(',')
		}

		switch assign := a.(type) {
		case Column:
			// 使用原本插入的值
			// "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?),(?,?,?,?) ON DUPLICATE KEY UPDATE `first_name`=VALUES(`first_name`),`last_name`=VALUES(`last_name`);"
			fd, ok := b.model.FieldMap[assign.name]
			if !ok {
				return errs.NewErrUnknownField(assign.name)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=VALUES(")
			b.quote(fd.ColName)
			b.sb.WriteString(")")
		case Assignment:
			// "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?) ON DUPLICATE KEY UPDATE `first_name`=?;"
			// 字段不对，或者说列不对
			fd, ok := b.model.FieldMap[assign.column]
			if !ok {
				return errs.NewErrUnknownField(assign.column)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=?")
			b.addArgs(assign.val)
		default:
			return errs.NewErrUnsupportedAssignableType(assign)
		}
	}
	return nil
}

type sqlite3Dialect struct {
	standardSQL
}

func (s *sqlite3Dialect) quoter() byte {
	return '`'
}

func (s *sqlite3Dialect) buildUpsert(b *builder, u *Upsert) error {
	b.sb.WriteString(" ON CONFLICT")
	if len(u.conflictColumns) > 0 {
		b.sb.WriteByte('(')
		for i, col := range u.conflictColumns {
			if i > 0 {
				b.sb.WriteByte(',')
			}
			err := b.buildColumn(col)
			if err != nil {
				return err
			}
		}
		b.sb.WriteByte(')')
	}
	b.sb.WriteString(" DO UPDATE SET ")

	for idx, assign := range u.assigns {
		if idx > 0 {
			b.sb.WriteByte(',')
		}
		switch assign := assign.(type) {
		case Column:
			fd, ok := b.model.FieldMap[assign.name]
			if !ok {
				return errs.NewErrUnknownField(assign.name)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=excluded.")
			b.quote(fd.ColName)
		case Assignment:
			err := b.buildColumn(assign.column)
			if err != nil {
				return err
			}
			b.sb.WriteString("=?")
			b.addArgs(assign.val)
		default:
			return errs.NewErrUnsupportedAssignableType(assign)
		}
	}
	return nil
}
