package template

import (
	"errors"
	"io/ioutil"
	"reflect"
	"runtime"
	"strings"
	"unicode"
	"unicode/utf8"
)

// TODO move to other package
// req:
// 	-1: s should be valid not exported identifier
//	 0: s should be valid identifier
// 	 1: s should be valid exported identifier
func validateIdent(s string, req int) bool {
	r, i := utf8.DecodeRuneInString(s)
	switch req {
	case -1:
		if !unicode.IsLower(r) && r != '_' {
			return false
		}
	case 0:
		if !unicode.IsLetter(r) && r != '_' {
			return false
		}
	case 1:
		if !unicode.IsUpper(r) {
			return false
		}
	default:
		panic("unknown requirements")
	}
	for _, r = range s[i:] {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}
	return true
}

// TODO move to other package
func IsValidIdent(s string) bool { return validateIdent(s, 0) }

// TODO move to other package
func IsValidExportedIdent(s string) bool { return validateIdent(s, 1) }

// TODO move to other package
func IsValidNotExportedIdent(s string) bool { return validateIdent(s, -1) }

func parse(s *string, t *Template, endExpected bool, funcs map[string]interface{}) error {
	for len(*s) > 0 {
		// Parse simple string (without leading directive)
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
		if len(*s) == 0 {
			break
		}
		// Parse smth with leading directive
		if i = strings.Index(*s, rightDelim); i == -1 {
			return errors.New("unclosed directive")
		}
		direct := (*s)[len(leftDelim):i]
		*s = (*s)[i+len(rightDelim):]
		switch {
		case strings.HasPrefix(direct, optionalBlockPrefix):
			name := direct[len(optionalBlockPrefix):]
			if !IsValidExportedIdent(name) {
				return errors.New("invalid identifier: '" + name + "'")
			}
			var subT Template
			if err := parse(s, &subT, true, funcs); err != nil {
				return err
			}
			t.tes = append(t.tes, teOptionalBlock{name: name, template: subT})
		case strings.HasPrefix(direct, loopBlockPrefix):
			name := direct[len(loopBlockPrefix):]
			if !IsValidExportedIdent(name) {
				return errors.New("invalid identifier: '" + name + "'")
			}
			var subT Template
			if err := parse(s, &subT, true, funcs); err != nil {
				return err
			}
			t.tes = append(t.tes, teLoopBlock{name: name, template: subT})
		case direct == endOfBlock:
			if !endExpected {
				return errors.New("unexpected end of block")
			}
			return nil
		default: // Possible tePlaceholder
			strs := strings.Split(direct, pipe) // Split always return at least 1 string
			method := false
			if strings.HasSuffix(strs[0], "()") {
				method = true
				strs[0] = strs[0][:len(strs[0])-2]
			}
			if !IsValidExportedIdent(strs[0]) {
				return errors.New("unknown block (may be identifier, but not exported)")
			}
			te := tePlaceholder{name: strs[0], method: method}
			for _, str := range strs[1:] {
				if strings.HasPrefix(str, ".") { // method
					str = str[1:]
					if !IsValidExportedIdent(str) {
						return errors.New("bad method name (may be not exported?)")
					}
					te.funcs = append(te.funcs, FuncMethod(str))
				} else { // function
					f, ok := funcs[str]
					if !ok {
						return errors.New("unknown function '" + str + "'")
					}
					te.funcs = append(te.funcs, FuncSimple{f})
				}
			}
			t.tes = append(t.tes, te)

		}

	}
	return nil
}

func Parse(s string, funcs ...interface{}) (*Template, error) {
	t := &Template{}
	funcsMap := make(map[string]interface{})
	for _, f := range funcs {
		name := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
		funcsMap[name] = f
	}
	return t, parse(&s, t, false, funcsMap)
}

func ParseFile(n string, funcs ...interface{}) (*Template, error) {
	b, err := ioutil.ReadFile(n)
	if err != nil {
		return nil, err
	}
	return Parse(string(b), funcs)
}

func MustParse(s string, funcs ...interface{}) *Template {
	t, err := Parse(s, funcs...)
	if err != nil {
		panic(err)
	}
	return t
}

func MustParseFile(n string, funcs ...interface{}) *Template {
	t, err := ParseFile(n, funcs...)
	if err != nil {
		panic(err)
	}
	return t
}
