package ttlcache

import (
	"testing"
	"time"
)

func TestSetSimple(t *testing.T) {
	c, err := NewTTLCache(time.Millisecond * 1000)
	if err != nil {
		t.Fatal(err)
	}

	k := "testkey"
	v := "testval"

	if err := c.Set(k, v); err != nil {
		t.Fatal(err)
	}
}

func TestSetGetSimple(t *testing.T) {
	c, err := NewTTLCache(time.Millisecond * 1000)
	if err != nil {
		t.Fatal(err)
	}

	k := "testkey"
	v := "testval"

	if err := c.Set(k, v); err != nil {
		t.Fatal(err)
	}

	r := c.Get(k)
	if r.(string) != v {
		t.Fatalf("expected value %s; received %s", v, r)
	}
}

func TestSetInvalidate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	ttl := time.Millisecond * 1000

	c, err := NewTTLCache(ttl)
	if err != nil {
		t.Fatal(err)
	}

	k := "testkey"
	v := "testval"

	if err := c.Set(k, v); err != nil {
		t.Fatal(err)
	}

	// wait to check
	time.Sleep(ttl - 250)

	r := c.Get(k)
	if r.(string) != v {
		t.Fatalf("expected value %s; received %s", v, r)
	}

	// wait to invalidate
	time.Sleep(ttl + 250)

	// confirm key is gone
	r = c.Get(k)
	if r != nil {
		t.Fatalf("expected nil value; received %s", r)
	}

}

func TestSetUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	ttl := time.Millisecond * 2000

	c, err := NewTTLCache(ttl)
	if err != nil {
		t.Fatal(err)
	}

	k := "testkey"
	v := "testval"

	if err := c.Set(k, v); err != nil {
		t.Fatal(err)
	}

	r := c.Get(k)
	if r.(string) != v {
		t.Fatalf("expected value %s; received %s", v, r)
	}

	// wait to invalidate
	time.Sleep(time.Millisecond * 1900)

	nv := "newval"

	if err := c.Set(k, nv); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond * 1500)

	r = c.Get(k)
	if r.(string) != nv {
		t.Fatalf("expected value %s; received %s", nv, r)
	}

}
