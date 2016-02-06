package ttlcache

func (t *TTLCache) Get(key string) interface{} {
	if k, ok := t.data[key]; ok {
		return k.Value
	}

	return nil
}
