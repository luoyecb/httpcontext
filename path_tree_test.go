package httpcontext

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	resultValue = 0
)

type testStruct struct {
	path         string
	result       int
	handlerIsNil bool
	params       map[string]string
}

func Test(t *testing.T) {
	assert := assert.New(t)

	tree := NewPathTree()

	tree.Put("/", func(*Context) { resultValue = 1 })
	tree.Put("/api/v1/user", func(*Context) { resultValue = 2 })
	tree.Put("/api/v1/{:user}/add", func(*Context) { resultValue = 3 })
	tree.Put("/api/v1/user/", func(*Context) { resultValue = 4 })

	tests := []*testStruct{
		{"/", 1, false, nil},
		{"/api/v1/lisi/add", 3, false, map[string]string{"user": "lisi"}},
		{"/api/v1/user", 2, false, nil},
		{"/api/v1/user/del", 0, true, nil},
		{"/api/v1/user/add", 0, true, nil},
		{"/api/v1/user/", 4, false, nil},
	}
	for _, test := range tests {
		handler, _, params := tree.FindHandler(test.path)

		if test.handlerIsNil {
			assert.Nil(handler)
		} else {
			assert.NotNil(handler)

			handler(&Context{})
			assert.Equal(test.result, resultValue)

			if test.params != nil {
				assert.Equal(test.params, params)
			}
		}
	}
}
