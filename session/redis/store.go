package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/coderi421/kyuu/session"
	redis "github.com/redis/go-redis/v9"
	"time"
)

var errSessionNotExist = errors.New("redis-session: session not found")
var errSessionAlreadyExist = errors.New("redis-session: sessionid 已经存在")

// StoreOption is a function type for configuring a Store.
type StoreOption func(store *Store)

type Store struct {
	prefix     string // reids 中 key 的前缀
	client     redis.Cmdable
	expiration time.Duration // 过期时间
}

// NewStore creates a new instance of the Store struct.
// It takes a redis.Cmdable client as the first argument and optional StoreOptions as the rest of the arguments.
// It returns a pointer to the created Store.
func NewStore(client redis.Cmdable, opts ...StoreOption) *Store {
	// Initialize the Store struct with default values.
	res := &Store{
		client:     client,
		prefix:     "session",
		expiration: time.Minute * 15,
	}

	// Apply the optional StoreOptions to modify the Store struct.
	for _, opt := range opts {
		opt(res)
	}

	// Return the created Store.
	return res
}

// WithPrefix is a function that returns a StoreOption function.
// StoreOption is a function that modifies the behavior of the Store type.
// This function sets the prefix value of the store to the provided prefix parameter.
func WithPrefix(prefix string) StoreOption {
	return func(store *Store) {
		store.prefix = prefix
	}
}

// WithExpiration sets the expiration duration for the Store.
func WithExpiration(expiration time.Duration) StoreOption {
	return func(store *Store) {
		store.expiration = expiration
	}
}

// key generates a unique key for the given ID by combining it with the store's prefix.
func (s *Store) key(id string) string {
	return fmt.Sprintf("%s_%s", s.prefix, id)
}

// Generate generates a new session in the store with the given ID.
// If a session with the same ID already exists, it returns an error.
// Otherwise, it creates a new session and returns it.
func (s *Store) Generate(ctx context.Context, id string) (session.Session, error) {
	//	const lua = `
	//redis.call("hset", KEYS[1], ARGV[1], ARGV[2])
	//return redis.call("pexpire", KEYS[1], ARGV[3])
	//`

	// The Lua script checks if a session with the given ID already exists.
	// If it does, it returns -1. Otherwise, it sets the session in Redis and sets an expiration time.
	const lua = `
	if redis.call("exists", KEYS[1])
	then
		return -1
	else
		redis.call("hset", KEYS[1], ARGV[1], ARGV[2])
		return redis.call("pexpire", KEYS[1], ARGV[3])
	end
	`

	// Generate the Redis key for the session ID
	key := s.key(id)

	// Evaluate the Lua script in Redis
	res, err := s.client.Eval(ctx, lua, []string{key}, "_sess_id", id, s.expiration.Milliseconds()).Int()
	if res < 0 {
		return nil, errSessionAlreadyExist
	}
	if err != nil {
		return nil, err
	}

	// Create a new Redis session and return it
	return &redisSession{
		key:    key,
		id:     id,
		client: s.client,
	}, nil
}

// Refresh updates the expiration time of a session in the store.
// It takes a context and the session ID as input and returns an error if any.
func (s *Store) Refresh(ctx context.Context, id string) error {
	// Generate the key for the session ID
	key := s.key(id)

	// Update the expiration time of the session in Redis
	affected, err := s.client.Expire(ctx, key, s.expiration).Result()
	if err != nil {
		return err
	}

	// If the session does not exist, return an error
	if !affected {
		return errSessionNotExist
	}

	return nil
}

// Remove removes an item from the store based on the provided ID.
func (s *Store) Remove(ctx context.Context, id string) error {
	// Delete the item from the store using the client's Del method.
	_, err := s.client.Del(ctx, s.key(id)).Result()
	return err
}

// Get retrieves a session from the store based on the provided ID.
// If the session does not exist, it returns an error.
func (s *Store) Get(ctx context.Context, id string) (session.Session, error) {
	// Generate the key for the session based on its ID.
	key := s.key(id)

	// Check if the session exists in the store.
	i, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	// If the session does not exist, return an error.
	if i < 0 {
		return nil, errSessionNotExist
	}

	// Create a new redisSession instance with the session details.
	return &redisSession{
		key:    key,
		id:     id,
		client: s.client,
	}, nil
}

// redisSession represents a session stored in Redis.
type redisSession struct {
	key    string        // The key under which the session is stored in Redis.
	id     string        // The ID of the session.
	client redis.Cmdable // The Redis client used to interact with Redis.
}

// Get retrieves the value associated with the given key from the Redis session.
// It returns the value as a string and an error, if any.
func (r *redisSession) Get(ctx context.Context, key string) (string, error) {
	// Use the HGet method of the Redis client to get the value associated with the key.
	return r.client.HGet(ctx, r.key, key).Result()
}

// Set sets a value in the Redis session.
func (r *redisSession) Set(ctx context.Context, key string, val string) error {
	// Lua script to check if the session exists and set the value
	const lua = `
if redis.call("exists", KEYS[1])
then
	return redis.call("hset", KEYS[1], ARGV[1], ARGV[2])
else
	return -1
end
`
	// Evaluate the Lua script and get the result
	res, err := r.client.Eval(ctx, lua, []string{r.key}, key, val).Int()
	if err != nil {
		return err
	}
	// Check if the session exists
	if res < 0 {
		return errSessionNotExist
	}
	return nil
}

// ID returns the session ID.
func (r *redisSession) ID() string {
	return r.id
}
