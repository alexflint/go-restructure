package restructure

import (
	"fmt"
	"reflect"
)

var (
	posType = reflect.TypeOf(Pos(0))

	emptyType     = reflect.TypeOf(struct{}{})
	stringType    = reflect.TypeOf("")
	byteSliceType = reflect.TypeOf([]byte{})
	submatchType  = reflect.TypeOf(Submatch{})
	scalarTypes   = []reflect.Type{
		emptyType,
		stringType,
		byteSliceType,
		submatchType,
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

// determines whether t is a struct type or a pointer to a struct type
func isStruct(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind() == reflect.Struct
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
func inflateScalar(dest reflect.Value, match *match, captureIndex int, role Role) error {
	if captureIndex == -1 {
		// This means the field generated a regex but we did not want the results
		return nil
	}

	// Get the subcapture for this field
	subcapture := match.captures[captureIndex]
	if !subcapture.wasMatched() {
		// This means the subcapture was optional and was not matched
		return nil
	}

	// Get the matched bytes
	buf := match.input[subcapture.begin:subcapture.end]

	// If dest is a nil pointer then allocate a new instance and assign the pointer to dest
	dest = ensureAlloc(dest)

	// Deal with each recognized type
	switch role {
	case StringScalarRole:
		dest.SetString(string(buf))
		return nil
	case ByteSliceScalarRole:
		dest.SetBytes(buf)
		return nil
	case SubmatchScalarRole:
		submatch := dest.Addr().Interface().(*Submatch)
		submatch.Begin = Pos(subcapture.begin)
		submatch.End = Pos(subcapture.end)
		submatch.Bytes = buf
		return nil
	}
	return fmt.Errorf("unable to capture into %s", dest.Type().String())
}

// inflate the position of a match into a Pos
func inflatePos(dest reflect.Value, match *match, captureIndex int) error {
	if captureIndex == -1 {
		// This means the field generated a regex but we did not want the results
		return nil
	}

	// Get the subcapture for this field
	subcapture := match.captures[captureIndex]
	if !subcapture.wasMatched() {
		// This means the subcapture was optional and was not matched
		return nil
	}

	// If dest is a nil pointer then allocate a new instance and assign the pointer to dest
	dest.SetInt(int64(subcapture.begin))
	return nil
}

// inflate the results of a match into a struct
func inflateStruct(dest reflect.Value, match *match, structure *Struct) error {
	// Get the subcapture for this field
	subcapture := match.captures[structure.capture]
	if !subcapture.wasMatched() {
		return nil
	}

	// If the field is a nil pointer then allocate an instance and assign pointer to dest
	dest = ensureAlloc(dest)

	// Inflate values into the struct fields
	for _, field := range structure.fields {
		switch field.role {
		case PosRole:
			val := dest.FieldByIndex(field.index)
			if err := inflatePos(val, match, field.capture); err != nil {
				return err
			}
		case StringScalarRole, ByteSliceScalarRole, SubmatchScalarRole:
			val := dest.FieldByIndex(field.index)
			if err := inflateScalar(val, match, field.capture, field.role); err != nil {
				return err
			}
		case SubstructRole:
			val := dest.FieldByIndex(field.index)
			if err := inflateStruct(val, match, field.child); err != nil {
				return err
			}
		}
	}
	return nil
}
