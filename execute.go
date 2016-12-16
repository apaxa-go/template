package template

import (
	"io"
)

func (t *Template) execute(wr io.Writer, data []interface{}) error {
	for _, te := range t.tes {
		err := te.Execute(wr, data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Template) Execute(wr io.Writer, data ...interface{}) error {
	return t.execute(wr, data)
}
