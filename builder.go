package restructure

import (
	"fmt"
	"reflect"
	"regexp/syntax"
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

func (b *builder) terminal(f reflect.StructField, fullName string) (*Field, *syntax.Regexp, error) {
	pattern := string(f.Tag)
	if pattern == "" {
		return nil, nil, nil
	}

	// TODO: check for sub-captures within expr and remove them
	expr, err := syntax.Parse(pattern, b.opts.SyntaxFlags)
	if err != nil {
		return nil, nil, fmt.Errorf(`%s: %v (pattern was "%s")`, fullName, err, f.Tag)
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
	}

	return field, expr, nil
}

func (b *builder) nonterminal(f reflect.StructField, fullName string) (*Field, *syntax.Regexp, error) {
	opstr := f.Tag
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
	}

	return field, expr, nil
}

func (b *builder) field(f reflect.StructField, fullName string) (*Field, *syntax.Regexp, error) {
	if isScalar(f.Type) {
		return b.terminal(f, fullName)
	}
	return b.nonterminal(f, fullName)
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
		exprs = append(exprs, expr)
		fields = append(fields, field)
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
