package path

import "testing"

func TestNamedPathSegment(t *testing.T) {
	path := "/foo/bar"
	segments := ConstructPath(path)
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

func TestIllegalPath(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	ConstructPath("foo")
}

func TestRestInMiddle(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	ConstructPath("/foo/bar/>>/baz")
}

func TestWhitespacePath(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	ConstructPath("/foo bar")
}

func TestWildcardPathSegmentString(t *testing.T) {
	path := "/{foo}"
	segments := ConstructPath(path)
	if len(segments) != 1 {
		t.Errorf("Expected 1 segment, got %d", len(segments))
	}
	if !segments[0].match("bar") {
		t.Errorf("Expected bar to match bar")
	}

	if segments[0].(*WildcardPathSegment).Value != "bar" {
		t.Errorf("Expected foo to be bar, got %s", segments[0].(*WildcardPathSegment).Value)
	}
}

func TestRestPathSegment(t *testing.T) {
	path := "/foo/bar/>>"
	segments := ConstructPath(path)
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
	segments := ConstructPath(path)
	if len(segments) != 3 {
		t.Errorf("Expected 3 segments, got %d", len(segments))
	}

	if !segments[0].match("foo") {
		t.Errorf("Expected foo to match foo")
	}

	if !segments[1].match("bar") {
		t.Errorf("Expected bar to match bar")
	}

	if segments[1].(*WildcardPathSegment).Value != "bar" {
		t.Errorf("Expected bar to be bar, got %s", segments[1].(*WildcardPathSegment).Value)
	}

	if !segments[2].match("baz") {
		t.Errorf("Expected baz to match baz")
	}
}

func TestMatchStringWithRestPathSegment(t *testing.T) {
	path := "/foo/bar/>>"
	segments := ConstructPath(path)
	var paths []pathSegment
	var ok bool

	if paths, ok = segments.Match("/foo/bar/baz/qux"); !ok {
		t.Errorf("Expected /foo/bar/baz/qux to match /foo/bar/>>")
	}

	if paths[2].(*RestPathSegment).Value != "baz/qux" {
		t.Errorf("Expected baz/qux to be baz/qux, got %s", paths[2].(*RestPathSegment).Value)
	}
}

func TestMatchStringWithWildcardPathSegment(t *testing.T) {
	path := "/{foo}/bar/{baz}"
	segments := ConstructPath(path)
	var paths []pathSegment
	var ok bool

	if paths, ok = segments.Match("/foo/bar/baz"); !ok {
		t.Errorf("Expected /foo/bar/baz to match /{foo}/bar/{baz}")
	}

	if paths[0].(*WildcardPathSegment).Value != "foo" {
		t.Errorf("Expected foo to be foo, got %s", paths[0].(*WildcardPathSegment).Value)
	}

	if paths[2].(*WildcardPathSegment).Value != "baz" {
		t.Errorf("Expected baz to be baz, got %s", paths[2].(*WildcardPathSegment).Value)
	}
}

func TestMatchStringWithWildcardPathSegmentAndRestPathSegment(t *testing.T) {
	path := "/{foo}/bar/{baz}/>>"
	segments := ConstructPath(path)
	var paths []pathSegment
	var ok bool

	if paths, ok = segments.Match("/x/bar/q/qux/a"); !ok {
		t.Errorf("Expected /x/bar/q/qux/a to match /{foo}/bar/{baz}/>>")
	}

	if paths[0].(*WildcardPathSegment).Value != "x" {
		t.Errorf("Expected x to be x, got %s", paths[0].(*WildcardPathSegment).Value)
	}

	if paths[2].(*WildcardPathSegment).Value != "q" {
		t.Errorf("Expected q to be q, got %s", paths[2].(*WildcardPathSegment).Value)
	}

	if paths[3].(*RestPathSegment).Value != "qux/a" {
		t.Errorf("Expected qux/a to be qux/a, got %s", paths[3].(*RestPathSegment).Value)
	}
}

func TestMatchInvalidString(t *testing.T) {
	path := "/"
	segments := ConstructPath(path)
	var ok bool

	if _, ok = segments.Match("foo/bar/baz"); ok {
		t.Errorf("Expected oo/bar/baz to not match /")
	}
}
func TestMatchTooLongPath(t *testing.T) {
	path := "/foo"
	segments := ConstructPath(path)
	var ok bool

	if _, ok = segments.Match("/foo/bar/baz"); ok {
		t.Errorf("Expected /foo/bar/baz to not match /foo")
	}
}

func TestMatchTooLongPath2(t *testing.T) {
	path := "/foo/bar"
	segments := ConstructPath(path)
	var ok bool

	if _, ok = segments.Match("/foo/bar/baz"); ok {
		t.Errorf("Expected /foo/bar/baz to not match /foo")
	}
}

func TestIndexMatch(t *testing.T) {
	path := "/"
	segments := ConstructPath(path)
	var ok bool

	if _, ok = segments.Match("/"); !ok {
		t.Errorf("Expected / to match /")
	}
}

func TestIndexMatchFail(t *testing.T) {
	path := "/"
	segments := ConstructPath(path)
	var ok bool

	if _, ok = segments.Match("/foo"); ok {
		t.Errorf("Expected /foo to not match /")
	}
}

func TestPathMatchFail(t *testing.T) {
	path := "/foo"
	segments := ConstructPath(path)
	var ok bool

	if _, ok = segments.Match("/bar"); ok {
		t.Errorf("Expected /bar to not match /foo")
	}
}

func TestPathMatchFail2(t *testing.T) {
	path := "/foo"
	segments := ConstructPath(path)
	var ok bool

	if _, ok = segments.Match("/foo/bar"); ok {
		t.Errorf("Expected /foo to not match /foo/bar")
	}
}
