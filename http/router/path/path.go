package path

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

type WildcardPathSegment struct {
	VariableName string
	Value        string
}

func (wps *WildcardPathSegment) match(segment string) bool {
	wps.Value = segment
	return true
}

type RestPathSegment struct {
	Value string
}

func (rps *RestPathSegment) match(segment string) bool {
	return true
}

type Path []pathSegment

func ConstructPath(path string) Path {
	dynamicRegex, _ := regexp.Compile("^\\{(\\w+)\\}$")

	if path[0] != '/' {
		panic("illegal path format " + path)
	}

	if path == "/" {
		return []pathSegment{}
	}

	parts := strings.Split(path[1:], "/")
	var segments []pathSegment
	for i, part := range parts {
		if part == ">>" {
			// this can only be the last part of the path
			if i != len(parts)-1 {
				panic("illegal path format " + path)
			}

			segments = append(segments, &RestPathSegment{})
		} else if dynamicRegex.MatchString(part) {
			sub := dynamicRegex.FindStringSubmatch(part)
			segments = append(segments, &WildcardPathSegment{
				VariableName: sub[1],
			})
		} else if ok, _ := regexp.MatchString("^\\S+$", part); ok {
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

func (p Path) Match(path string) (Path, bool) {
	if path[0] != '/' {
		return nil, false
	}

	if len(p) == 0 && path != "/" {
		return nil, false
	} else if len(p) == 0 && path == "/" {
		return p, true
	}

	parts := strings.Split(path[1:], "/")
	if _, ok := p[len(p)-1].(*RestPathSegment); !ok && len(parts) != len(p) {
		return nil, false
	}

	clone := make([]pathSegment, len(p))
	copy(clone, p)

	for i, part := range parts {
		if !clone[i].match(part) {
			return nil, false
		}

		// if we have a rest path segment, we're done
		if _, ok := clone[i].(*RestPathSegment); ok {
			clone[i].(*RestPathSegment).Value = strings.Join(parts[i:], "/")
			return clone, true
		}
	}

	return clone, true
}
