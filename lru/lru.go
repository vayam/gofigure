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
)

// Cache is an LRU cache. It is not safe for concurrent access.
type Cache struct {
	// Maxsize is the sum of cache entry sizes before
	// an item is evicted. Zero means no limit.
	MaxSize int64

	ll    *list.List
	cache map[interface{}]*list.Element
	Size  int64
}

// A Key may be any value that is comparable. See http://golang.org/ref/spec#Comparison_operators
type Key interface{}

type entry struct {
	key   Key
	value interface{}
	Size  int64
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

// Add adds a value to the cache.
func (c *Cache) Add(key Key, value interface{}, size int64) bool {
	if c.cache == nil {
		c.cache = make(map[interface{}]*list.Element)
		c.ll = list.New()
	}

	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*entry).value = value
		return true
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

	// Add item to cache
	ele := c.ll.PushFront(&entry{key, value, size})
	c.Size += size
	c.cache[key] = ele

	if c.MaxSize <= 0 {
		return true
	}

	//remove old entries
	for c.Size > c.MaxSize {
		c.RemoveOldest()
	}
	return true
}

// Get looks up a key's value from the cache.
func (c *Cache) Get(key Key) (value interface{}, ok bool) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry).value, true
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
