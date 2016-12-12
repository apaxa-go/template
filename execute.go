package template

import (
	"errors"
	"github.com/apaxa-go/helper/strconvh"
	"io"
)

func (t *Template) execute(wr io.Writer, topData, parentData, data interface{}) error {
	for _, te := range t.tes {
		err := te.Execute(wr, topData, parentData, data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Template) Execute(wr io.Writer, data interface{}) error {
	return t.execute(wr, data, nil, data)
}

func (t *Template) executeFlat(wr io.Writer, data []interface{}, dataI *int) error {
	for _, te := range t.tes {
		err := te.ExecuteFlat(wr, data, dataI)
		if err != nil {
			return err
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
		return errors.New("more than needed arguments: required " + strconvh.FormatInt(dataI) + ", got " + strconvh.FormatInt(len(data)))
	}
	return nil
}
