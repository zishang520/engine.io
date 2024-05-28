package types

import (
	"reflect"
	"testing"
)

func TestSet(t *testing.T) {

	s := NewSet[int](1, 2, 3)

	s.Add(4, 5)
	expectedLen := 5
	if len := s.Len(); len != expectedLen {
		t.Errorf("Add method failed, expected length %d, got %d", expectedLen, len)
	}

	s.Delete(3)
	expectedLen = 4
	if len := s.Len(); len != expectedLen {
		t.Errorf("Delete method failed, expected length %d, got %d", expectedLen, len)
	}

	s.Clear()
	if len := s.Len(); len != 0 {
		t.Errorf("Clear method failed, expected length 0, got %d", len)
	}

	s.Add(1, 2, 3)
	tests := []struct {
		key      int
		expected bool
	}{
		{key: 1, expected: true},
		{key: 4, expected: false},
	}
	for _, test := range tests {
		if has := s.Has(test.key); has != test.expected {
			t.Errorf("Has method failed for key %d, expected %t, got %t", test.key, test.expected, has)
		}
	}

	expectedMap := map[int]Void{1: NULL, 2: NULL, 3: NULL}
	if all := s.All(); !reflect.DeepEqual(all, expectedMap) {
		t.Errorf("All method failed, expected %v, got %v", expectedMap, all)
	}
}

func TestMap(t *testing.T) {
	_map := &Map[string, string]{}

	t.Run("Swap", func(t *testing.T) {
		if _, b := _map.Swap("Swap", "Value"); b != false {
			t.Fatalf(`*Map.Swap("Swap") = %t, want match for %t`, b, false)
		}
		if _, b := _map.Swap("Swap", "123"); b != true {
			t.Fatalf(`*Map.Swap("Swap") = %t, want match for %t`, b, true)
		}
		if expunged := _map.dirty["Swap"].expunged; expunged == nil {
			t.Fatalf(`*Map.Swap("Swap").expunged = %v, want to match not nil`, expunged)
		}
		_map.Delete("Swap")
		if n := _map.Len(); n != 0 {
			t.Fatalf(`*Map.Len() = %d, want match for %d`, n, 0)
		}
		if keys := _map.Keys(); len(keys) != 0 {
			t.Fatalf(`*Map.Keys() = %v, want match for []`, keys)
		}
		_map.Store("key", "value")
		if values := _map.Values(); len(values) != 1 {
			t.Fatalf(`*Map.Values() = %v, want match for []`, values)
		}
	})
}
