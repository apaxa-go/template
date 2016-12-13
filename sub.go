package template

import (
	"errors"
	"io"
)

type teSubBlock struct {
	arg      tePlaceholder
	template *Template
}

func (te teSubBlock) NumArgs() int { return 1 }

// *s must begin with if directive
func parseTESub(s *string, funcs map[string]interface{}) (b teSubBlock, err error) {
	directive, _ := extractDirective(s)
	directive = directive[len(subBlockPrefix):] // Only placeholder definition

	b.arg, err = parsePlaceholder(directive, funcs)
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

func (te teSubBlock) Execute(wr io.Writer, topData, parentData, data interface{}) error {
	d,err:=te.arg.CompileInterface(topData,parentData,data)
	if err!=nil{
		return err
	}

	err=te.template.execute(wr, topData, data, d)
	if err!=nil{
		return err
	}

	return nil
}

func (te teSubBlock) ExecuteFlat(wr io.Writer, data []interface{}, dataI *int) error {
	if *dataI >= len(data) {
		return errors.New("not enough arguments")
	}

	err := te.template.execute(wr, data, nil,data[*dataI])
	if err != nil {
		return err
	}

	*dataI++

	return nil
}
