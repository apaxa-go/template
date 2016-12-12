package template

type Template struct {
	tes []Te // template elements
}

func (t *Template) NumArgs() int {
	var r int
	for _, te := range t.tes {
		r += te.NumArgs()
	}
	return r
}
