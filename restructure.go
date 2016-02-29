package restructure

import (
	"fmt"
	"reflect"
	"regexp/syntax"

	"github.com/alexflint/go-restructure/regex"
)

// Style determines whether we are in Perl or POSIX or custom mode
type Style int

const (
	Perl Style = iota
	POSIX
	CustomStyle
)

// Options represents optional parameters for compilation
type Options struct {
	Style       Style // Style can be set to Perl, POSIX, or CustomStyle
	SyntaxFlags syntax.Flags
}

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

// Regexp is a regular expression that captures submatches into struct fields.
type Regexp struct {
	class interface{}
	re    *regex.Regexp
	t     reflect.Type
	opts  Options
}

// Find attempts to match the regular expression against the input string. It
// returns true if there was a match, and also populates the fields of the provided
// struct with the contents of each submatch.

// Find attempts to match the regular expression against the input string. It
// returns true if there was a match, and also populates the fields of the provided
// struct with the contents of each submatch.
func (r *Regexp) Find(dest interface{}, s string) bool {
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

	err := inflate(v, match, r.class)
	if err != nil {
		panic(err)
	}
	return true
}

// String returns a string representation of the regular expression
func (r *Regexp) String() string {
	return r.re.String()
}

// Compile constructs a regular expression from the struct fields on the
// provided struct.
func Compile(proto interface{}, opts Options) (*Regexp, error) {
	return CompileType(reflect.TypeOf(proto), opts)
}

// CompileType is like Compile but takes a reflect.Type instead.
func CompileType(t reflect.Type, opts Options) (*Regexp, error) {
	// We do this so that the zero value for Options gives us Perl mode,
	// which is also the default used by the standard library regexp package
	switch opts.Style {
	case Perl:
		opts.SyntaxFlags = syntax.Perl
	case POSIX:
		opts.SyntaxFlags = syntax.POSIX
	}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Traverse the struct
	b := newBuilder(opts)
	class, expr, err := b.build(t)
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
		class: class,
		re:    re,
		t:     t,
		opts:  opts,
	}, nil
}

// MustCompile is like Compile but panics if there is a compilation error
func MustCompile(proto interface{}, opts Options) *Regexp {
	re, err := Compile(proto, opts)
	if err != nil {
		panic(err)
	}
	return re
}

// MustCompileType is like CompileType but panics if there is a compilation error
func MustCompileType(t reflect.Type, opts Options) *Regexp {
	re, err := CompileType(t, opts)
	if err != nil {
		panic(err)
	}
	return re
}

// Find constructs a regular expression from the given struct and executes it on the
// given string, placing submatches into the fields of the struct. The first parameter
// must be a non-nil struct pointer. It returns true if the match succeeded. The only
// errors that are returned are compilation errors.
func Find(dest interface{}, s string) (bool, error) {
	re, err := Compile(dest, Options{})
	if err != nil {
		return false, err
	}
	return re.Find(dest, s), nil
}
