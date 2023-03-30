package http

import (
	"regexp"
	"strings"
)

type pathSegment interface {
	match(path string) bool
}

type namedPathSegment struct {
	Name string
}

func (nps *namedPathSegment) match(segment string) bool {
	return nps.Name == segment
}

type wildcardPathSegment struct {
	VariableName string
	Value        string
}

func (wps *wildcardPathSegment) match(segment string) bool {
	wps.Value = segment
	return true
}

type restPathSegment struct {
	Value string
}

func (rps *restPathSegment) match(segment string) bool {
	return true
}

type path []pathSegment

func constructPath(path string) path {
	dynamicRegex, _ := regexp.Compile("^\\{(\\w+)\\}$")

	if path[0] != '/' {
		panic("illegal path format " + path)
	}

	parts := strings.Split(path[1:], "/")
	var segments []pathSegment
	for i, part := range parts {
		if part == ">>" {
			// this can only be the last part of the path
			if i != len(parts)-1 {
				panic("illegal path format " + path)
			}

			segments = append(segments, &restPathSegment{})
		} else if dynamicRegex.MatchString(part) {
			sub := dynamicRegex.FindStringSubmatch(part)
			segments = append(segments, &wildcardPathSegment{
				VariableName: sub[1],
			})
		} else if ok, _ := regexp.MatchString("^\\w+$", part); ok {
			segments = append(segments, &namedPathSegment{
				Name: part,
			})
			continue
		} else {
			panic("illegal path format " + path)
		}
	}

	return segments
}

func (p path) match(path string) (path, bool) {
	if path[0] != '/' {
		return nil, false
	}

	parts := strings.Split(path[1:], "/")
	if len(parts) < len(p) {
		return nil, false
	}

	clone := make([]pathSegment, len(p))
	copy(clone, p)

	for i, part := range parts {
		if i >= len(clone) {
			return nil, false
		}

		if !clone[i].match(part) {
			return nil, false
		}

		// if we have a rest path segment, we're done
		if _, ok := clone[i].(*restPathSegment); ok {
			clone[i].(*restPathSegment).Value = strings.Join(parts[i:], "/")
			return clone, true
		}
	}

	return clone, true
}
