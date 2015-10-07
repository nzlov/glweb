package glweb

import (
	"fmt"
	"regexp"
)

type route struct {
	method  string
	regex   *regexp.Regexp
	g       *glf
	pattern string
}

var routeReg1 = regexp.MustCompile(`:[^/#?()\.\\]+`)
var routeReg2 = regexp.MustCompile(`\*\*`)

func newRoute(g *glf, method, action string) *route {
	route := route{method, nil, g, action}
	pattern := action
	pattern = routeReg1.ReplaceAllStringFunc(pattern, func(m string) string {
		return fmt.Sprintf(`(?P<%s>[^/#?]+)`, m[1:])
	})
	var index int
	pattern = routeReg2.ReplaceAllStringFunc(pattern, func(m string) string {
		index++
		return fmt.Sprintf(`(?P<_%d>[^#?]*)`, index)
	})
	pattern += `\/?`
	route.regex = regexp.MustCompile(pattern)
	return &route
}

type RouteMatch int

const (
	NoMatch RouteMatch = iota
	StarMatch
	OverloadMatch
	ExactMatch
)

//Higher number = better match
func (r RouteMatch) BetterThan(o RouteMatch) bool {
	return r > o
}

func (r route) MatchMethod(method string) RouteMatch {
	switch {
	case method == r.method:
		return ExactMatch
	case method == "HEAD" && r.method == "GET":
		return OverloadMatch
	case r.method == "*":
		return StarMatch
	default:
		return NoMatch
	}
}

func (r route) Match(method string, path string) (RouteMatch, map[string]string) {
	// add Any method matching support
	match := r.MatchMethod(method)
	if match == NoMatch {
		return match, nil
	}

	matches := r.regex.FindStringSubmatch(path)
	if len(matches) > 0 && matches[0] == path {
		params := make(map[string]string)
		for i, name := range r.regex.SubexpNames() {
			if len(name) > 0 {
				params[name] = matches[i]
			}
		}
		return match, params
	}
	return NoMatch, nil
}

var urlReg = regexp.MustCompile(`:[^/#?()\.\\]+|\(\?P<[a-zA-Z0-9]+>.*\)`)

// URLWith returns the url pattern replacing the parameters for its values
func (r *route) URLWith(args []string) string {
	if len(args) > 0 {
		argCount := len(args)
		i := 0
		url := urlReg.ReplaceAllStringFunc(r.pattern, func(m string) string {
			var val interface{}
			if i < argCount {
				val = args[i]
			} else {
				val = m
			}
			i += 1
			return fmt.Sprintf(`%v`, val)
		})

		return url
	}
	return r.pattern
}

func (r *route) Pattern() string {
	return r.pattern
}

func (r *route) Method() string {
	return r.method
}
