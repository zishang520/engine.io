package types

import (
	"testing"
)

func TestSet(t *testing.T) {
	set := NewSet("test", "string")

	t.Run("Add", func(t *testing.T) {
		if b := set.Add("Add"); b != true {
			t.Fatalf(`*Set.Add("Add") = %t, want match for %t`, b, true)
		}
	})

	t.Run("Has", func(t *testing.T) {
		if b := set.Has("Add"); b != true {
			t.Fatalf(`*Set.Has("Add") = %t, want match for %t`, b, true)
		}
	})

	t.Run("Len", func(t *testing.T) {
		if l := set.Len(); l != 3 {
			t.Fatalf(`*Set.Len() = %d, want match for %d`, l, 3)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if b := set.Delete("Add"); b != true {
			t.Fatalf(`*Set.Delete("Add") = %t, want match for %t`, b, true)
		}

		if l := set.Len(); l != 2 {
			t.Fatalf(`*Set.Len() = %d, want match for %d`, l, 2)
		}
	})

	t.Run("Keys", func(t *testing.T) {
		if l := len(set.Keys()); l != 2 {
			t.Fatalf(`len(*Set.Keys()) = %d, want match for %d`, l, 2)
		}
	})

	t.Run("All", func(t *testing.T) {
		_tmp := set.All()
		if l := len(_tmp); l != 2 {
			t.Fatalf(`len(*Set.All()) = %d, want match for %d`, l, 2)
		}
		delete(_tmp, "test")
		if b := set.Has("test"); b != true {
			t.Fatalf(`*Set.Has("test") = %t, want match for %t`, b, true)
		}
	})

	t.Run("Clear", func(t *testing.T) {
		set.Clear()
		if l := set.Len(); l != 0 {
			t.Fatalf(`*Set.Len() = %d, want match for %d`, l, 0)
		}
	})
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
