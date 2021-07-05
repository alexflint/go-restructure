package restructure

import (
	"fmt"
	"reflect"
)

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
