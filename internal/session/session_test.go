package session

import (
	"httpDebugger/internal/sessiondata"
	"testing"
	"time"
)

func createTestSession(id string) *sessiondata.Session {
	return &sessiondata.Session{
		ID:        id,
		Timestamp: time.Now(),
		Request:   &sessiondata.RequestData{},
	}
}

func TestNewInMemoryStore(t *testing.T) {
	store := NewInMemoryStore(10)
	if store == nil {
		t.Fatal("NewInMemoryStore returned nil")
	}
	if store.maxSize != 10 {
		t.Errorf("maxSize was not set correctly: got %d, want %d", store.maxSize, 10)
	}
}

func TestStoreAndGet(t *testing.T) {
	store := NewInMemoryStore(10)
	session1 := createTestSession("session-1")
	session2 := createTestSession("session-2")

	store.Store(session1)
	store.Store(session2)

	retrievedSession, err := store.Get("session-1")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	if retrievedSession.ID != "session-1" {
		t.Errorf("Get() retrieved wrong session")
	}

	_, err = store.Get("non-existent-id")
	if err == nil {
		t.Errorf("Get() did not return error for non-existent session")
	}
}

func TestGetAll(t *testing.T) {
	store := NewInMemoryStore(10)
	session1 := createTestSession("session-1")
	session2 := createTestSession("session-2")

	store.Store(session1)
	store.Store(session2)

	sessions := store.GetAll()
	if len(sessions) != 2 {
		t.Errorf("GetAll() returned wrong number of sessions: got %d, want %d", len(sessions), 2)
	}
	if sessions[0].ID != "session-1" {
		t.Errorf("GetAll() returned sessions in wrong order. Expected 'session-1', got '%s'", sessions[0].ID)
	}
	if sessions[1].ID != "session-2" {
		t.Errorf("GetAll() returned sessions in wrong order. Expected 'session-2', got '%s'", sessions[1].ID)
	}
}

func TestStoreMaxSize(t *testing.T) {
	maxSize := 3
	store := NewInMemoryStore(maxSize)
	store.Store(createTestSession("1"))
	store.Store(createTestSession("2"))
	store.Store(createTestSession("3"))

	if len(store.GetAll()) != 3 {
		t.Fatalf("Initial store size incorrect: got %d, want %d", len(store.GetAll()), 3)
	}

	store.Store(createTestSession("4"))

	sessions := store.GetAll()
	if len(sessions) != maxSize {
		t.Errorf("Store size is incorrect after adding new session: got %d, want %d", len(sessions), maxSize)
	}

	if sessions[0].ID != "2" {
		t.Errorf("Oldest session was not removed. Oldest is %s, should be '2'", sessions[0].ID)
	}
	if _, err := store.Get("1"); err == nil {
		t.Errorf("Get() for oldest session '1' should have failed")
	}

	if _, err := store.Get("4"); err != nil {
		t.Errorf("Get() for newest session '4' failed")
	}
}
