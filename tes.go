package template

import "io"

const (
	leftDelim                 = "{{"
	rightDelim                = "}}"
	optionalBlockPrefix       = "if "
	optionalElseIfBlockPrefix = "else if "
	optionalElseBlockPrefix   = "else"
	loopBlockPrefix           = "range "
	endOfBlock                = "end"
	pipe                      = "|"
)

// Te is a template element
type Te interface {
	NumArgs() int // Number of arguments required for template element. Calculates as for flat compile!!!
	Execute(wr io.Writer, topData, parentData, data interface{}) error
	ExecuteFlat(wr io.Writer, data []interface{}, dataI *int) error
}
