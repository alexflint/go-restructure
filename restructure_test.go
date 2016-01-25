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
	Head string   `\w+`
	Tail *DotName `?`
	_    struct{} `$`
}

func TestMatchNameDotName(t *testing.T) {
	pattern, err := Compile(DotExpr{}, Options{})
	require.NoError(t, err)

	var v DotExpr
	assert.True(t, pattern.Find(&v, "foo.bar"))
	assert.Equal(t, "foo", v.Head)
	require.NotNil(t, v.Tail)
	assert.Equal(t, ".", v.Tail.Dot)
	assert.Equal(t, "bar", v.Tail.Name)
}

func TestMatchNameDotNameHeadOnly(t *testing.T) {
	pattern, err := Compile(DotExpr{}, Options{})
	require.NoError(t, err)

	var v DotExpr
	assert.True(t, pattern.Find(&v, "head"))
	assert.Equal(t, "head", v.Head)
	assert.Nil(t, v.Tail)
}

func TestMatchNameDotNameFails(t *testing.T) {
	pattern, err := Compile(DotExpr{}, Options{})
	require.NoError(t, err)

	var v DotExpr
	assert.False(t, pattern.Find(&v, ".oops"))
}

type URL struct {
	_      string `^`
	Scheme string `[[:alpha:]]+`
	_      string `://`
	Host   string `.*`
	_      string `$`
}

func TestMatchURL(t *testing.T) {
	pattern, err := Compile(URL{}, Options{})
	require.NoError(t, err)

	var v URL
	require.True(t, pattern.Find(&v, "http://example.com"))
	assert.Equal(t, "http", v.Scheme)
	assert.Equal(t, "example.com", v.Host)
}

type PtrURL struct {
	_      struct{} `^`
	Scheme *string  `[[:alpha:]]+`
	_      struct{} `://`
	Host   *string  `.*`
	_      struct{} `$`
}

func TestMatchPtrURL(t *testing.T) {
	pattern, err := Compile(PtrURL{}, Options{})
	require.NoError(t, err)

	var v PtrURL
	require.True(t, pattern.Find(&v, "http://example.com"))
	require.NotNil(t, v.Scheme)
	require.NotNil(t, v.Host)
	assert.Equal(t, "http", *v.Scheme)
	assert.Equal(t, "example.com", *v.Host)
}

func TestMatchPtrURLFailed(t *testing.T) {
	pattern, err := Compile(PtrURL{}, Options{})
	require.NoError(t, err)

	var v PtrURL
	require.False(t, pattern.Find(&v, "oops"))
	assert.Nil(t, v.Scheme)
	assert.Nil(t, v.Host)
}

type HasSubcaptures struct {
	Name string `a(bc)?d`
}

func TestRemoveSubcaptures(t *testing.T) {
	pattern, err := Compile(HasSubcaptures{}, Options{})
	require.NoError(t, err)

	var v HasSubcaptures
	require.True(t, pattern.Find(&v, "abcd"))
	assert.Equal(t, "abcd", v.Name)
}
