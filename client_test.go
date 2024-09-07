package easy_http

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetQueryParam(t *testing.T) {
	c := MustNewClient(&Option{})

	c.NewRequest().SetQueryParam()

	c.NewRequest().set

	assert.Equal(t, "test1", c.QueryParam.Get("test1"))
	assert.Equal(t, "test2", c.QueryParam.Get("test2"))
	assert.Equal(t, "test3", c.QueryParam.Get("test3"))
	assert.Equal(t, []string{"test41", "test42"}, c.QueryParam["test4"])
	assert.Equal(t, "test5", c.QueryParam.Get("test5"))
}
