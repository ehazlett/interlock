package ttlcache

import (
	"fmt"
	"sync"
	"time"
)

type Key struct {
	updated time.Time
	Value   interface{}
}

type TTLCache struct {
	data         map[string]*Key
	ttl          time.Duration
	lock         *sync.Mutex
	reapCallback func(key string, val interface{})
}

func NewTTLCache(ttl time.Duration) (*TTLCache, error) {
	if ttl <= time.Duration(time.Millisecond*100) {
		return nil, fmt.Errorf("ttl too low")
	}

	c := &TTLCache{
		data:         map[string]*Key{},
		ttl:          ttl,
		lock:         &sync.Mutex{},
		reapCallback: func(key string, val interface{}) {},
	}

	t := time.NewTicker(ttl)
	go func() {
		for range t.C {
			c.reap()
		}
	}()

	return c, nil
}

func (t *TTLCache) SetCallback(f func(k string, v interface{})) {
	t.reapCallback = f
}
