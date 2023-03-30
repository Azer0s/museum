package http

import "testing"

func TestNamedPathSegment(t *testing.T) {
	path := "/foo/bar"
	segments := constructPath(path)
	if len(segments) != 2 {
		t.Errorf("Expected 2 segments, got %d", len(segments))
	}

	if !segments[0].match("foo") {
		t.Errorf("Expected foo to match foo")
	}

	if !segments[1].match("bar") {
		t.Errorf("Expected bar to match bar")
	}
}

func TestWildcardPathSegmentString(t *testing.T) {
	path := "/{foo}"
	segments := constructPath(path)
	if len(segments) != 1 {
		t.Errorf("Expected 1 segment, got %d", len(segments))
	}
	if !segments[0].match("bar") {
		t.Errorf("Expected bar to match bar")
	}

	if segments[0].(*wildcardPathSegment).Value != "bar" {
		t.Errorf("Expected foo to be bar, got %s", segments[0].(*wildcardPathSegment).Value)
	}
}

func TestRestPathSegment(t *testing.T) {
	path := "/foo/bar/>>"
	segments := constructPath(path)
	if len(segments) != 3 {
		t.Errorf("Expected 3 segments, got %d", len(segments))
	}

	if !segments[0].match("foo") {
		t.Errorf("Expected foo to match foo")
	}

	if !segments[1].match("bar") {
		t.Errorf("Expected bar to match bar")
	}

	if !segments[2].match("baz") {
		t.Errorf("Expected baz to match baz")
	}
}

func TestRestPathSegmentWithWildcard(t *testing.T) {
	path := "/foo/{bar}/>>"
	segments := constructPath(path)
	if len(segments) != 3 {
		t.Errorf("Expected 3 segments, got %d", len(segments))
	}

	if !segments[0].match("foo") {
		t.Errorf("Expected foo to match foo")
	}

	if !segments[1].match("bar") {
		t.Errorf("Expected bar to match bar")
	}

	if segments[1].(*wildcardPathSegment).Value != "bar" {
		t.Errorf("Expected bar to be bar, got %s", segments[1].(*wildcardPathSegment).Value)
	}

	if !segments[2].match("baz") {
		t.Errorf("Expected baz to match baz")
	}
}

func TestMatchStringWithRestPathSegment(t *testing.T) {
	path := "/foo/bar/>>"
	segments := constructPath(path)
	var paths []pathSegment
	var ok bool

	if paths, ok = segments.match("/foo/bar/baz/qux"); !ok {
		t.Errorf("Expected /foo/bar/baz/qux to match /foo/bar/>>")
	}

	if paths[2].(*restPathSegment).Value != "baz/qux" {
		t.Errorf("Expected baz/qux to be baz/qux, got %s", paths[2].(*restPathSegment).Value)
	}
}

func TestMatchStringWithWildcardPathSegment(t *testing.T) {
	path := "/{foo}/bar/{baz}"
	segments := constructPath(path)
	var paths []pathSegment
	var ok bool

	if paths, ok = segments.match("/foo/bar/baz"); !ok {
		t.Errorf("Expected /foo/bar/baz to match /{foo}/bar/{baz}")
	}

	if paths[0].(*wildcardPathSegment).Value != "foo" {
		t.Errorf("Expected foo to be foo, got %s", paths[0].(*wildcardPathSegment).Value)
	}

	if paths[2].(*wildcardPathSegment).Value != "baz" {
		t.Errorf("Expected baz to be baz, got %s", paths[2].(*wildcardPathSegment).Value)
	}
}

func TestMatchStringWithWildcardPathSegmentAndRestPathSegment(t *testing.T) {
	path := "/{foo}/bar/{baz}/>>"
	segments := constructPath(path)
	var paths []pathSegment
	var ok bool

	if paths, ok = segments.match("/x/bar/q/qux/a"); !ok {
		t.Errorf("Expected /x/bar/q/qux/a to match /{foo}/bar/{baz}/>>")
	}

	if paths[0].(*wildcardPathSegment).Value != "x" {
		t.Errorf("Expected x to be x, got %s", paths[0].(*wildcardPathSegment).Value)
	}

	if paths[2].(*wildcardPathSegment).Value != "q" {
		t.Errorf("Expected q to be q, got %s", paths[2].(*wildcardPathSegment).Value)
	}

	if paths[3].(*restPathSegment).Value != "qux/a" {
		t.Errorf("Expected qux/a to be qux/a, got %s", paths[3].(*restPathSegment).Value)
	}
}
