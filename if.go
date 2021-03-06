package template

import (
	"errors"
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

func (te teOptionalBlock) Execute(wr io.Writer,  data []interface{}) error {
	for _, b := range te.ifs {
		do, err := b.condition.CompileBool( data)
		if err != nil {
			return err
		}
		if do {
			return b.template.execute(wr, data)
		}
	}
	if te.els != nil {
		return te.els.execute(wr,  data)
	}
	return nil
}
