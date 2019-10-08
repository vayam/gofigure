/*
Copyright 2013 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Derived from original: https://github.com/golang/groupcache/tree/master/lru

Changes Made by Naren Venkataraman
Modified the code to handle max size of cache entries instead max count of cache entries
*/

package lru

import (
	"container/list"
	"math"
	"time"
)

// Cache is an LRU cache. It is not safe for concurrent access.
type Cache struct {
	// Maxsize is the sum of cache entry sizes before
	// an item is evicted. Zero means no limit.
	MaxSize int64

	// TTL is the maximum time a single item can remain in cache.
	// If the value is 0, items do not expire.
	TTL time.Duration

	ll    *list.List
	cache map[interface{}]*list.Element
	Size  int64
}

// A Key may be any value that is comparable. See http://golang.org/ref/spec#Comparison_operators
type Key interface{}

type entry struct {
	key     Key
	value   interface{}
	Size    int64
	Expires time.Time
}

func (e *entry) expired(now time.Time) bool {
	return !e.Expires.IsZero() && now.After(e.Expires)
}

func (e *entry) Expired() bool {
	return e.expired(time.Now())
}

// New creates a new Cache.
// If maxSize is zero, the cache has no limit and it's assumed
// that eviction is done by the caller.
func New(maxSize int64) *Cache {
	return &Cache{
		MaxSize: maxSize,
		ll:      list.New(),
		cache:   make(map[interface{}]*list.Element),
	}
}

func (c *Cache) addWithExpiration(key Key, value interface{}, size int64, expires time.Time) bool {
	if c.cache == nil {
		c.cache = make(map[interface{}]*list.Element)
		c.ll = list.New()
	}

	if size < 0 {
		return false
	}

	// Entry by itself is over the max capacity
	if c.MaxSize > 0 && size > c.MaxSize {
		return false
	}

	// Adding the entry would lead to integer overflow
	if c.MaxSize > 0 && math.MaxInt64-c.Size < size {
		return false
	}

	// If item already exists with this key, replace it with a new one
	// using the new value and size
	if ee, ok := c.cache[key]; ok {
		c.removeElement(ee)
	}

	// Add item to cache
	e := &entry{
		key:     key,
		value:   value,
		Size:    size,
		Expires: expires,
	}
	ele := c.ll.PushFront(e)
	c.Size += size
	c.cache[key] = ele

	if c.MaxSize <= 0 {
		return true
	}

	// Remove expired entries
	for c.Size > c.MaxSize {
		if c.RemoveExpired(1) == 0 {
			break
		}
	}

	// Remove old entries
	for c.Size > c.MaxSize {
		c.RemoveOldest()
	}

	return true
}

// Add adds a value to the cache.
func (c *Cache) Add(key Key, value interface{}, size int64) bool {
	var expires time.Time
	if c.TTL > 0 {
		expires = time.Now().Add(c.TTL)
	}

	return c.addWithExpiration(key, value, size, expires)
}

// AddWithExpiration adds a value to the cache and sets its expiration explicitly
func (c *Cache) AddWithExpiration(key Key, value interface{}, size int64, expires time.Time) bool {
	return c.addWithExpiration(key, value, size, expires)
}

// Get looks up a key's value from the cache.
func (c *Cache) Get(key Key) (value interface{}, ok bool) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		kv := ele.Value.(*entry)
		if kv.Expired() {
			return
		}
		c.ll.MoveToFront(ele)
		return kv.value, true
	}
	return
}

// Remove removes the provided key from the cache.
func (c *Cache) Remove(key Key) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

// RemoveOldest removes the oldest item from the cache.
func (c *Cache) RemoveOldest() {
	if c.cache == nil {
		return
	}
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

// RemoveExpired removes expired items from the cache.
// Priority for removal is given to the oldest expired items. The max parameter
// determines the maximum number of items to remove. A value of 0 for max will
// remove all expired items.
// Returns the number of items removed.
func (c *Cache) RemoveExpired(max int) int {
	if c.cache == nil {
		return 0
	}
	removed := 0
	now := time.Now()
	for e := c.ll.Back(); e != nil; {
		kv := e.Value.(*entry)
		prev := e.Prev()
		if kv.expired(now) {
			c.removeElement(e)
			removed++
			if max > 0 && removed == max {
				break
			}
		}
		e = prev
	}
	return removed
}

func (c *Cache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
	c.Size -= kv.Size
}

// Len returns the number of items in the cache.
func (c *Cache) Len() int {
	if c.cache == nil {
		return 0
	}
	return c.ll.Len()
}
