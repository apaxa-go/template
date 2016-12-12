package template

import (
	"errors"
	"github.com/apaxa-go/helper/strconvh"
	"io"
	"reflect"
)

type teLoopBlock struct {
	template *Template
	v        tePlaceholder
	els      *Template
}

func (te teLoopBlock) NumArgs() int {
	r := 1 // Number or slice
	if te.els != nil {
		r += te.els.NumArgs()
	}
	return r
}

// *s must begin with loop directive
func parseTELoop(s *string, funcs map[string]interface{}) (b teLoopBlock, err error) {
	directive, _ := extractDirective(s)
	directive = directive[len(loopBlockPrefix):] // Only placeholder definition

	// Loop block
	var placeholder tePlaceholder
	placeholder, err = parsePlaceholder(directive, funcs)
	if err != nil {
		return
	}
	b.v = placeholder

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

func (te teLoopBlock) Execute(wr io.Writer, topData, parentData, data interface{}) error {
	do, err := te.v.CompileInterface(topData, parentData, data)
	if err != nil {
		return err
	}

	var doEls bool
	switch value := reflect.ValueOf(do); value.Kind() {
	case reflect.Int:
		doEls = value.Int() == 0
		for i := 0; i < int(value.Int()); i++ {
			err = te.template.execute(wr, topData, data, i)
			if err != nil {
				return err
			}
		}
	case reflect.Slice:
		doEls = value.Len() == 0
		for i := 0; i < value.Len(); i++ {
			err = te.template.execute(wr, topData, data, value.Index(i).Interface())
			if err != nil {
				return err
			}
		}
	default:
		return errors.New("in loop block directive alowed only int and slice types, but got " + value.Kind().String())
	}

	if doEls && te.els != nil {
		return te.els.execute(wr, topData, parentData, data)
	}

	return nil
}

func (te teLoopBlock) ExecuteFlat(wr io.Writer, data []interface{}, dataI *int) error {
	if *dataI >= len(data) {
		return errors.New("not enough arguments")
	}
	value := reflect.ValueOf(data[*dataI])

	var doEls bool
	switch value.Kind() {
	case reflect.Int:
		doEls = value.Int() == 0
		for i := 0; i < int(value.Int()); i++ {
			err := te.template.execute(wr, data, data[*dataI], i) // pass data[dataI] as parent data instead of nil because it may be hard to work with root data as []interface{}
			if err != nil {
				return err
			}
		}
	case reflect.Slice:
		doEls = value.Len() == 0
		for i := 0; i < value.Len(); i++ {
			err := te.template.execute(wr, data, data[*dataI], value.Index(i).Interface()) // pass data[dataI] as parent data instead of nil because it may be hard to work with root data as []interface{}
			if err != nil {
				return err
			}
		}
	default:
		return errors.New("in loop block directive alowed only int and slice types, but got " + value.Kind().String() + " (" + strconvh.FormatInt(*dataI) + ")")
	}
	*dataI++ // still for loop condition

	if te.els != nil {
		if doEls {
			return te.els.executeFlat(wr, data, dataI)
		}
		*dataI += te.els.NumArgs()
	}

	return nil
}
