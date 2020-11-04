package util

import (
	"net/url"
	"reflect"
	"testing"
)

func TestParseCommaSeparatedQuery(t *testing.T) {
	for _, tt := range []struct {
		name     string
		in       url.Values
		field    string
		defaults []string
		want     map[string]struct{}
	}{
		{
			"fall back to defaults",
			url.Values{"t": []string{"a,b,", "b"}},
			"f",
			[]string{"x", "y", "z"},
			map[string]struct{}{"x": {}, "y": {}, "z": {}},
		},
		{
			"no defaults",
			url.Values{"t": []string{"a,b,", "b"}},
			"f",
			[]string{},
			map[string]struct{}{},
		},
		{
			"repeated values",
			url.Values{"t": []string{"a,b,", "b"}},
			"t",
			nil,
			map[string]struct{}{"a": {}, "b": {}},
		},
		{
			"repeated",
			url.Values{"t": []string{"a,b,", "b", "c"}},
			"t",
			[]string{},
			map[string]struct{}{"a": {}, "b": {}, "c": {}},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseCommaSeparatedQuery(tt.in, tt.field, tt.defaults...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expected %v got %v", tt.want, got)
			}
		})
	}
}
