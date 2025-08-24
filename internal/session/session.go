package session

import (
	"errors"
	"httpDebugger/internal/sessiondata"
	"sync"
)

type InMemoryStore struct {
	sessions    map[string]*sessiondata.Session
	order       []*sessiondata.Session
	mutex       sync.RWMutex
	maxSize     int
	subscribers []func()
	mu          sync.RWMutex
}

func NewInMemoryStore(maxSize int) *InMemoryStore {
	return &InMemoryStore{
		sessions:    make(map[string]*sessiondata.Session),
		order:       make([]*sessiondata.Session, 0),
		maxSize:     maxSize,
		subscribers: make([]func(), 0),
	}
}

func (s *InMemoryStore) Subscribe(callback func()) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.subscribers = append(s.subscribers, callback)
}

func (s *InMemoryStore) notifySubscribers() {
	subscribers := func() []func() {
		s.mutex.RLock()
		defer s.mutex.RUnlock()
		result := make([]func(), len(s.subscribers))
		copy(result, s.subscribers)
		return result
	}()

	for _, callback := range subscribers {
		go callback()
	}
}

func (s *InMemoryStore) Store(session *sessiondata.Session) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.sessions[session.ID] = session
	s.order = append(s.order, session)

	if len(s.order) > s.maxSize {
		oldest := s.order[0]
		delete(s.sessions, oldest.ID)
		s.order = s.order[1:]
	}

	go s.notifySubscribers()
	return nil
}

func (s *InMemoryStore) GetAll() []*sessiondata.Session {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make([]*sessiondata.Session, len(s.order))
	copy(result, s.order)
	return result
}

func (s *InMemoryStore) Get(id string) (*sessiondata.Session, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	session, exists := s.sessions[id]
	if !exists {
		return nil, errors.New("session not found")
	}
	return session, nil
}

func (s *InMemoryStore) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.sessions = make(map[string]*sessiondata.Session)
	s.order = make([]*sessiondata.Session, 0)
	go s.notifySubscribers()
}
