package restructure

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/alexflint/go-restructure/regex"
)

var (
	emptyType     = reflect.TypeOf(struct{}{})
	stringType    = reflect.TypeOf("")
	byteArrayType = reflect.TypeOf([]byte{})
	scalarTypes   = []reflect.Type{
		emptyType,
		stringType,
		byteArrayType,
	}
)

type region struct {
	begin, end int
}

func (r region) wasMatched() bool {
	return r.begin != -1 && r.end != -1
}

type match struct {
	input    []byte
	captures []region
}

func matchFromIndices(indices []int, input []byte) *match {
	match := &match{
		input: input,
	}
	for i := 0; i < len(indices); i += 2 {
		match.captures = append(match.captures, region{indices[i], indices[i+1]})
	}
	return match
}

type Regexp struct {
	st *Struct
	re *regex.Regexp
	t  reflect.Type
}

func (r *Regexp) Match(dest interface{}, s string) bool {
	v := reflect.ValueOf(dest)
	input := []byte(s)

	// Check the type
	expected := reflect.PtrTo(r.t)
	if v.Type() != expected {
		panic(fmt.Errorf("expected destination to be %s but got %T",
			expected.String(), dest))
	}

	// Execute the regular expression
	indices := r.re.FindSubmatchIndex(input)
	if indices == nil {
		return false
	}

	// Inflate matches into original struct
	match := matchFromIndices(indices, input)
	// for i, region := range match.captures {
	// 	log.Printf("  %d: %d...%d", i, region.begin, region.end)
	// }

	err := inflateStruct(v, match, r.st)
	if err != nil {
		panic(err)
	}
	return true
}

func Compile(proto interface{}) (*Regexp, error) {
	return CompileType(reflect.TypeOf(proto))
}

func CompileType(t reflect.Type) (*Regexp, error) {
	// Traverse the struct
	b := newBuilder()
	st, expr, err := b.structure(t)
	if err != nil {
		return nil, err
	}

	// Compile regular expression
	re, err := regex.CompileSyntax(expr)
	if err != nil {
		return nil, err
	}

	// Return
	return &Regexp{
		st: st,
		re: re,
		t:  t,
	}, nil
}

func prettyPrint(x interface{}) string {
	buf, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(buf)
}
