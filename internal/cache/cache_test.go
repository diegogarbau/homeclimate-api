package cache_test

import (
	"testing"
	"time"

	"homeclimate-api/internal/cache"
)

func TestCache_SetAndGet(t *testing.T) {
	c := cache.New(5 * time.Minute)

	c.Set("key1", "value1")

	val, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected to find key1")
	}
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}
}

func TestCache_Miss(t *testing.T) {
	c := cache.New(5 * time.Minute)

	_, ok := c.Get("nonexistent")
	if ok {
		t.Fatal("expected miss for nonexistent key")
	}
}

func TestCache_Expiry(t *testing.T) {
	c := cache.New(100 * time.Millisecond)

	c.Set("key1", "value1")

	time.Sleep(150 * time.Millisecond)

	_, ok := c.Get("key1")
	if ok {
		t.Fatal("expected key1 to be expired")
	}
}

func TestCache_Delete(t *testing.T) {
	c := cache.New(5 * time.Minute)

	c.Set("key1", "value1")
	c.Delete("key1")

	_, ok := c.Get("key1")
	if ok {
		t.Fatal("expected key1 to be deleted")
	}
}

func TestCache_Overwrite(t *testing.T) {
	c := cache.New(5 * time.Minute)

	c.Set("key1", "value1")
	c.Set("key1", "value2")

	val, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected to find key1")
	}
	if val != "value2" {
		t.Errorf("expected value2, got %v", val)
	}
}
