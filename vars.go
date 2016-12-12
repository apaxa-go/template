package template

import (
	"errors"
	"reflect"
)

// error is always nil, it is just for interface
func GetAddr(a reflect.Value) (reflect.Value, error) {
	if a.CanAddr() {
		return a.Addr(), nil
	}

	vp := reflect.New(a.Type())
	vp.Elem().Set(a)
	return vp, nil
}

func Dereference(a reflect.Value) (reflect.Value, error) {
	if kind := a.Kind(); kind != reflect.Ptr {
		return reflect.Value{}, errors.New("unable to dereference " + kind.String())
	}
	return a.Elem(), nil
}
