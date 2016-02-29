package restructure

import (
	"fmt"
	"reflect"
)

type unionType struct {
	iface   reflect.Type
	structs []reflect.Type
}

var (
	// unions is a map from interface types to a union type
	unions = make(map[reflect.Type]*unionType)
)

// RegisterUnions registers an interface type
func RegisterUnion(ifacePtr interface{}, structs ...interface{}) {
	// Get the interface type
	ifacePtrType := reflect.TypeOf(ifacePtr)
	if ifacePtrType.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("expected a pointer to an interface but got %T", ifacePtr))
	}
	ifaceType := ifacePtrType.Elem()
	if ifaceType.Kind() != reflect.Interface {
		panic(fmt.Sprintf("expected a pointer to an interface but got %s", ifaceType))
	}

	// Check for duplicates
	if _, present := unions[ifaceType]; present {
		panic(fmt.Sprintf("%v already registered", ifaceType))
	}

	// Check the struct types
	var structTypes []reflect.Type
	for _, st := range structs {
		structType := reflect.TypeOf(st)
		if structType.Kind() != reflect.Struct {
			panic(fmt.Sprintf("expected a struct but got %T", st))
		}

		// check that the struct implements the interface
		if !structType.Implements(ifaceType) {
			panic(fmt.Sprintf("%s does not implement %s", structType, ifaceType))
		}

		structTypes = append(structTypes, structType)
	}

	unions[ifaceType] = &unionType{
		iface:   ifaceType,
		structs: structTypes,
	}
}
