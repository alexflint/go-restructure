package restructure

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type DotName struct {
	Dot  string `\.`
	Name string `\w+`
}

type DotExpr struct {
	_    struct{} `^`
	Head string   `foo`
	Tail *DotName `?`
	_    struct{} `$`
}

func TestMatchNameDotName(t *testing.T) {
	pattern, err := Compile(DotExpr{})
	require.NoError(t, err)

	var v DotExpr
	assert.True(t, pattern.Match(&v, "foo.bar"))
	assert.Equal(t, "foo", v.Head)
	require.NotNil(t, v.Tail)
	assert.Equal(t, ".", v.Tail.Dot)
	assert.Equal(t, "bar", v.Tail.Name)
}
