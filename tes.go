package template

import "io"

const (
	leftDelim                 = "{{"
	rightDelim                = "}}"
	optionalBlockPrefix       = "if "
	optionalElseIfBlockPrefix = "else if "
	optionalElseBlockPrefix   = "else"
	loopBlockPrefix           = "range "
	subBlockPrefix                  = "sub "
	endOfBlock                = "end"
	pipe                      = "|"
)

// Te is a template element
type Te interface {
	Execute(wr io.Writer, data []interface{}) error
}
