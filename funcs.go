package template

import (
	"errors"
	"reflect"
)

// TODO may be move some of it to other packages?

func callFunction(f reflect.Value, args []reflect.Value) (r []reflect.Value, err error) { // Panic safe
	defer func() {
		if rec := recover(); rec != nil {
			if str, ok := rec.(string); ok {
				err = errors.New(str)
			} else if e, ok := rec.(error); ok {
				err = e
			} else {
				err = errors.New("unable to call function")
			}
		}
	}()
	r = f.Call(args)
	return
}

func callFunctionSingleResult(f reflect.Value, args []reflect.Value) (r reflect.Value, err error) {
	values, err := callFunction(f, args)
	if err != nil {
		return
	}
	if len(values) != 1 {
		return reflect.Value{}, errors.New("unable to call function: multiple or no result have been returned")
	}
	return values[0], nil
}

type FuncSimple struct { // Handle function type with 1 in argument and 1 out argument
	f interface{} // function
}

func (f FuncSimple) Apply(a reflect.Value) (reflect.Value, error) { // Pass 'a' to FuncSimple and return result and optional error
	fValue := reflect.ValueOf(f.f)
	if fValue.Kind() != reflect.Func {
		return reflect.Value{}, errors.New("try to apply not a function: " + fValue.Kind().String())
	}

	return callFunctionSingleResult(fValue, []reflect.Value{a})
}

type FuncMethod string // Handle method applied to variable (no in arguments, 1 out argument)

func (f FuncMethod) Apply(a reflect.Value) (reflect.Value, error) { // Apply method "FuncMethod" to 'a' with no arguments and return result and optional error
	// First try to call method directly
	fValue := a.MethodByName(string(f))
	if fValue.IsValid() {
		return callFunctionSingleResult(fValue, nil)
	}

	origType:=a.Type().String()
	// Second try to call method on value or on pointer
	if a.Kind()==reflect.Ptr{
		a=a.Elem()
	}else{
		a,_=addrGetter.Apply(struct {}{},a)
	}
	fValue = a.MethodByName(string(f))
	if fValue.IsValid() {
		return callFunctionSingleResult(fValue, nil)
	}

	return reflect.Value{}, errors.New("unable to call method '" + string(f) + "': no such method on " + origType)
}

type FuncGetter string // Handle field name to extract

func (f FuncGetter) Apply(a reflect.Value) (reflect.Value, error) {
	// Dereference a if it is a pointer
	if a.Kind()==reflect.Ptr{
		a=a.Elem()
	}

	//
	if a.Kind() != reflect.Struct {
		return reflect.Value{}, errors.New("unable to get field '" + string(f) + "' from " + a.Kind().String())
	}
	field := a.FieldByName(string(f))
	if !field.IsValid() {
		return reflect.Value{}, errors.New("no such field: '" + string(f) + "'")
	}
	return field, nil
}
