package gecko

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTopicExprParse(t *testing.T) {
	te := newTopicExpr("/device/+/status/#")
	assert.Equal(t, "device", te.exprs[0])
	assert.Equal(t, "+", te.exprs[1])
	assert.Equal(t, "status", te.exprs[2])
	assert.Equal(t, "#", te.exprs[3])
}

func TestTopicExprMatchesDynamic(t *testing.T) {
	te := newTopicExpr("/device/+/status/#")
	assert.True(t, te.matches("/device/11/status/success"))
	assert.True(t, te.matches("/device/22/status/success/0"))
	assert.True(t, te.matches("/device/33/status/error"))
	assert.True(t, te.matches("/device/AA/status/error/1"))
	assert.True(t, te.matches("/device/BB/status/error/1/2"))
	assert.False(t, te.matches(""))
	assert.False(t, te.matches("/"))
	assert.False(t, te.matches("/device"))
	assert.False(t, te.matches("/device/123"))
	assert.False(t, te.matches("/device/123/sensors"))
	assert.False(t, te.matches("/device/123/sensors/00"))
	assert.False(t, te.matches("/user/1000"))
	assert.False(t, te.matches("/user"))
}

func TestTopicExprMatchesStatic(t *testing.T) {
	te := newTopicExpr("/device/0755/status/error")
	assert.True(t, te.matches("/device/0755/status/error"))
	assert.False(t, te.matches(""))
	assert.False(t, te.matches("/"))
	assert.False(t, te.matches("/device"))
	assert.False(t, te.matches("/device/123"))
	assert.False(t, te.matches("/device/123/sensors"))
	assert.False(t, te.matches("/device/123/sensors/00"))
	assert.False(t, te.matches("/user/1000"))
	assert.False(t, te.matches("/user"))
}
