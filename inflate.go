package restructure

import (
	"fmt"
	"reflect"
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
	region := match.captures[captureIndex]
	if !region.wasMatched() {
		return nil
	}

	dest = ensureAlloc(dest)
	buf := match.input[region.begin:region.end]
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
			if err := inflateStruct(val, match, field.child); err != nil {
				return err
			}
		}
	}
	return nil
}
