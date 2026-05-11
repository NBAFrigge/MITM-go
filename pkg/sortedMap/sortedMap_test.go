package sortedMap

import (
	"encoding/json"
	"testing"
)

func TestNew(t *testing.T) {
	sm := New()
	if sm.Len() != 0 {
		t.Errorf("expected empty map, got %d entries", sm.Len())
	}
	if len(sm.Order) != 0 {
		t.Errorf("expected empty order, got %d entries", len(sm.Order))
	}
}

func TestPutAndGet(t *testing.T) {
	sm := New()
	sm.Put("key1", "value1")
	sm.Put("key2", 42)

	val, ok := sm.Get("key1")
	if !ok || val != "value1" {
		t.Errorf("Get(key1) = %v, %v; want value1, true", val, ok)
	}

	val, ok = sm.Get("key2")
	if !ok || val != 42 {
		t.Errorf("Get(key2) = %v, %v; want 42, true", val, ok)
	}

	_, ok = sm.Get("nonexistent")
	if ok {
		t.Error("Get(nonexistent) should return false")
	}
}

func TestPutPreservesInsertionOrder(t *testing.T) {
	sm := New()
	keys := []string{"c", "a", "b", "z", "m"}
	for _, k := range keys {
		sm.Put(k, k)
	}

	order := sm.Keys()
	if len(order) != len(keys) {
		t.Fatalf("expected %d keys, got %d", len(keys), len(order))
	}
	for i, k := range keys {
		if order[i] != k {
			t.Errorf("Order[%d] = %q, want %q", i, order[i], k)
		}
	}
}

func TestPutUpdateDoesNotDuplicateOrder(t *testing.T) {
	sm := New()
	sm.Put("key", "value1")
	sm.Put("key", "value2")

	if sm.Len() != 1 {
		t.Errorf("expected 1 entry after update, got %d", sm.Len())
	}
	if len(sm.Order) != 1 {
		t.Errorf("expected 1 order entry after update, got %d", len(sm.Order))
	}
	val, _ := sm.Get("key")
	if val != "value2" {
		t.Errorf("expected updated value2, got %v", val)
	}
}

func TestDelete(t *testing.T) {
	sm := New()
	sm.Put("a", 1)
	sm.Put("b", 2)
	sm.Put("c", 3)

	sm.Delete("b")

	if sm.Len() != 2 {
		t.Errorf("expected 2 entries after delete, got %d", sm.Len())
	}
	if _, ok := sm.Get("b"); ok {
		t.Error("deleted key should not be retrievable")
	}
	for _, k := range sm.Order {
		if k == "b" {
			t.Error("deleted key should not be in Order")
		}
	}
}

func TestDeleteNonExistent(t *testing.T) {
	sm := New()
	sm.Put("a", 1)
	sm.Delete("nonexistent")
	if sm.Len() != 1 {
		t.Errorf("delete of nonexistent key should not change len")
	}
}

func TestEqual_Same(t *testing.T) {
	sm1 := New()
	sm1.Put("a", "1")
	sm1.Put("b", "2")

	sm2 := New()
	sm2.Put("a", "1")
	sm2.Put("b", "2")

	if !sm1.Equal(sm2) {
		t.Error("expected equal maps to be equal")
	}
}

func TestEqual_DifferentOrder(t *testing.T) {
	sm1 := New()
	sm1.Put("a", "1")
	sm1.Put("b", "2")

	sm2 := New()
	sm2.Put("b", "2")
	sm2.Put("a", "1")

	if sm1.Equal(sm2) {
		t.Error("maps with different insertion order should not be equal")
	}
}

func TestEqual_DifferentLength(t *testing.T) {
	sm1 := New()
	sm1.Put("a", "1")
	sm1.Put("b", "2")

	sm2 := New()
	sm2.Put("a", "1")

	if sm1.Equal(sm2) {
		t.Error("maps with different lengths should not be equal")
	}
}

func TestEqual_DifferentValues(t *testing.T) {
	sm1 := New()
	sm1.Put("a", "1")

	sm2 := New()
	sm2.Put("a", "2")

	if sm1.Equal(sm2) {
		t.Error("maps with different values should not be equal")
	}
}

func TestOrderedValues(t *testing.T) {
	sm := New()
	sm.Put("first", 1)
	sm.Put("second", 2)
	sm.Put("third", 3)

	vals := sm.OrderedValues()
	if len(vals) != 3 {
		t.Fatalf("expected 3 values, got %d", len(vals))
	}
	if vals[0] != 1 || vals[1] != 2 || vals[2] != 3 {
		t.Errorf("OrderedValues = %v, want [1 2 3]", vals)
	}
}

func TestLen(t *testing.T) {
	sm := New()
	if sm.Len() != 0 {
		t.Errorf("expected 0, got %d", sm.Len())
	}
	sm.Put("a", 1)
	if sm.Len() != 1 {
		t.Errorf("expected 1, got %d", sm.Len())
	}
	sm.Delete("a")
	if sm.Len() != 0 {
		t.Errorf("expected 0 after delete, got %d", sm.Len())
	}
}

func TestMarshalUnmarshalJSON(t *testing.T) {
	sm := New()
	sm.Put("name", "alice")
	sm.Put("score", float64(99))

	data, err := json.Marshal(sm)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	sm2 := New()
	if err := json.Unmarshal(data, sm2); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if sm2.Len() != 2 {
		t.Errorf("expected 2 entries after unmarshal, got %d", sm2.Len())
	}
	val, ok := sm2.Get("name")
	if !ok || val != "alice" {
		t.Errorf("Get(name) after unmarshal = %v, %v", val, ok)
	}
}

func TestString(t *testing.T) {
	sm := New()
	sm.Put("key", "value")
	s := sm.String()
	if s == "" {
		t.Error("String() should not be empty for non-empty map")
	}
}
