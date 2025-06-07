package cache

import "testing"

var strToIntCache = New[string, int]()

func TestCache(t *testing.T) {
	fooVal, exists := strToStrKeyCache.Get("foo")

	if exists {
		t.Fatalf("get from empty cache returns something")
	}

	stringKeyCache.Set("foo", 123)
	fooVal, exists = stringKeyCache.Get("foo")

	if !exists {
		t.Fatalf("value for key '%s' do not exist", "foo")
	}

	fooValInt, ok := fooVal.(int)

	if !ok {
		t.Fatal("assertion to int failed for value %v", fooVal)
	}

	if fooValInt != 123 {
		t.Fatalf("got %v for the 'foo' key, expected 123", fooValInt)
	}
}
