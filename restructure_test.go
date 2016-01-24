package restructure

import (
	"encoding/json"
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
	Head string   `\w+`
	Tail *DotName `?`
	_    struct{} `$`
}

func TestMatchNameDotName(t *testing.T) {
	pattern, err := Compile(DotExpr{}, restructure.Options{})
	require.NoError(t, err)

	var v DotExpr
	assert.True(t, pattern.Match(&v, "foo.bar"))
	assert.Equal(t, "foo", v.Head)
	require.NotNil(t, v.Tail)
	assert.Equal(t, ".", v.Tail.Dot)
	assert.Equal(t, "bar", v.Tail.Name)
}

func TestMatchNameDotNameHeadOnly(t *testing.T) {
	pattern, err := Compile(DotExpr{}, restructure.Options{})
	require.NoError(t, err)

	var v DotExpr
	assert.True(t, pattern.Match(&v, "head"))
	assert.Equal(t, "head", v.Head)
	assert.Nil(t, v.Tail)
}

func TestMatchNameDotNameFails(t *testing.T) {
	pattern, err := Compile(DotExpr{}, restructure.Options{})
	require.NoError(t, err)

	var v DotExpr
	assert.False(t, pattern.Match(&v, ".oops"))
}

type URL struct {
	_      string `^`
	Scheme string `[[:alpha:]]+`
	_      string `://`
	Host   string `[[:alnum:]]+`
	_      string `\?`
	Query  string `.*`
	_      string `$`
}

func TestMatchURL(t *testing.T) {
	pattern, err := Compile(URL{}, restructure.Options{})
	require.NoError(t, err)

	var v URL
	assert.True(t, pattern.Match(&v, "http://example.com"))
	assert.Equal(t, "http", v.Scheme)
	assert.Equal(t, "foo", v.Host)
}

type PtrURL struct {
	_      struct{} `^`
	Scheme *string  `[[:alpha:]]+`
	_      struct{} `://`
	Host   *string  `[[:alnum:]]\w*`
	_      struct{} `$`
}

func TestMatchPtrURL(t *testing.T) {
	pattern, err := Compile(URL{}, restructure.Options{})
	require.NoError(t, err)

	var v PtrURL
	assert.True(t, pattern.Match(&v, "http://example.com"))
	assert.NotNil(t, pattern)
	assert.Equal(t, "http", v.Scheme)
	assert.Equal(t, "foo", v.Host)
}

func prettyPrint(x interface{}) string {
	buf, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(buf)
}
