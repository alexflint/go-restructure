package restructure

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type DotName struct {
	Dot  string `regex:"\\."`
	Name string `regex:"\\w+"`
}

type DotExpr struct {
	_    struct{} `regex:"^"`
	Head string   `regex:"\\w+"`
	Tail *DotName `regex:"?"`
	_    struct{} `regex:"$"`
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
	_      string `regex:"^"`
	Scheme string `regex:"[[:alpha:]]+" json:"scheme"`
	_      string `regex:"://"`
	Host   string `regex:".*" json:"host"`
	_      string `regex:"$"`
}

func TestMatchURL(t *testing.T) {
	pattern, err := Compile(URL{}, Options{})
	require.NoError(t, err)

	var v URL
	require.True(t, pattern.Find(&v, "http://example.com"))
	assert.Equal(t, "http", v.Scheme)
	assert.Equal(t, "example.com", v.Host)
}

func TestCombinationWithJSONTags(t *testing.T) {
	pattern, err := Compile(URL{}, Options{})
	require.NoError(t, err)

	var v URL
	require.True(t, pattern.Find(&v, "http://example.com"))

	js, err := json.Marshal(&v)
	require.NoError(t, err)

	assert.Equal(t, "{\"scheme\":\"http\",\"host\":\"example.com\"}", string(js))
}

type PtrURL struct {
	_      struct{} `regex:"^"`
	Scheme *string  `regex:"[[:alpha:]]+"`
	_      struct{} `regex:"://"`
	Host   *string  `regex:".*"`
	_      struct{} `regex:"$"`
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
