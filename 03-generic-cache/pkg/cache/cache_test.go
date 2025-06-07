package cache

import "testing"

var strToIntCache = New[string, int]()

func TestCache(t *testing.T) {
	fooVal, exists := strToIntCache.Get("foo")

	if exists {
		t.Fatalf("get from empty cache returns something")
	}

	strToIntCache.Set("foo", 123)
	fooVal, exists = strToIntCache.Get("foo")

	if !exists {
		t.Fatalf("value for key '%s' do not exist", "foo")
	}

	if fooVal != 123 {
		t.Fatalf("got %v for the 'foo' key, expected 123", fooVal)
	}
}
