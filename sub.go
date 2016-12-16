package template

import (
	"errors"
	"io"
)

type teSubBlock struct {
	args      placeholders
	template *Template
}

func (te teSubBlock) NumArgs() int { return 1 }

// *s must begin with if directive
func parseTESub(s *string, funcs map[string]interface{}) (b teSubBlock, err error) {
	directive, _ := extractDirective(s)
	directive = directive[len(subBlockPrefix):] // Only placeholder definition

	b.args, err = parsePlaceholders(directive, funcs)
	if err != nil {
		return
	}

	b.template, err = parse(s, funcs)
	if err != nil {
		return
	}

	// prepare for end block
	directive, err = extractDirective(s)
	if err != nil {
		return
	}

	// End directive
	if directive != endOfBlock {
		err = errors.New("expect end of sub block, but got '" + directive + "'")
	}
	return
}

func (te teSubBlock) Execute(wr io.Writer, data []interface{}) error {
	d, err := te.args.CompileInterfaces(data)
	if err != nil {
		return err
	}

	err = te.template.execute(wr, d)
	if err != nil {
		return err
	}

	return nil
}
