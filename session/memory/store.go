package memory

import (
	"context"
	"errors"
	"github.com/coderi421/kyuu/session"
	cache "github.com/patrickmn/go-cache"
	"sync"
	"time"
)

type Store struct {
	// 如果难以确保同一个 id 不会被多个 goroutine 来操作，就加上这个
	mutex sync.RWMutex
	c     *cache.Cache
	// 利用一个内存缓存来帮助我们管理过期时间
	expiration time.Duration
}

// NewStore creates a new Store instance.
// The expiration parameter specifies the duration for which the cached values
func NewStore(expiration time.Duration) *Store {
	return &Store{
		c:          cache.New(expiration, time.Second),
		expiration: expiration,
	}
}

// Generate creates a new session with the given ID and returns it.
// It also stores the session in the memory cache with an expiration time.
func (s *Store) Generate(ctx context.Context, id string) (session.Session, error) {
	// Lock the mutex to ensure exclusive access to the Store.
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Create a new memorySession with the given ID.
	sess := &memorySession{
		id:   id,
		data: make(map[string]string),
	}

	// Store the session in the memory cache with the specified expiration time.
	s.c.Set(sess.ID(), sess, s.expiration)

	// Return the created session and nil error.
	return sess, nil
}

// Refresh updates the expiration time of a session in the store.
// It takes a context and an ID as input and returns an error if the session is not found.
func (s *Store) Refresh(ctx context.Context, id string) error {
	// Lock the mutex to prevent concurrent access to the store.
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Retrieve the session from the cache using the provided ID.
	sess, ok := s.c.Get(id)
	if !ok {
		return errors.New("session not found")
	}

	// Update the expiration time of the session in the cache.
	s.c.Set(sess.(*memorySession).ID(), sess, s.expiration)

	return nil
}

// Remove removes an item from the store by its ID.
func (s *Store) Remove(ctx context.Context, id string) error {
	// Acquire the lock to ensure exclusive access to the store.
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Delete the item from the store by its ID.
	s.c.Delete(id)

	// Return nil error to indicate successful removal.
	return nil
}

// Get retrieves a session from the store by its ID.
// It returns the session if found, otherwise it returns an error.
func (s *Store) Get(ctx context.Context, id string) (session.Session, error) {
	// Acquire a read lock to protect concurrent access to the store.
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Check if the session exists in the cache.
	sess, ok := s.c.Get(id)
	if !ok {
		return nil, errors.New("session not found")
	}

	// Return the session cast to memorySession.
	return sess.(*memorySession), nil
}

// memorySession represents a session stored in memory.
type memorySession struct {
	mutex sync.RWMutex
	id    string
	// Data stored in the session
	data map[string]string
}

// Get retrieves the value associated with the given key from the memorySession.
// It returns the value if found, otherwise it returns an error.
func (m *memorySession) Get(ctx context.Context, key string) (string, error) {
	// Lock the mutex to prevent concurrent reads/writes to the data map.
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	val, ok := m.data[key]
	if !ok {
		return "", errors.New("找不到这个 key")
	}

	return val, nil
}

// Set adds or updates a value in the session data.
// It acquires a lock to ensure thread safety and updates the data map with the provided key-value pair.
// It returns an error if any occurs.
func (m *memorySession) Set(ctx context.Context, key string, val string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.data[key] = val
	return nil
}

func (m *memorySession) ID() string {
	return m.id
}
