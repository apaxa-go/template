package template

import (
	"errors"
	"io"
	"reflect"
)

type teLoopBlock struct {
	template *Template
	v        placeholders
	els      *Template
}

// *s must begin with loop directive
func parseTELoop(s *string, funcs map[string]interface{}) (b teLoopBlock, err error) {
	directive, _ := extractDirective(s)
	directive = directive[len(loopBlockPrefix):] // Only placeholder definitions

	// Loop block
	var placeholders placeholders
	placeholders, err = parsePlaceholders(directive, funcs)
	if err != nil {
		return
	}
	b.v = placeholders

	var subT *Template
	subT, err = parse(s, funcs)
	if err != nil {
		return
	}
	b.template = subT

	// prepare for else/end block
	directive, err = extractDirective(s)
	if err != nil {
		return
	}

	// Else block
	if directive == optionalElseBlockPrefix {
		var subT *Template
		subT, err = parse(s, funcs)
		if err != nil {
			return
		}

		b.els = subT

		// prepare for end block
		directive, err = extractDirective(s)
		if err != nil {
			return
		}
	}

	// End directive
	if directive != endOfBlock {
		err = errors.New("expect end of loop block, but got '" + directive + "'")
	}
	return
}

func (te teLoopBlock) Execute(wr io.Writer, data []interface{}) error {
	do, err := te.v[0].CompileInterface( data)
	if err != nil {
		return err
	}

	var doEls bool
	switch value := reflect.ValueOf(do); value.Kind() {
	case reflect.Int:
		doEls = value.Int() == 0
		if !doEls {
			args,err:=te.v[1:].CompileInterfaces(data)
			if err!=nil{
				return err
			}
			args=append([]interface{}{nil},args...)	// Add space for 'i'
			for i := 0; i < int(value.Int()); i++ {
				args[0]=i
				err = te.template.execute(wr, args)
				if err != nil {
					return err
				}
			}
		}
	case reflect.Slice:
		doEls = value.Len() == 0
		if !doEls {
			args,err:=te.v[1:].CompileInterfaces(data)
			if err!=nil{
				return err
			}
			args=append([]interface{}{nil,nil},args...)	// Add space for 'i' & do[i]
			for i := 0; i < value.Len(); i++ {
				args[0]=value.Index(i).Interface()
				args[1]=i
				err = te.template.execute(wr, args)
				if err != nil {
					return err
				}
			}
		}
	default:
		return errors.New("in loop block directive alowed only int and slice types, but got " + value.Kind().String())
	}

	if doEls && te.els != nil {
		return te.els.execute(wr, data)
	}

	return nil
}
