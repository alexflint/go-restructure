package restructure

import (
	"fmt"
	"reflect"
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

// determines whether t is a scalar type or a pointer to a scalar type
func isScalar(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for _, u := range scalarTypes {
		if t == u {
			return true
		}
	}
	return false
}

// ensureAlloc replaces nil pointers with newly allocated objects
func ensureAlloc(dest reflect.Value) reflect.Value {
	if dest.Kind() == reflect.Ptr {
		if dest.IsNil() {
			dest.Set(reflect.New(dest.Type().Elem()))
		}
		return dest.Elem()
	}
	return dest
}

// inflate the results of a match into a string
func inflateScalar(dest reflect.Value, match *match, captureIndex int) error {
	if captureIndex == -1 {
		// This means the field generated a regex but we did not want the results
		return nil
	}
	region := match.captures[captureIndex]
	if !region.wasMatched() {
		// This means the region was optional and was not matched
		return nil
	}

	buf := match.input[region.begin:region.end]

	dest = ensureAlloc(dest)
	switch dest.Type() {
	case stringType:
		dest.SetString(string(buf))
		return nil
	case byteArrayType:
		dest.SetBytes(buf)
		return nil
	case emptyType:
		// ignore the value
		return nil
	}
	return fmt.Errorf("unable to capture into %s", dest.Type().String())
}

// inflate the results of a match into a struct
func inflateStruct(dest reflect.Value, match *match, structure *Struct) error {
	if !match.captures[structure.capture].wasMatched() {
		return nil
	}

	dest = ensureAlloc(dest)
	for _, field := range structure.fields {
		val := dest.FieldByIndex(field.index)
		if isScalar(val.Type()) {
			if err := inflateScalar(val, match, field.capture); err != nil {
				return err
			}
		} else if field.child != nil {
			if err := inflate(val, match, field.child); err != nil {
				return err
			}
		}
	}
	return nil
}

// inflate the results of a match into a union
func inflateUnion(dest reflect.Value, match *match, union *Union) error {
	if dest.Kind() == reflect.Ptr {
		dest = dest.Elem()
	}
	for i, distjunct := range union.disjuncts {
		if match.captures[distjunct.capture].wasMatched() {
			ptr := reflect.New(union.class.structs[i])
			if err := inflateStruct(ptr, match, distjunct); err != nil {
				return err
			}
			dest.Set(ptr)
			return nil
		}
	}
	return nil
}

// inflate the result of a match
func inflate(dest reflect.Value, match *match, class interface{}) error {
	switch class := class.(type) {
	case *Struct:
		if err := inflateStruct(dest, match, class); err != nil {
			return err
		}
	case *Union:
		if err := inflateUnion(dest, match, class); err != nil {
			return err
		}
	default:
		panic(fmt.Errorf("invalid class: %T", class))
	}
	return nil
}
