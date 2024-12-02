package restructure

import (
	"errors"
	"fmt"
	"reflect"
	"regexp/syntax"
	"strings"
)

// A Role determines how a struct field is inflated
type Role int

const (
	EmptyRole Role = iota
	PosRole
	SubstructRole
	StringScalarRole
	IntScalarRole
	ByteSliceScalarRole
	SubmatchScalarRole
)

// A Struct describes how to inflate a match into a struct
type Struct struct {
	capture int
	fields  []*Field
}

// A Field describes how to inflate a match into a field
type Field struct {
	capture int     // index of the capture for this field
	index   []int   // index of this field within its parent struct
	child   *Struct // descendant struct; nil for terminals
	role    Role
}

func isExported(f reflect.StructField) bool {
	return f.PkgPath == ""
}

// A builder builds stencils from structs using reflection
type builder struct {
	numCaptures int
	opts        Options
}

func newBuilder(opts Options) *builder {
	return &builder{
		opts: opts,
	}
}

func (b *builder) nextCaptureIndex() int {
	k := b.numCaptures
	b.numCaptures++
	return k
}

func (b *builder) extractTag(tag reflect.StructTag) (string, error) {
	// Allow tags that look like either `regexp:"\\w+"` or just `\w+`
	if s := tag.Get("regexp"); s != "" {
		return s, nil
	} else if strings.Contains(string(tag), `regexp:"`) {
		return "", errors.New("incorrectly escaped struct tag")
	} else {
		return string(tag), nil
	}
}

func removeCaptures(expr *syntax.Regexp) ([]*syntax.Regexp, error) {
	if expr.Op == syntax.OpCapture {
		return expr.Sub, nil
	}
	return []*syntax.Regexp{expr}, nil
}

func (b *builder) terminal(f reflect.StructField, fullName string) (*Field, *syntax.Regexp, error) {
	pattern, err := b.extractTag(f.Tag)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %v", fullName, err)
	}
	if pattern == "" {
		return nil, nil, nil
	}

	// Parse the pattern
	expr, err := syntax.Parse(pattern, b.opts.SyntaxFlags)
	if err != nil {
		return nil, nil, fmt.Errorf(`%s: %v (pattern was "%s")`, fullName, err, f.Tag)
	}

	// Remove capture nodes within the AST
	expr, err = transform(expr, removeCaptures)
	if err != nil {
		return nil, nil, fmt.Errorf(`failed to remove captures from "%s": %v`, pattern, err)
	}

	// Determine the kind
	t := f.Type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	var role Role
	switch t {
	case emptyType:
		role = EmptyRole
	case stringType:
		role = StringScalarRole
	case intType:
		role = IntScalarRole
	case byteSliceType:
		role = ByteSliceScalarRole
	case submatchType:
		role = SubmatchScalarRole
	}

	captureIndex := -1
	if isExported(f) {
		captureIndex = b.nextCaptureIndex()
		expr = &syntax.Regexp{
			Op:   syntax.OpCapture,
			Sub:  []*syntax.Regexp{expr},
			Name: f.Name,
			Cap:  captureIndex,
		}
	}
	field := &Field{
		index:   f.Index,
		capture: captureIndex,
		role:    role,
	}

	return field, expr, nil
}

func (b *builder) pos(f reflect.StructField, fullName string) (*Field, *syntax.Regexp, error) {
	if !isExported(f) {
		return nil, nil, nil
	}
	captureIndex := b.nextCaptureIndex()
	empty := &syntax.Regexp{
		Op: syntax.OpEmptyMatch,
	}
	expr := &syntax.Regexp{
		Op:   syntax.OpCapture,
		Sub:  []*syntax.Regexp{empty},
		Name: f.Name,
		Cap:  captureIndex,
	}
	field := &Field{
		index:   f.Index,
		capture: captureIndex,
		role:    PosRole,
	}

	return field, expr, nil
}

func (b *builder) nonterminal(f reflect.StructField, fullName string) (*Field, *syntax.Regexp, error) {
	opstr, err := b.extractTag(f.Tag)
	if err != nil {
		return nil, nil, err
	}
	child, expr, err := b.structure(f.Type)
	if err != nil {
		return nil, nil, err
	}

	switch opstr {
	case "?":
		if f.Type.Kind() != reflect.Ptr {
			return nil, nil, fmt.Errorf(`%s is marked with "?" but is not a pointer`, fullName)
		}
		expr = &syntax.Regexp{
			Sub: []*syntax.Regexp{expr},
			Op:  syntax.OpQuest,
		}
	case "":
		// nothing to do
	default:
		return nil, nil, fmt.Errorf("invalid op \"%s\" for non-slice field on %s", opstr, fullName)
	}

	captureIndex := b.nextCaptureIndex()
	expr = &syntax.Regexp{
		Op:   syntax.OpCapture,
		Sub:  []*syntax.Regexp{expr},
		Name: f.Name,
		Cap:  captureIndex,
	}
	field := &Field{
		index:   f.Index,
		capture: captureIndex,
		child:   child,
		role:    SubstructRole,
	}

	return field, expr, nil
}

func (b *builder) field(f reflect.StructField, fullName string) (*Field, *syntax.Regexp, error) {
	if isScalar(f.Type) {
		return b.terminal(f, fullName)
	} else if isStruct(f.Type) {
		return b.nonterminal(f, fullName)
	} else if f.Type == posType {
		return b.pos(f, fullName)
	}
	return nil, nil, nil
}

func (b *builder) structure(t reflect.Type) (*Struct, *syntax.Regexp, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Select a capture index first so that the struct comes before its fields
	captureIndex := b.nextCaptureIndex()

	var exprs []*syntax.Regexp
	var fields []*Field
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		field, expr, err := b.field(f, t.Name()+"."+f.Name)
		if err != nil {
			return nil, nil, err
		}
		if field != nil {
			exprs = append(exprs, expr)
			fields = append(fields, field)
		}
	}

	// Wrap in a concat
	expr := &syntax.Regexp{
		Sub: exprs,
		Op:  syntax.OpConcat,
	}

	// Wrap in a capture
	expr = &syntax.Regexp{
		Sub: []*syntax.Regexp{expr},
		Op:  syntax.OpCapture,
		Cap: captureIndex,
	}

	st := &Struct{
		fields:  fields,
		capture: captureIndex,
	}

	return st, expr, nil
}
