package restructure

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertRegion(t *testing.T, s string, begin int, end int, r *Submatch) {
	assert.NotNil(t, r)
	assert.Equal(t, s, string(r.Bytes))
	assert.EqualValues(t, begin, r.Begin)
	assert.EqualValues(t, end, r.End)
}

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

type DotNameRegion struct {
	Dot  *Submatch `regexp:"\\."`
	Name *Submatch `regexp:"\\w+"`
}

type DotExprRegion struct {
	_    struct{}       `regexp:"^"`
	Head Submatch       `regexp:"\\w+"`
	Tail *DotNameRegion `regexp:"?"`
	_    struct{}       `regexp:"$"`
}

func TestMatchNameDotNameRegion(t *testing.T) {
	pattern, err := Compile(DotExprRegion{}, Options{})
	require.NoError(t, err)

	var v DotExprRegion
	assert.True(t, pattern.Find(&v, "foo.bar"))
	assertRegion(t, "foo", 0, 3, &v.Head)
	assert.NotNil(t, v.Tail)
	assertRegion(t, ".", 3, 4, v.Tail.Dot)
	assertRegion(t, "bar", 4, 7, v.Tail.Name)
}

type DotNamePos struct {
	Begin  Pos
	Dot    string `regexp:"\\."`
	Middle Pos
	Name   string `regexp:"\\w+"`
	End    Pos
}

type DotExprPos struct {
	Begin  Pos
	_      struct{} `regexp:"^"`
	Head   string   `regexp:"\\w+"`
	Middle Pos
	Tail   *DotNamePos `regexp:"?"`
	_      struct{}    `regexp:"$"`
	End    Pos
}

func TestMatchNameDotNamePos(t *testing.T) {
	pattern, err := Compile(DotExprPos{}, Options{})
	require.NoError(t, err)

	var v DotExprPos
	assert.True(t, pattern.Find(&v, "foo.bar"))
	assert.EqualValues(t, 0, v.Begin)
	assert.EqualValues(t, 3, v.Middle)
	assert.EqualValues(t, 3, v.Tail.Begin)
	assert.EqualValues(t, 4, v.Tail.Middle)
	assert.EqualValues(t, 7, v.Tail.End)
	assert.EqualValues(t, 7, v.End)
}

type DegeneratePos struct {
	X Pos
	Y Pos
}

func TestDegeneratePos(t *testing.T) {
	// This tests what happens if there are degenerate position captures
	pattern, err := Compile(DegeneratePos{}, Options{})
	require.NoError(t, err)
	var v DegeneratePos
	assert.True(t, pattern.Find(&v, "abc"))
	assert.EqualValues(t, 0, v.X)
	assert.EqualValues(t, 0, v.Y)
}

type UnexportedPos struct {
	Exported   Pos
	unexported Pos
	_          struct{} `regexp:"$"`
}

func TestUnexportedPos(t *testing.T) {
	// This tests what happens if there are non-exported Pos fields
	pattern, err := Compile(UnexportedPos{}, Options{})
	require.NoError(t, err)
	var v UnexportedPos
	assert.True(t, pattern.Find(&v, "abc"))
	assert.EqualValues(t, 3, v.Exported)
	assert.EqualValues(t, 0, v.unexported) // should be ignored
}

type Word struct {
	S string `\w+`
}

func TestFindAllWords_Simple(t *testing.T) {
	pattern := MustCompile(Word{}, Options{})
	var words []Word
	pattern.FindAll(&words, "ham is spam", -1)
	require.Len(t, words, 3)
	assert.EqualValues(t, "ham", words[0].S)
	assert.EqualValues(t, "is", words[1].S)
	assert.EqualValues(t, "spam", words[2].S)
}

func TestFindAllWords_Ptr(t *testing.T) {
	pattern := MustCompile(Word{}, Options{})
	var words []*Word
	pattern.FindAll(&words, "ham is spam", -1)
	require.Len(t, words, 3)
	assert.EqualValues(t, "ham", words[0].S)
	assert.EqualValues(t, "is", words[1].S)
	assert.EqualValues(t, "spam", words[2].S)
}

func TestFindAllWords_NoMatches(t *testing.T) {
	pattern := MustCompile(Word{}, Options{})
	var words []*Word
	pattern.FindAll(&words, "*&!", -1)
	require.Empty(t, words)
}

func TestFindAllWords_ByValueSlicePanics(t *testing.T) {
	pattern := MustCompile(Word{}, Options{})
	var words []*Word
	// This should panic because words is passed by value not by pointer:
	assert.Panics(t, func() { pattern.FindAll(words, "*&!", -1) })
}

type WordSubmatch struct {
	S *Submatch `\w+`
}

func TestFindAllWords_Regions(t *testing.T) {
	pattern := MustCompile(WordSubmatch{}, Options{})
	var words []*WordSubmatch
	pattern.FindAll(&words, "ham is spam", -1)
	require.Len(t, words, 3)
	assertRegion(t, "ham", 0, 3, words[0].S)
	assertRegion(t, "is", 4, 6, words[1].S)
	assertRegion(t, "spam", 7, 11, words[2].S)
}

type ExprWithInt struct {
	Number int    `regexp:"^\\d+"`
	_      string `regexp:"\\s+"`
	Animal string `regexp:"\\w+$"`
}

func TestMatchWithInt(t *testing.T) {
	pattern, err := Compile(ExprWithInt{}, Options{})
	require.NoError(t, err)

	var v ExprWithInt
	assert.True(t, pattern.Find(&v, "4 wombats"))
	assert.Equal(t, 4, v.Number)
	assert.Equal(t, "wombats", v.Animal)
}
