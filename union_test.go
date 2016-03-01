package restructure

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	var scalar Scalar
	RegisterUnion(&scalar, StringLiteral{}, IntLiteral{})
}

type Scalar interface{}

type StringLiteral struct {
	_   struct{} `"`
	Str string   `[^"]*`
	_   struct{} `"`
}

type IntLiteral struct {
	Digits string `[0-9]+`
}

type Assignment struct {
	Var string   `\w+`
	_   struct{} `=`
	Val Scalar
}

func TestUnionMatch(t *testing.T) {
	var assign Assignment
	pat := MustCompile(&assign, Options{})
	require.True(t, pat.Find(&assign, "abc=123"))

	assert.Equal(t, "abc", assign.Var)
	require.IsType(t, &IntLiteral{}, assign.Val)

	lit := assign.Val.(*IntLiteral)
	assert.Equal(t, "123", lit.Digits)
}

func TestUnionTopLevelMatch(t *testing.T) {
	var lit Scalar
	pat := MustCompile(&lit, Options{})
	require.True(t, pat.Find(&lit, `"xyz"`))
	require.IsType(t, &StringLiteral{}, lit)
	assert.Equal(t, "xyz", lit.(*StringLiteral).Str)
}

func TestUnionNonmatch(t *testing.T) {
	var assign Assignment
	pat := MustCompile(&assign, Options{})
	assert.False(t, pat.Find(&assign, "abc=x7"))
}
