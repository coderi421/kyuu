package reflect

import (
	"github.com/coderi421/kyuu/advance/reflect/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_interateFileds(t *testing.T) {
	testCases := []struct {
		name       string
		input      any
		wantFields map[string]any
		wantErr    error
	}{
		{
			// 普通结构体
			name:       "normal struct",
			input:      types.User{Name: "Tom"},
			wantFields: map[string]any{"Name": "Tom", "age": 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := interateFileds(tc.input)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantFields, res)
		})
	}
}
