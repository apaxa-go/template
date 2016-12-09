package template

import (
	"github.com/apaxa-go/helper/bytesh"
	"html/template"
	"strings"
	"testing"
)

func TestTemplate_Compile(t *testing.T) {
	type testElement struct {
		template string
		data     interface{}
		funcs    []interface{}
		r        string
		err      bool
	}
	tests := []testElement{
		{"Text", struct{}{}, nil, "Text", false},
		{"Prefix {{Text}} suffix", struct{ Text string }{"text"}, nil, "Prefix text suffix", false},
		{"{{Text}} Prefix {{Text}} suff{{Suffix}}", struct{ Text, Suffix string }{"text", "ix"}, nil, "text Prefix text suffix", false},
		{"{{if Text}}Text{{end}}", struct{ TextOpt bool }{true}, nil, "Text", false},
		{"{{if Text}}Text{{end}}", struct{ TextOpt bool }{false}, nil, "", false},
		{
			"{{if Text}}Text{{Suffix}}{{end}}",
			struct {
				TextOpt bool
				Text    struct {
					Suffix string
				}
			}{
				true, struct{ Suffix string }{Suffix: " suffix"},
			},
			nil,
			"Text suffix",
			false,
		},
		{
			"Prefix {{if Text}}Text{{Suffix}}{{end}}",
			struct {
				TextOpt bool
				Text    struct {
					Suffix string
				}
			}{
				false, struct{ Suffix string }{Suffix: " suffix"},
			},
			nil,
			"Prefix ",
			false,
		},
		{"{{range Points}}.{{end}}", struct{ Points int }{3}, nil, "...", false},
		{
			"{{Prefix}} {{range Texts}} {{Text}} {{end}} {{Suffix}}",
			struct {
				Prefix string
				Texts  []struct{ Text string }
				Suffix string
			}{
				"Prefix",
				[]struct{ Text string }{
					{Text: "1"},
					{Text: "2"},
					{Text: "3"},
				},
				"Suffix",
			},
			nil,
			"Prefix  1  2  3  Suffix",
			false,
		},
		{"{{Prefix|strings.ToLower}}", struct{ Prefix string }{"pREFIX"}, []interface{}{strings.ToLower}, "prefix", false},
		{"{{Prefix|strings.ToLower|html/template.HTMLEscapeString}}", struct{ Prefix string }{"<pREFIX"}, []interface{}{strings.ToLower, template.HTMLEscapeString}, "&lt;prefix", false},
	}
	for _, v := range tests {
		template, err := Parse(v.template, v.funcs...)
		if err != nil {
			t.Error(err)
		}

		if r, err := template.Compile(v.data); err != nil != v.err || r != v.r {
			t.Errorf("%v: expect %v %v, got %v %v", v.template, v.r, v.err, r, err)
		}

		buf := bytesh.NewBuffer(nil)
		if err := template.Execute(buf, v.data); err != nil != v.err || string(buf.Bytes()) != v.r {
			t.Errorf("%v: expect %v %v, got %v %v", v.template, v.r, v.err, string(buf.Bytes()), err)
		}
	}
}
