package ttlcache

import (
	"time"

	log "github.com/Sirupsen/logrus"
)

func (t *TTLCache) reap() {
	t.lock.Lock()
	defer t.lock.Unlock()
	for k, v := range t.data {
		elapsed := time.Since(v.updated)
		if elapsed >= t.ttl {
			log.Debugf("reaping key: %s", k)
			delete(t.data, k)

			// callback
			t.reapCallback(k, v.Value)
		}
	}
}
