package session

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestClear_EmptiesStore(t *testing.T) {
	store := NewInMemoryStore(10)
	store.Store(createTestSession("1"))
	store.Store(createTestSession("2"))

	store.Clear()

	if sessions := store.GetAll(); len(sessions) != 0 {
		t.Errorf("expected 0 sessions after clear, got %d", len(sessions))
	}
	if _, err := store.Get("1"); err == nil {
		t.Error("expected error getting session after clear")
	}
}

func TestClear_ResetsSessionCount(t *testing.T) {
	store := NewInMemoryStore(10)
	store.Store(createTestSession("1"))
	store.Store(createTestSession("2"))

	store.Clear()

	if got := store.SessionCount(); got != 0 {
		t.Errorf("SessionCount after Clear = %d, want 0", got)
	}
}

func TestClear_CountResumesFromZero(t *testing.T) {
	store := NewInMemoryStore(10)
	store.Store(createTestSession("1"))
	store.Store(createTestSession("2"))
	store.Clear()
	store.Store(createTestSession("3"))

	if got := store.SessionCount(); got != 1 {
		t.Errorf("SessionCount after Clear+Store = %d, want 1", got)
	}
}

func TestSessionCount_Increments(t *testing.T) {
	store := NewInMemoryStore(10)

	if got := store.SessionCount(); got != 0 {
		t.Errorf("initial SessionCount = %d, want 0", got)
	}

	store.Store(createTestSession("a"))
	if got := store.SessionCount(); got != 1 {
		t.Errorf("SessionCount after 1 store = %d, want 1", got)
	}

	store.Store(createTestSession("b"))
	if got := store.SessionCount(); got != 2 {
		t.Errorf("SessionCount after 2 stores = %d, want 2", got)
	}
}

func TestSessionCount_EvictionDoesNotStopIncrement(t *testing.T) {
	store := NewInMemoryStore(2)
	store.Store(createTestSession("1"))
	store.Store(createTestSession("2"))
	store.Store(createTestSession("3"))

	if got := store.SessionCount(); got != 3 {
		t.Errorf("SessionCount with eviction = %d, want 3", got)
	}
	if sessions := store.GetAll(); len(sessions) != 2 {
		t.Errorf("store should hold only 2 sessions, got %d", len(sessions))
	}
}

func TestSubscribe_NotifiedOnStore(t *testing.T) {
	store := NewInMemoryStore(10)
	notified := make(chan struct{}, 1)

	store.Subscribe(func() {
		select {
		case notified <- struct{}{}:
		default:
		}
	})

	store.Store(createTestSession("1"))

	select {
	case <-notified:
	case <-time.After(time.Second):
		t.Error("subscriber not notified after Store")
	}
}

func TestSubscribe_NotifiedOnClear(t *testing.T) {
	store := NewInMemoryStore(10)
	store.Store(createTestSession("1"))

	notified := make(chan struct{}, 1)
	store.Subscribe(func() {
		select {
		case notified <- struct{}{}:
		default:
		}
	})

	store.Clear()

	select {
	case <-notified:
	case <-time.After(time.Second):
		t.Error("subscriber not notified after Clear")
	}
}

func TestSubscribe_MultipleSubscribers(t *testing.T) {
	store := NewInMemoryStore(10)
	count := make(chan struct{}, 2)

	store.Subscribe(func() { count <- struct{}{} })
	store.Subscribe(func() { count <- struct{}{} })

	store.Store(createTestSession("1"))

	received := 0
	timeout := time.After(time.Second)
	for received < 2 {
		select {
		case <-count:
			received++
		case <-timeout:
			t.Errorf("expected 2 notifications, got %d", received)
			return
		}
	}
}

func TestConcurrentStoreAndGet(t *testing.T) {
	store := NewInMemoryStore(1000)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			store.Store(createTestSession(fmt.Sprintf("session-%d", i)))
		}(i)
	}

	wg.Wait()

	if sessions := store.GetAll(); len(sessions) != 100 {
		t.Errorf("expected 100 sessions after concurrent stores, got %d", len(sessions))
	}
}

func TestConcurrentStoreAndRead(t *testing.T) {
	store := NewInMemoryStore(1000)
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			store.Store(createTestSession(fmt.Sprintf("write-%d", i)))
		}(i)
		go func() {
			defer wg.Done()
			store.GetAll()
		}()
	}

	wg.Wait()
}
