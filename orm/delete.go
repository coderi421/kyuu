package orm

type Deleter[T any] struct {
	builder

	table string
	where []Predicate
	db    *DB
}

// NewSelector creates a new instance of Selector.
func NewDeleter[T any](db *DB) *Deleter[T] {
	return &Deleter[T]{
		db: db,
	}
}

// Build generates a DELETE query based on the provided parameters.
// It returns the generated query string and any associated arguments,
// or an error if there was a problem building the query.
func (d *Deleter[T]) Build() (*Query, error) {
	var (
		t   T
		err error
	)

	// 从缓存中读取model
	d.model, err = d.db.r.Get(&t)
	if err != nil {
		return nil, err
	}

	_, _ = d.sb.WriteString("DELETE FROM ")

	// If the table name is not provided, use the name of the T struct.
	if d.table == "" {
		d.sb.WriteByte('`')
		d.sb.WriteString(d.model.TableName)
		d.sb.WriteByte('`')
	} else {
		d.sb.WriteString(d.table)
	}

	// If there are any WHERE clauses, add them to the query.
	if len(d.where) > 0 {
		d.sb.WriteString(" WHERE ")
		err = d.buildPredicates(d.where)
		if err != nil {
			return nil, err
		}
	}

	d.sb.WriteByte(';')
	return &Query{
		SQL:  d.sb.String(),
		Args: d.args,
	}, nil
}

// From sets the table for the Deleter and returns a pointer to the Deleter.
// The table parameter specifies the name of the table to delete from.
func (d *Deleter[T]) From(table string) *Deleter[T] {
	d.table = table
	return d
}

// Where accepts predicates and adds them to the Deleter's where clause.
//
// Parameters:
// predicates: A list of predicates to add to the where clause.
//
// Returns:
// *Deleter[T]: The Deleter object with the updated where clause.
func (d *Deleter[T]) Where(predicates ...Predicate) *Deleter[T] {
	d.where = predicates
	return d
}
