package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/KantapatSg/backend-interview-lab/platform/internal/domain"
	red "github.com/redis/go-redis/v9"
)

type Cache struct{ client *red.Client }

func New(addr string) *Cache                    { return &Cache{client: red.NewClient(&red.Options{Addr: addr})} }
func (c *Cache) Ping(ctx context.Context) error { return c.client.Ping(ctx).Err() }
func (c *Cache) Close() error                   { return c.client.Close() }

func (c *Cache) Get(ctx context.Context, key string, dst *domain.Job) error {
	value, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, red.Nil) {
		return domain.ErrNotFound
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(value, dst)
}
func (c *Cache) Set(ctx context.Context, key string, value domain.Job, ttl time.Duration) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, payload, ttl).Err()
}
