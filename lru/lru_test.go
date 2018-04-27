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
*/

package lru

import (
	"math"
	"testing"
	"time"
)

type simpleStruct struct {
	int
	string
}

type complexStruct struct {
	int
	simpleStruct
}

var getTests = []struct {
	name       string
	keyToAdd   interface{}
	keyToGet   interface{}
	expectedOk bool
}{
	{"string_hit", "myKey", "myKey", true},
	{"string_miss", "myKey", "nonsense", false},
	{"simple_struct_hit", simpleStruct{1, "two"}, simpleStruct{1, "two"}, true},
	{"simeple_struct_miss", simpleStruct{1, "two"}, simpleStruct{0, "noway"}, false},
	{"complex_struct_hit", complexStruct{1, simpleStruct{2, "three"}},
		complexStruct{1, simpleStruct{2, "three"}}, true},
}

func TestGet(t *testing.T) {
	for _, tt := range getTests {
		lru := New(1024)
		lru.Add(tt.keyToAdd, 1234, 1024)
		val, ok := lru.Get(tt.keyToGet)
		if ok != tt.expectedOk {
			t.Fatalf("%s: cache hit = %v; want %v", tt.name, ok, !ok)
		} else if ok && val != 1234 {
			t.Fatalf("%s expected get to return 1234 but got %v", tt.name, val)
		}
	}
}

func TestRemove(t *testing.T) {
	lru := New(1024)
	lru.Add("myKey", 1234, 20)
	if val, ok := lru.Get("myKey"); !ok {
		t.Fatal("TestRemove returned no match")
	} else if val != 1234 {
		t.Fatalf("TestRemove failed.  Expected %d, got %v", 1234, val)
	}

	lru.Remove("myKey")
	if _, ok := lru.Get("myKey"); ok {
		t.Fatal("TestRemove returned a removed entry")
	}
}

func TestRemoveOld(t *testing.T) {
	lru := New(1024)
	lru.Add("myKey", 1234, 20)
	if val, ok := lru.Get("myKey"); !ok {
		t.Fatal("TestRemoveOld returned no match")
	} else if val != 1234 {
		t.Fatalf("TestRemoveOld failed.  Expected %d, got %v", 1234, val)
	}

	lru.Add("myKey1", 1234, 1024)

	if _, ok := lru.Get("myKey"); ok {
		t.Fatal("TestRemoveOld returned a removed entry")
	}
}

func TestAddOverCapacity(t *testing.T) {
	lru := New(1024)
	ok := lru.Add("myKey", 1234, 1025)
	if ok {
		t.Fatal("TestAddOverCapacity returned true")
	}

}

func TestAddOverflow(t *testing.T) {
	lru := New(math.MaxInt64)
	lru.Add("myKey", 1234, math.MaxInt64)
	ok := lru.Add("myKey1", 1234, 1)
	if ok {
		t.Fatal("TestAddOverflow returned true")
	}
}

func TestAddUnlimitedOverflow(t *testing.T) {
	lru := New(0)
	lru.Add("myKey", 1234, math.MaxInt64)
	ok := lru.Add("myKey1", 1234, 1)
	if !ok {
		t.Fatal("TestAddUnlimitedOverflow returned false")
	}
}

func TestGetExpired(t *testing.T) {
	t.Parallel()
	lru := New(math.MaxInt64)
	lru.TTL = 100 * time.Millisecond
	lru.Add("myKey", 1234, 1)
	time.Sleep(200 * time.Millisecond)
	if _, ok := lru.Get("myKey"); ok {
		t.Fatal("TestGetExpired returned an expired item")
	}
}

func TestGetNotExpired(t *testing.T) {
	lru := New(math.MaxInt64)
	lru.TTL = 100 * time.Millisecond
	lru.Add("myKey", 1234, 1)
	if _, ok := lru.Get("myKey"); !ok {
		t.Fatal("TestGetNotExpired did not return a non-expired item")
	}
}

func TestRemoveExpiredExpired(t *testing.T) {
	t.Parallel()
	lru := New(math.MaxInt64)
	lru.TTL = 100 * time.Millisecond
	lru.Add("myKey1", 1234, 5)
	lru.Add("myKey2", 5678, 5)
	len1 := lru.Len()
	size1 := lru.Size
	time.Sleep(200 * time.Millisecond)
	removed := lru.RemoveExpired(0)
	len2 := lru.Len()
	size2 := lru.Size
	if len1 != 2 || size1 != 10 || removed != 2 || len2 != 0 || size2 != 0 {
		t.Fatal("TestRemoveExpired failed to remove all expired items")
	}
}

func TestRemoveExpiredMax(t *testing.T) {
	t.Parallel()
	lru := New(math.MaxInt64)
	lru.TTL = 100 * time.Millisecond
	lru.Add("myKey1", 1234, 5)
	lru.Add("myKey2", 5678, 5)
	len1 := lru.Len()
	size1 := lru.Size
	time.Sleep(200 * time.Millisecond)
	removed := lru.RemoveExpired(1)
	len2 := lru.Len()
	size2 := lru.Size
	if len1 != 2 || size1 != 10 || removed != 1 || len2 != 1 || size2 != 5 {
		t.Fatal("TestRemoveExpiredMax failed to remove exactly 1 expired item")
	}
}

func TestRemoveExpiredNotExpired(t *testing.T) {
	lru := New(math.MaxInt64)
	lru.TTL = 100 * time.Millisecond
	lru.Add("myKey1", 1234, 5)
	lru.Add("myKey2", 5678, 5)
	len1 := lru.Len()
	size1 := lru.Size
	removed := lru.RemoveExpired(0)
	len2 := lru.Len()
	size2 := lru.Size
	if len1 != 2 || size1 != 10 || removed != 0 || len2 != 2 || size2 != 10 {
		t.Fatal("TestRemoveExpiredNotExpired failed to not remove non-expired items")
	}
}
