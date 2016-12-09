package template

import (
	"reflect"
	"testing"
)

func TestParseString(t *testing.T) {
	type testElement struct {
		t string
		r map[string]string
		o []string
	}

	tests := []testElement{
		testElement{`asdf`, map[string]string{"0": "asdf"}, []string{"0"}},               // 0
		testElement{`<? ?>asdf`, map[string]string{"0": "asdf"}, []string{"0"}},          // 1
		testElement{`<? 1  ?>asdf`, map[string]string{"1": "asdf"}, []string{"1"}},       // 2
		testElement{`<?  abc ?>asdf`, map[string]string{"abc": "asdf"}, []string{"abc"}}, // 3
		testElement{`asdf<? 	?>`, map[string]string{"0": "asdf", "1": ""}, []string{"0", "1"}}, // 4

		testElement{``, map[string]string{"0": ""}, []string{"0"}}, // 5

		testElement{`<? 213 ?>s1<? ?>s2`, map[string]string{"213": "s1", "214": "s2"}, []string{"213", "214"}},                          // 6
		testElement{`<? 213 ?>s1<? a ?>s2<? ?>s3`, map[string]string{"213": "s1", "a": "s2", "214": "s3"}, []string{"213", "a", "214"}}, // 7
	}

	for i, test := range tests {
		if r, o := ParseString(test.t); !reflect.DeepEqual(r, test.r) || !reflect.DeepEqual(o, test.o) {
			t.Errorf("%v: Expected: %v Got: %v\n\tExpected order: %v Got: %v", i, test.r, r, test.o, o)
		}
	}
}
