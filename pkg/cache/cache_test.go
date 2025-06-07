package cache

import "testing"

var strToIntCache = New[string, int]()

func TestCache(t *testing.T) {
	fooVal := strToStrKeyCache.Get("foo")

	if fooVal != nil {
		t.Fatalf("got nil for not setted key, expected %v", fooVal)
	}

	stringKeyCache.Set("foo", 123)
	fooVal = stringKeyCache.Get("foo")
	fooValInt, ok := fooVal.(int)

	if !ok {
		t.Fatal("assertion to int failed for value %v", fooVal)
	}

	if fooValInt != 123 {
		t.Fatalf("got %v for the 'foo' key, expected 123", fooValInt)
	}
}
