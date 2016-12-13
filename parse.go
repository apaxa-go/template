package template

import (
	"errors"
	"io/ioutil"
	"reflect"
	"runtime"
	"strings"
)

func parse(s *string, funcs map[string]interface{}) (t *Template, err error) {
	t = new(Template)
	for len(*s) > 0 {
		parseTEString(s, t)

		if len(*s) == 0 {
			break
		}

		var directive string
		directive, err = getDirective(s)
		if err != nil {
			return
		}

		switch {
		case strings.HasPrefix(directive, optionalBlockPrefix):
			var b teOptionalBlock
			b, err = parseTEOptional(s, funcs)
			if err != nil {
				return
			}
			t.tes = append(t.tes, b)
		case strings.HasPrefix(directive, loopBlockPrefix):
			var b teLoopBlock
			b, err = parseTELoop(s, funcs)
			if err != nil {
				return
			}
			t.tes = append(t.tes, b)
		case strings.HasPrefix(directive, subBlockPrefix):
			var b teSubBlock
			b, err = parseTESub(s, funcs)
			if err != nil {
				return
			}
			t.tes = append(t.tes, b)
		case strings.HasPrefix(directive, optionalElseIfBlockPrefix):
			fallthrough
		case strings.HasPrefix(directive, optionalElseBlockPrefix):
			fallthrough
		case strings.HasPrefix(directive, endOfBlock):
			return
		default: // Possible tePlaceholder
			extractDirective(s) // No need for result as getDirective already do it
			var p tePlaceholder
			p, err = parsePlaceholder(directive, funcs)
			if err != nil {
				return
			}
			t.tes = append(t.tes, p)
		}

	}
	return
}

func Parse(s string, funcs ...interface{}) (*Template, error) {
	funcsMap := NewFuncs()
	for _, f := range funcs {
		name := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
		funcsMap[name] = f
	}
	t, err := parse(&s, funcsMap)
	if err == nil && len(s) != 0 {
		err = errors.New("template parsed partyally, may be unexpected end-like directive?")
	}
	return t, err
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
