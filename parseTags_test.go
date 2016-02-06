package stats

import (
	"reflect"
	"testing"
)

func TestParseTags(t *testing.T) {
	var tsts = []struct {
		input string
		name  string
		tags  map[string]string
	}{
		{"a", "a", map[string]string{}},
		{"a{foo=bar}", "a", map[string]string{"foo": "bar"}},
		{"a{foo=bar,bax=baz}", "a", map[string]string{"foo": "bar", "bax": "baz"}},
	}
	for i, tst := range tsts {
		n, tgs, err := parseTags(tst.input)
		if err != nil {
			t.Errorf("test %d, returned error %s", i, err)
			continue
		}
		if n != tst.name {
			t.Errorf("test %d, expected name %s, but found %s", i, tst.name, n)
		}
		if !reflect.DeepEqual(tgs, tst.tags) {
			t.Errorf("test %d, expected tags %s, but found %s", i, tst.tags, tgs)
		}
	}
}

func TestJoinTags(t *testing.T) {
	var tsts = []struct {
		input  map[string]string
		output string
	}{
		{map[string]string{"a": "b", "c": "d"}, "a=b,c=d"},
		{map[string]string{"a": "b"}, "a=b"},
		{map[string]string{"a": "b", "host": "foo"}, "host=foo,a=b"},
		{map[string]string{"a": "b", "host": "foo", "r": "q", "c": "d"}, "host=foo,a=b,c=d,r=q"},
	}
	for i, tst := range tsts {
		out := joinTags(tst.input)
		if out != tst.output {
			t.Errorf("test %d: '%s' != '%s'", i, out, tst.output)
		}
	}
}
