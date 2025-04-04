package util

import (
	"errors"
	"reflect"
)

func DeepMerge[T any](a, b T) (T, error) {
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	if aVal.Kind() == reflect.Ptr && bVal.Kind() == reflect.Ptr {
		aVal = aVal.Elem()
		bVal = bVal.Elem()
	} else {
		return a, errors.New("both arguments must be pointers to structs")
	}

	// Create a new object of the same type to store the merge
	merged := reflect.New(aVal.Type()).Elem()

	for i := 0; i < aVal.NumField(); i++ {
		aField := aVal.Field(i)
		bField := bVal.Field(i)

		if !aField.CanSet() {
			continue
		}

		// If bField is not zero, it takes precedence
		if !isZeroValue(bField) {
			if aField.Kind() == reflect.Struct {
				deeperMerged, err := DeepMerge(aField.Interface(), bField.Interface())
				if err != nil {
					return a, err
				}
				merged.Field(i).Set(reflect.ValueOf(deeperMerged))
			} else if aField.Kind() == reflect.Map {
				// Handle merging of maps
				for _, key := range bField.MapKeys() {
					merged.Field(i).SetMapIndex(key, bField.MapIndex(key))
				}
			} else {
				merged.Field(i).Set(bField)
			}
		} else {
			merged.Field(i).Set(aField)
		}
	}

	return merged.Interface().(T), nil
}

func isZeroValue(v reflect.Value) bool {
	return v.Interface() == reflect.Zero(v.Type()).Interface()
}
