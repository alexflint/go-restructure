package restructure

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type DotName struct {
	Dot  string `regexp:"\\."`
	Name string `regexp:"\\w+"`
}

type DotExpr struct {
	_    struct{} `regexp:"^"`
	Head string   `regexp:"\\w+"`
	Tail *DotName `regexp:"?"`
	_    struct{} `regexp:"$"`
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
	_      string `regexp:"^"`
	Scheme string `regexp:"[[:alpha:]]+" json:"scheme"`
	_      string `regexp:"://"`
	Host   string `regexp:".*" json:"host"`
	_      string `regexp:"$"`
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
	_      struct{} `regexp:"^"`
	Scheme *string  `regexp:"[[:alpha:]]+"`
	_      struct{} `regexp:"://"`
	Host   *string  `regexp:".*"`
	_      struct{} `regexp:"$"`
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

type NakedURL struct {
	_      string `^`
	Scheme string `[[:alpha:]]+`
	_      string `://`
	Host   string `.*`
	_      string `$`
}

func TestMatchNakedURL(t *testing.T) {
	pattern, err := Compile(NakedURL{}, Options{})
	require.NoError(t, err)

	var v NakedURL
	require.True(t, pattern.Find(&v, "http://example.com"))
	assert.Equal(t, "http", v.Scheme)
	assert.Equal(t, "example.com", v.Host)
}

type Nothing struct {
	X string
}

func TestEmptyPattern(t *testing.T) {
	pattern, err := Compile(Nothing{}, Options{})
	require.NoError(t, err)

	var v Nothing
	require.True(t, pattern.Find(&v, "abc"))
}

type Malformed struct {
	X string `regexp:"\w"` // this is malformed because \w is not a valid escape sequence
}

func TestErrorOnMalformedTag(t *testing.T) {
	_, err := Compile(Malformed{}, Options{})
	assert.Error(t, err)
}
