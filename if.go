package template

import (
	"errors"
	"github.com/apaxa-go/helper/strconvh"
	"io"
	"strings"
)

type teOptionalBlockElement struct {
	condition tePlaceholder
	template  *Template
}
type teOptionalBlock struct {
	ifs []teOptionalBlockElement
	els *Template
}

func (te teOptionalBlock) NumArgs() int {
	var r int
	for i := range te.ifs {
		r++ // for condition placeholder
		r += te.ifs[i].template.NumArgs()
	}
	if te.els != nil {
		r += te.els.NumArgs()
	}
	return r
}

// *s must begin with if directive
func parseTEOptional(s *string, funcs map[string]interface{}) (b teOptionalBlock, err error) {
	directive, _ := extractDirective(s)
	directive = directive[len(optionalBlockPrefix):] // Only placeholder definition
	// if & else-if blocks
	for true {
		var placeholder tePlaceholder
		placeholder, err = parsePlaceholder(directive, funcs)
		if err != nil {
			return
		}

		var subT *Template
		subT, err = parse(s, funcs)
		if err != nil {
			return
		}

		b.ifs = append(b.ifs, teOptionalBlockElement{placeholder, subT})

		// prepare for next iteration
		directive, err = extractDirective(s)
		if err != nil {
			return
		}
		if !strings.HasPrefix(directive, optionalElseIfBlockPrefix) {
			break
		}
		directive = directive[len(optionalElseIfBlockPrefix):]
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
		err = errors.New("expect end of if block, but got '" + directive + "'")
	}
	return
}

func (te teOptionalBlock) Execute(wr io.Writer, topData, parentData, data interface{}) error {
	for _, b := range te.ifs {
		do, err := b.condition.CompileBool(topData, parentData, data)
		if err != nil {
			return err
		}
		if do {
			return b.template.execute(wr, topData, parentData, data)
		}
	}
	if te.els != nil {
		return te.els.execute(wr, topData, parentData, data)
	}
	return nil
}

func (te teOptionalBlock) ExecuteFlat(wr io.Writer, data []interface{}, dataI *int) error {
	done := false
	for _, b := range te.ifs {
		var do = false
		if !done {
			if *dataI >= len(data) {
				return errors.New("not enough arguments")
			}
			var ok bool
			do, ok = data[*dataI].(bool)
			if !ok {
				return errors.New("argument " + strconvh.FormatInt(*dataI) + " is not bool")
			}
		}
		*dataI++
		if do {
			done = true
			err := b.template.executeFlat(wr, data, dataI)
			if err != nil {
				return err
			}
		} else {
			*dataI += b.template.NumArgs()
		}
	}

	// else block
	if te.els != nil {
		if !done {
			return te.els.executeFlat(wr, data, dataI)
		}
		*dataI += te.els.NumArgs()
	}
	return nil
}
