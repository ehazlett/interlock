# TTLCache
TTLCache is a very simple TTL based in-memory cache written in Go.

[![Build Status](https://travis-ci.org/ehazlett/ttlcache.svg?branch=master)](https://travis-ci.org/ehazlett/ttlcache) [![GoDoc](https://godoc.org/github.com/ehazlett/ttlcache?status.svg)](https://godoc.org/github.com/ehazlett/ttlcache) [![Coverage](https://gocover.io/_badge/github.com/ehazlett/ttlcache)](https://gocover.io/_badge/github.com/ehazlett/ttlcache)

# Usage

```go
// error handling omitted for example

c, _ := NewTTLCache(time.Millisecond * 1000)

k := "testkey"
v := "testval"

c.Set(k, v)

r := c.Get(k)
fmt.Println(r.(string))
```

# Expiration Callback
A callback can be specified to be called upon key expiration:

```go
// error handling omitted for example

func callback(k string, v interface{}) {
    fmt.Printf("key %s expired\n", k)
}

c, _ := NewTTLCache(time.Millisecond * 1000)
c.SetCallback(callback)

k := "testkey"
v := "testval"

c.Set(k, v)
```
