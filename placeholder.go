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

type dataLevel int

const (
	currentDataLevel dataLevel = iota
	parentDataLevel            = iota
	topDataLevel               = iota
)

type tePlaceholder struct {
	dataLevel dataLevel
	funcs     []Func
}

func (te tePlaceholder) NumArgs() int {
	return 1
}

// * => get value of parent address (dereference parent value).
// & => get address of parent value.
// .Field => get field Field of parent value (with optional parent value preparation).
// .Method() => get result of execution method Method on parent value. Method must return single result.
// Function() => get result of execution function Function with passing parent value as single argument.
func parsePlaceholder(s string, funcs map[string]interface{}) (p tePlaceholder, err error) {
	const (
		dot    = "."
		invoke = "()"
		indexLeft="["
		indexRight="]"
	)

	strs := strings.Split(s, pipe)
	skipFirst := true
	switch strs[0] {
	case "":
		fallthrough
	case ".":
		p.dataLevel = currentDataLevel
	case "..":
		p.dataLevel = parentDataLevel
	case "...":
		p.dataLevel = topDataLevel
	default:
		p.dataLevel = currentDataLevel
		skipFirst = false
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
		case strings.HasPrefix(str, indexLeft) && strings.HasSuffix(str,indexRight):	// Access by index
			iStr:=str[len(indexLeft):len(str)-len(indexRight)]
			var i int
			i,err = strconvh.ParseInt(iStr)
			if err!=nil{
				return
			}
			p.funcs=append(p.funcs, Index(i))
		default:
			err = errors.New("unknown element in placeholder: '" + str + "'")
			return
		}
	}
	return
}

func (te tePlaceholder) CompileInterface(topData, parentData, data interface{}) (interface{}, error) {
	var value reflect.Value
	switch te.dataLevel {
	case currentDataLevel:
		value = reflect.ValueOf(data)
	case parentDataLevel:
		value = reflect.ValueOf(parentData)
	case topDataLevel:
		value = reflect.ValueOf(topData)
	}
	for _, f := range te.funcs {
		var err error
		value, err = f.Apply(value)
		if err != nil {
			return nil, err
		}
	}
	if !value.IsValid(){
		return nil,nil
	}
	return value.Interface(), nil
}

func (te tePlaceholder) Compile(topData, parentData, data interface{}) (string, error) {
	v, err := te.CompileInterface(topData, parentData, data)
	if err != nil {
		return "", err
	}
	switch v2 := v.(type) {
	case string:
		return v2, nil
	default:
		return "", errors.New("unable to compile placeholder: result not a string")
	}
}

func (te tePlaceholder) CompileBool(topData, parentData, data interface{}) (bool, error) {
	v, err := te.CompileInterface(topData, parentData, data)
	if err != nil {
		return false, err
	}
	switch v2 := v.(type) {
	case bool:
		return v2, nil
	default:
		return false, errors.New("unable to compile placeholder: result not a bool")
	}
}

func (te tePlaceholder) CompileInt(topData, parentData, data interface{}) (int, error) {
	v, err := te.CompileInterface(topData, parentData, data)
	if err != nil {
		return 0, err
	}
	switch v2 := v.(type) {
	case int:
		return v2, nil
	default:
		return 0, errors.New("unable to compile placeholder: result not an int")
	}
}

func (te tePlaceholder) Execute(wr io.Writer, topData, parentData, data interface{}) error {
	str, err := te.Compile(topData, parentData, data)
	if err != nil {
		return err
	}
	_, err = wr.Write([]byte(str))
	return err
}

func (te tePlaceholder) ExecuteFlat(wr io.Writer, data []interface{}, dataI *int) error {
	if *dataI >= len(data) {
		return errors.New("not enough arguments")
	}
	str, ok := data[*dataI].(string)
	*dataI++
	if !ok {
		return errors.New("unable to compile placeholder: result not a string (" + strconvh.FormatInt(*dataI) + ")")
	}
	_, err := wr.Write([]byte(str))
	return err
}
