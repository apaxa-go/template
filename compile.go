package template

import (
	"errors"
	"github.com/apaxa-go/helper/strconvh"
	"io"
	"reflect"
)

func (t *Template) CompileSimple() (string, error) {
	var r string
	for _, te := range t.tes {
		switch te := te.(type) {
		case teString:
			r += string(te)
		case tePlaceholder:
			return "", errors.New("no required arguments")
		case teOptionalBlock:
			return "", errors.New("no required arguments")
		case teLoopBlock:
			return "", errors.New("no required arguments")
		default:
			return "", errors.New("unexpected type of template element")
		}
	}
	return r, nil
}
func (t *Template) ExecuteSimple(wr io.Writer) error {
	for _, te := range t.tes {
		switch te := te.(type) {
		case teString:
			if _, err := wr.Write([]byte(te)); err != nil {
				return err
			}
		case tePlaceholder:
			return errors.New("no required arguments")
		case teOptionalBlock:
			return errors.New("no required arguments")
		case teLoopBlock:
			return errors.New("no required arguments")
		default:
			return errors.New("unexpected type of template element")
		}
	}
	return nil
}
func (t *Template) Compile(data interface{}) (string, error) {
	value := reflect.ValueOf(data)
	if value.Kind() != reflect.Struct {
		return "", errors.New("data should be if type struct")
	}
	var r string
	for _, te := range t.tes {
		switch te := te.(type) {
		case teString:
			r += string(te)
		case tePlaceholder:
			var subValue reflect.Value
			if te.method{
				var err error
				subValue,err=callFunctionSingleResult(value.MethodByName(te.name),nil)
				if err!=nil{
					return "",err
				}
			}else{
				subValue = value.FieldByName(te.name)
			}
			if !subValue.IsValid() {
				return "", errors.New("no required field/method " + te.name)
			}
			str, err := te.Compile(subValue.Interface())
			if err != nil {
				return "", err
			}
			r += str
		case teOptionalBlock: // Must have optField (bool), may have field (struct)
			switch subValue := value.FieldByName(te.optField()); subValue.Kind() {
			case reflect.Bool:
				if subValue.Bool() {
					subValue = value.FieldByName(te.name)
					var str string
					var err error
					if !subValue.IsValid() { // Call sub template without arguments
						str, err = te.template.CompileSimple()
					} else {
						str, err = te.template.Compile(subValue.Interface())
					}
					if err != nil {
						return "", err
					}
					r += str
				}
			default:
				return "", errors.New("no required field (bool) " + string(te.optField()))
			}
		case teLoopBlock: // Must have field ([]struct or int)
			switch subValue := value.FieldByName(te.name); subValue.Kind() {
			case reflect.Slice:
				for i := 0; i < subValue.Len(); i++ {
					str, err := te.template.Compile(subValue.Index(i).Interface())
					if err != nil {
						return "", err
					}
					r += str
				}
			case reflect.Int:
				for i := 0; i < int(subValue.Int()); i++ {
					str, err := te.template.CompileSimple()
					if err != nil {
						return "", err
					}
					r += str
				}
			default:
				return "", errors.New("no required slice field (struct ot int) " + te.name)
			}
		default:
			return "", errors.New("unexpected type of template element")
		}
	}
	return r, nil
}

func (t *Template) Execute(wr io.Writer, data interface{}) error {
	value := reflect.ValueOf(data)
	if value.Kind() != reflect.Struct {
		return errors.New("data should be if type struct")
	}
	for _, te := range t.tes {
		switch te := te.(type) {
		case teString:
			if _, err := wr.Write([]byte(te)); err != nil {
				return err
			}
		case tePlaceholder:
			var subValue reflect.Value
			if te.method{
				var err error
				subValue,err=callFunctionSingleResult(value.MethodByName(te.name),nil)
				if err!=nil{
					return err
				}
			}else{
				subValue = value.FieldByName(te.name)
			}
			if !subValue.IsValid() {
				return errors.New("no required field/method " + te.name)
			}
			str, err := te.Compile(subValue.Interface())
			if err != nil {
				return  err
			}
			if _, err := wr.Write([]byte(str)); err != nil {
				return err
			}
		case teOptionalBlock: // Must have optField (bool), may have field (struct)
			switch subValue := value.FieldByName(te.optField()); subValue.Kind() {
			case reflect.Bool:
				if subValue.Bool() {
					subValue = value.FieldByName(te.name)
					var err error
					if !subValue.IsValid() { // Call sub template without arguments
						err = te.template.ExecuteSimple(wr)
					} else {
						err = te.template.Execute(wr, subValue.Interface())
					}
					if err != nil {
						return err
					}
				}
			default:
				return errors.New("no required field (bool) " + string(te.optField()))
			}
		case teLoopBlock: // Must have field ([]struct or int)
			switch subValue := value.FieldByName(te.name); subValue.Kind() {
			case reflect.Slice:
				for i := 0; i < subValue.Len(); i++ {
					err := te.template.Execute(wr, subValue.Index(i).Interface())
					if err != nil {
						return nil
					}
				}
			case reflect.Int:
				for i := 0; i < int(subValue.Int()); i++ {
					err := te.template.ExecuteSimple(wr)
					if err != nil {
						return nil
					}
				}
			default:
				return errors.New("no required slice field (struct ot int) " + te.name)
			}
		default:
			return errors.New("unexpected type of template element")
		}
	}
	return nil
}

// 'if' te are inlined, 'loop' te may be inlined (if depends only on number of iterations, so int used; in other words if its template require no argument) or called as usual (if slice is used; in other words if its template require some arguments).
// Because of possibly compilation number of arguments must be constant for specific template.
// So for 'if' blocks arguments should be passed even if block disabled.
// Arguments should be passed in order it is in template.
// There is no smart variable manipulation: if variable used more than once it must be passed at each position it used.
func (t *Template) compileFlat(data []interface{}, dataI *int) (string, error) {
	var r string
	for _, te := range t.tes {
		switch te := te.(type) {
		case teString:
			r += string(te)
		case tePlaceholder:
			str, err := te.Compile(data[*dataI])
			if err != nil {
				return "", err
			}
			*dataI++
			r += str
		case teOptionalBlock:
			b, ok := data[*dataI].(bool)
			if !ok {
				return "", errors.New("required field " + strconvh.FormatInt(*dataI) + " (" + string(te.optField()) + ") is not of type bool")
			}
			*dataI++
			if b {
				str, err := te.template.compileFlat(data, dataI)
				if err != nil {
					return "", err
				}
				r += str
			} else {
				*dataI += te.template.NumArgs()
			}
		case teLoopBlock:
			if te.template.NumArgs() == 0 { // TODO copy this logic (dependency on NumArgs) to original compiles/executes
				count, ok := data[*dataI].(int)
				if !ok {
					return "", errors.New("required field " + strconvh.FormatInt(*dataI) + " (" + te.name + ") is not of type int")
				}
				*dataI++
				for i := 0; i < count; i++ {
					str, err := te.template.CompileSimple()
					if err != nil {
						return "", err
					}
					r += str
				}
			} else {
				subValue := reflect.ValueOf(data[*dataI])
				if subValue.Kind() != reflect.Slice {
					return "", errors.New("required field " + strconvh.FormatInt(*dataI) + " (" + te.name + ") is not of type slice")
				}
				*dataI++
				for i := 0; i < subValue.Len(); i++ {
					str, err := te.template.Compile(subValue.Index(i).Interface())
					if err != nil {
						return "", err
					}
					r += str
				}
			}
		default:
			return "", errors.New("unexpected type of template element")
		}
	}
	return r, nil
}

func (t *Template) CompileFlat(data ...interface{}) (string, error) {
	dataI := 0
	str, err := t.compileFlat(data, &dataI)
	if err != nil {
		return "", err
	}
	if dataI != len(data) {
		return "", errors.New("more than need arguments: required " + strconvh.FormatInt(dataI) + ", got " + strconvh.FormatInt(len(data)))
	}
	return str, nil
}

func (t *Template) executeFlat(wr io.Writer, data []interface{}, dataI *int) error {
	for _, te := range t.tes {
		switch te := te.(type) {
		case teString:
			if _, err := wr.Write([]byte(te)); err != nil {
				return err
			}
		case tePlaceholder:
			str, err := te.Compile(data[*dataI])
			if err != nil {
				return err
			}
			*dataI++
			if _, err := wr.Write([]byte(str)); err != nil {
				return err
			}
		case teOptionalBlock:
			b, ok := data[*dataI].(bool)
			if !ok {
				return errors.New("required field " + strconvh.FormatInt(*dataI) + " (" + string(te.optField()) + ") is not of type bool")
			}
			*dataI++
			if b {
				if err := te.template.executeFlat(wr, data, dataI); err != nil {
					return err
				}
			} else {
				*dataI += te.template.NumArgs()
			}
		case teLoopBlock:
			if te.template.NumArgs() == 0 {
				count, ok := data[*dataI].(int)
				if !ok {
					return errors.New("required field " + strconvh.FormatInt(*dataI) + " (" + te.name + ") is not of type int")
				}
				*dataI++
				for i := 0; i < count; i++ {
					if err := te.template.ExecuteSimple(wr); err != nil {
						return err
					}
				}
			} else {
				subValue := reflect.ValueOf(data[*dataI])
				if subValue.Kind() != reflect.Slice {
					return errors.New("required field " + strconvh.FormatInt(*dataI) + " (" + te.name + ") is not of type slice")
				}
				*dataI++
				for i := 0; i < subValue.Len(); i++ {
					if err := te.template.Execute(wr, subValue.Index(i).Interface()); err != nil {
						return err
					}
				}
			}
		default:
			return errors.New("unexpected type of template element")
		}
	}
	return nil
}

func (t *Template) ExecuteFlat(wr io.Writer, data ...interface{}) error {
	dataI := 0
	if err := t.executeFlat(wr, data, &dataI); err != nil {
		return err
	}
	if dataI != len(data) {
		return errors.New("more than need arguments: required " + strconvh.FormatInt(dataI) + ", got " + strconvh.FormatInt(len(data)))
	}
	return nil
}
