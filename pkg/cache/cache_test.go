package cache

import "testing"

var stringKeyCache = New[string]()

func TestCache(t *testing.T) {
	fooVal := stringKeyCache.Get("foo")

	if fooVal != nil {
		t.Fatalf("expected nil for not setted key, got %v", fooVal)
	}

	stringKeyCache.Set("foo", 123)
	fooVal = stringKeyCache.Get("foo")
	fooValInt, ok := fooVal.(int)

	if !ok {
		t.Fatal("expected int the key, but type assertion failed")
	}

	if fooValInt != 123 {
		t.Fatalf("expected 123 for the 'foo' key, got %v", fooValInt)
	}
}
