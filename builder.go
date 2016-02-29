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

// A Union describes how to inflate a match into a registered union type
type Union struct {
	class     *unionType
	disjuncts []*Struct
}

// A Field describes how to inflate a match into a field
type Field struct {
	capture int         // index of the capture for this field
	index   []int       // index of this field within its parent struct
	child   interface{} // descendant struct or union; nil for terminals
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

func (b *builder) terminalField(f reflect.StructField, fullName string) (*Field, *syntax.Regexp, error) {
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

func (b *builder) nonterminalField(f reflect.StructField, fullName string) (*Field, *syntax.Regexp, error) {
	opstr := f.Tag
	child, expr, err := b.build(f.Type)
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
		return b.terminalField(f, fullName)
	}
	return b.nonterminalField(f, fullName)
}

func (b *builder) build(t reflect.Type) (interface{}, *syntax.Regexp, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Struct:
		return b.structure(t)
	case reflect.Interface:
		return b.iface(t)
	default:
		return nil, nil, fmt.Errorf("unable to build regexp for %v", t)
	}
}

func (b *builder) structure(t reflect.Type) (*Struct, *syntax.Regexp, error) {
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

func (b *builder) iface(t reflect.Type) (*Union, *syntax.Regexp, error) {
	unionType, found := unions[t]
	if !found {
		return nil, nil, fmt.Errorf("interface types can only be used when registered with RegisterUnion")
	}

	var structs []*Struct
	var exprs []*syntax.Regexp
	for _, structType := range unionType.structs {
		st, expr, err := b.structure(structType)
		if err != nil {
			return nil, nil, err
		}
		structs = append(structs, st)
		exprs = append(exprs, expr)
	}

	// Wrap in a disjunction
	expr := &syntax.Regexp{
		Sub: exprs,
		Op:  syntax.OpAlternate,
	}

	union := &Union{
		class:     unionType,
		disjuncts: structs,
	}

	return union, expr, nil
}
