package ufwb

import (
	"fmt"
	"github.com/101loops/iszero"
	"reflect"
)

// extend copies all fields from src into dst which are zero, recursing into sub fields
func extend(dst, src interface{}) error {

	if dst == nil {
		return fmt.Errorf("dst is null", dst)
	}

	if src == nil {
		return fmt.Errorf("src is null", dst)
	}

	dstV := reflect.ValueOf(dst)
	srcV := reflect.ValueOf(src)

	//dstT := reflect.TypeOf(dst)
	//srcT := reflect.TypeOf(src)

	if dstV.Type() != srcV.Type() {
		return fmt.Errorf("dst of %T must be same type of src %T", dst, src)
	}

	if k := dstV.Kind(); k == reflect.Interface || k == reflect.Ptr {

		if srcV.IsNil() {
			return nil
		}

		dstV = dstV.Elem()
		srcV = srcV.Elem()
	}

	for i := 0; i < srcV.NumField(); i++ {

		d := dstV.Field(i)
		s := srcV.Field(i)

		switch d.Kind() {
		case reflect.Array, reflect.Slice:
			// Iterate and copy each element
			for j := 0; j < s.Len(); j++ {
				err := extend(d.Index(j).Interface(), s.Index(j).Interface())
				if err != nil {
					return err
				}
			}

		case reflect.Struct, reflect.Ptr:
			extend(d.Interface(), s.Interface())

		case reflect.Map:
			return fmt.Errorf("map type not supported")

		default: // Really just normal fields
			if iszero.Value(d) {
				// Skip dst field if already set
				d.Set(srcV.Field(i))
			}
		}

	}

	return nil
}
