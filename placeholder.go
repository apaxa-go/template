package template

import (
	"errors"
	"github.com/apaxa-go/helper/strconvh"
	"io"
	"reflect"
	"strings"
)

type Func interface {
	Apply(reflect.Value) (reflect.Value, error)
}

type tePlaceholder struct {
	argNum int
	funcs  []Func
}

type placeholders []tePlaceholder

// * => get value of parent address (dereference parent value).
// & => get address of parent value.
// .Field => get field Field of parent value (with optional parent value preparation).
// .Method() => get result of execution method Method on parent value. Method must return single result.
// Function() => get result of execution function Function with passing parent value as single argument.
func parsePlaceholder(s string, funcs map[string]interface{}) (p tePlaceholder, err error) {
	const (
		dot        = "."
		invoke     = "()"
		indexLeft  = "["
		indexRight = "]"
	)

	if s==""{
		return
	}

	strs := strings.Split(s, pipe)
	skipFirst := false
	if len(strs[0])>0 && strs[0][0] == 'a' {
		if i, err := strconvh.ParseInt(strs[0][1:]); err == nil {
			p.argNum = i
			skipFirst = true
		}
	}

	if skipFirst {
		strs = strs[1:]
	}

	for _, str := range strs {
		if len(str) == 0 {
			err = errors.New("unable to parse empty placeholder in '" + s + "'")
			return
		}
		switch {
		case str == "*":
			p.funcs = append(p.funcs, dereferencer{})
		case str == "&":
			p.funcs = append(p.funcs, addrGetter{})
		case strings.HasPrefix(str, dot) && strings.HasSuffix(str, invoke): // Method
			name := str[len(dot) : len(str)-len(invoke)]
			if !IsValidExportedIdent(name) {
				err = errors.New("invalid method name: '" + name + "'")
				return
			}
			p.funcs = append(p.funcs, FuncMethod(name))
		case strings.HasPrefix(str, dot): // Field
			name := str[len(dot):]
			if !IsValidExportedIdent(name) {
				err = errors.New("invalid field name: '" + name + "'")
				return
			}
			p.funcs = append(p.funcs, FuncGetter(name))
		case strings.HasSuffix(str, invoke): // Function
			name := str[:len(str)-len(invoke)]
			f, ok := funcs[name]
			if !ok {
				err = errors.New("unknown function '" + name + "'")
				return
			}
			p.funcs = append(p.funcs, FuncSimple{f})
		case strings.HasPrefix(str, indexLeft) && strings.HasSuffix(str, indexRight): // Access by index
			iStr := str[len(indexLeft) : len(str)-len(indexRight)]
			var i int
			i, err = strconvh.ParseInt(iStr)
			if err != nil {
				return
			}
			p.funcs = append(p.funcs, Index(i))
		default:
			err = errors.New("unknown element in placeholder: '" + str + "'")
			return
		}
	}
	return
}

func parsePlaceholders(s string, funcs map[string]interface{}) (p placeholders, err error) {
	args := strings.Split(s, ",")
	p = make([]tePlaceholder, len(args))
	for i, arg := range args {
		p[i], err = parsePlaceholder(arg, funcs)
		if err != nil {
			return
		}
	}
	return
}

func (ps placeholders) CompileInterfaces(data []interface{}) (r []interface{}, err error) {
	r = make([]interface{}, len(ps))
	for i, p := range ps {
		r[i], err = p.CompileInterface(data)
		if err != nil {
			return
		}
	}
	return
}

func (te tePlaceholder) CompileInterface(data []interface{}) (interface{}, error) {
	value := reflect.ValueOf(data[te.argNum])
	for _, f := range te.funcs {
		var err error
		value, err = f.Apply(value)
		if err != nil {
			return nil, err
		}
	}
	if !value.IsValid() {
		return nil, nil
	}
	return value.Interface(), nil
}

func (te tePlaceholder) Compile(data []interface{}) (string, error) {
	v, err := te.CompileInterface(data)
	if err != nil {
		return "", err
	}
	switch v2 := v.(type) {
	case string:
		return v2, nil
	default:
		return "", errors.New("unable to compile placeholder: result is "+reflect.TypeOf(v).String()+", not a string")
	}
}

func (te tePlaceholder) CompileBool(data []interface{}) (bool, error) {
	v, err := te.CompileInterface(data)
	if err != nil {
		return false, err
	}
	switch v2 := v.(type) {
	case bool:
		return v2, nil
	default:
		return false, errors.New("unable to compile placeholder: result is "+reflect.TypeOf(v).String()+", not a bool")
	}
}

func (te tePlaceholder) CompileInt(data []interface{}) (int, error) {
	v, err := te.CompileInterface(data)
	if err != nil {
		return 0, err
	}
	switch v2 := v.(type) {
	case int:
		return v2, nil
	default:
		return 0, errors.New("unable to compile placeholder: result is "+reflect.TypeOf(v).String()+", not an int")
	}
}

func (te tePlaceholder) Execute(wr io.Writer, data []interface{}) error {
	str, err := te.Compile(data)
	if err != nil {
		return err
	}
	_, err = wr.Write([]byte(str))
	return err
}
