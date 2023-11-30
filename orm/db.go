package orm

import "github.com/coderi421/kyuu/orm/model"

type DBOption func(*DB)

type DB struct {
	r model.Registry
}

// NewDB creates a new instance of DB with the provided options.
// It returns a pointer to the DB and an error if any.
func NewDB(opts ...DBOption) (*DB, error) {
	// Initialize a new DB instance with an empty registry.
	db := &DB{
		r: model.NewRegistry(),
	}

	// Apply each option to the DB instance.
	for _, opt := range opts {
		opt(db)
	}

	// Return the DB instance and no error.
	return db, nil
}

// MustNewDB creates a new DB with the provided options.
// If the creation fails, it panics.
func MustNewDB(opts ...DBOption) *DB {
	// Attempt to create a new DB using the provided options.
	db, err := NewDB(opts...)
	if err != nil {
		// If an error occurs, panic with the error message.
		panic(err)
	}
	// Return the created DB.
	return db
}
