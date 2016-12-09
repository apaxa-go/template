package template

import (
	"errors"
	"reflect"
)

type Func interface {
	Apply(interface{}) (interface{}, error)
}

func callFunction(f reflect.Value, args []reflect.Value) (r []reflect.Value, err error) { // Panic safe
	defer func() {
		if rec := recover(); err != nil {
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
	f interface{}
}

func (f FuncSimple) Apply(a interface{}) (interface{}, error) { // Pass 'a' to FuncSimple and return result and optional error
	value := reflect.ValueOf(f.f)
	if value.Kind() != reflect.Func {
		return nil, errors.New("try to apply not a function: " + value.Kind().String())
	}
	value, err := callFunctionSingleResult(value, []reflect.Value{reflect.ValueOf(a)})
	if err != nil {
		return nil, err
	}
	return value.Interface(), nil
}

type FuncMethod string                                          // Handle method applied to variable (no in arguments, 1 out argument)
func (f FuncMethod) Apply(a interface{}) (interface{}, error) { // Apply method "FuncMethod" to 'a' with no arguments and return result and optional error
	value := reflect.ValueOf(a)
	value = value.MethodByName(string(f))
	if !value.IsValid() {
		return nil, errors.New("unable to call method '" + string(f) + "': no such method")
	}
	value, err := callFunctionSingleResult(value, nil)
	if err != nil {
		return nil, err
	}
	return value.Interface(), nil
}

type teString string

func (te teString) NumArgs() int {
	return 0
}

type tePlaceholder struct {
	name   string
	method bool // If true => .name(), if false => .name
	funcs  []Func
}

func (te tePlaceholder) Compile(v interface{}) (string, error) {
	for _, f := range te.funcs {
		var err error
		v, err = f.Apply(v)
		if err != nil {
			return "", err
		}
	}
	switch v2 := v.(type) {
	case string:
		return v2, nil
	default:
		return "", errors.New("unable to compile placeholder: result not string")
	}
}

func (te tePlaceholder) NumArgs() int {
	return 1
}

type teOptionalBlock struct {
	template Template
	name     string
}

func (te teOptionalBlock) optField() string {
	return te.name + "Opt"
}
func (te teOptionalBlock) NumArgs() int {
	return 1 + te.template.NumArgs() // Bool flag + all sub args
}

type teLoopBlock struct {
	template Template
	name     string
}

func (te teLoopBlock) NumArgs() int {
	return 1 // Number or slice
}

type Te interface {
	NumArgs() int // Number of arguments required for template element. Calculates as for flat compile!!!
}

type Template struct {
	tes []Te
}

func (t *Template) NumArgs() int {
	var r int
	for _, te := range t.tes {
		r += te.NumArgs()
	}
	return r
}

const (
	leftDelim           = "{{"
	rightDelim          = "}}"
	optionalBlockPrefix = "if "
	loopBlockPrefix     = "range "
	endOfBlock          = "end"
	pipe                = "|"
)
