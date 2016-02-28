package ttlcache

import (
	"time"
)

func (t *TTLCache) Set(key string, val interface{}) error {
	k := &Key{
		Value: val,
	}
	// update key
	k.updated = time.Now()

	t.data[key] = k

	return nil
}
