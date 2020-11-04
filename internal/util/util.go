package util

import (
	"net/url"
	"strings"
)

// ParseCommaSeparatedQuery turns a query string of the form
//    .../?x=a,b,c,d&x=f
// into a map[string]struct{}.
func ParseCommaSeparatedQuery(q url.Values, field string, defaults ...string) map[string]struct{} {
	var ret = make(map[string]struct{})

	objects, ok := q[field]
	if !ok {
		for _, d := range defaults {
			ret[d] = struct{}{}
		}
		return ret
	}
	for _, obj := range objects {
		for _, a := range strings.Split(obj, ",") {
			if a != "" {
				if _, ok := ret[a]; !ok {
					ret[a] = struct{}{}
				}
			}
		}
	}
	return ret
}
