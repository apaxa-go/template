package template

import (
	"io"
	"strings"
)

type teString string

func (te teString) NumArgs() int {
	return 0
}

// Parse simple string (without leading directive)
func parseTEString(s *string, t *Template) {
	i := strings.Index(*s, leftDelim)
	if i != 0 {
		var str string
		if i == -1 {
			i = len(*s)
		}
		str = (*s)[:i]
		*s = (*s)[i:]

		t.tes = append(t.tes, teString(str))
	}
}

func (te teString) Execute(wr io.Writer, topData, parentData, data interface{}) error {
	_, err := wr.Write([]byte(te))
	return err
}

//
//func (te teString) Compile(data interface{}) (string, error) {
//	return string(te), nil
//}
