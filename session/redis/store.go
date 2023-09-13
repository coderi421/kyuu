package redis

import (
	"context"
	"github.com/coderi421/kyuu/session"
	redis "github.com/redis/go-redis/v9"
)

type Store struct {
}

func (s *Store) Generate(ctx context.Context, id string) (session.Session, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Store) Refresh(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}

func (s *Store) Remove(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}

func (s *Store) Get(ctx context.Context, id string) (session.Session, error) {
	//TODO implement me
	panic("implement me")
}

// redisSession represents a session stored in Redis.
type redisSession struct {
	key    string        // The key under which the session is stored in Redis.
	id     string        // The ID of the session.
	client redis.Cmdable // The Redis client used to interact with Redis.
}

func (r *redisSession) Get(ctx context.Context, key string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (r *redisSession) Set(ctx context.Context, key string, val string) error {
	//TODO implement me
	panic("implement me")
}

func (r *redisSession) ID() string {
	//TODO implement me
	panic("implement me")
}
