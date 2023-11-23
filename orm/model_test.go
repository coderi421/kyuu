package orm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestModelWithTableName(t *testing.T) {
	testCases := []struct {
		name          string
		val           any
		opt           ModelOpt
		wantTableName string
		wantErr       error
	}{
		{
			name:          "empty string",
			val:           &TestModel{},
			opt:           ModelWithTableName(""),
			wantTableName: "",
		},
		{
			name:          "table name",
			val:           &TestModel{},
			opt:           ModelWithTableName("test_model_t"),
			wantTableName: "test_model_t",
		},
	}

	r := NewRegistry()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.Register(tc.val, tc.opt)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantTableName, m.tableName)
		})
	}
}
