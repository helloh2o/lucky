package utils

import (
	"sync"
)

// sync Map
type Map struct {
	sync.RWMutex
	m map[interface{}]interface{}
}

func (m *Map) init() {
	if m.m == nil {
		m.m = make(map[interface{}]interface{})
	}
}

// UnsafeGet by key
func (m *Map) UnsafeGet(key interface{}) interface{} {
	if m.m == nil {
		return nil
	} else {
		return m.m[key]
	}
}

// Get by key
func (m *Map) Get(key interface{}) interface{} {
	m.RLock()
	defer m.RUnlock()
	return m.UnsafeGet(key)
}

// UnsafeSet by k v
func (m *Map) UnsafeSet(key interface{}, value interface{}) {
	m.init()
	m.m[key] = value
}

// Set by k v
func (m *Map) Set(key interface{}, value interface{}) {
	m.Lock()
	defer m.Unlock()
	m.UnsafeSet(key, value)
}

// TestAndSet k v v
func (m *Map) TestAndSet(key interface{}, value interface{}) interface{} {
	m.Lock()
	defer m.Unlock()

	m.init()

	if v, ok := m.m[key]; ok {
		return v
	} else {
		m.m[key] = value
		return nil
	}
}

// UnsafeDel by key
func (m *Map) UnsafeDel(key interface{}) {
	m.init()
	delete(m.m, key)
}

// Del by key
func (m *Map) Del(key interface{}) {
	m.Lock()
	defer m.Unlock()
	m.UnsafeDel(key)
}

// UnsafeLen of map
func (m *Map) UnsafeLen() int {
	if m.m == nil {
		return 0
	} else {
		return len(m.m)
	}
}

// Len map
func (m *Map) Len() int {
	m.RLock()
	defer m.RUnlock()
	return m.UnsafeLen()
}

// UnsafeRange map
func (m *Map) UnsafeRange(f func(interface{}, interface{})) {
	if m.m == nil {
		return
	}
	for k, v := range m.m {
		f(k, v)
	}
}

// RLockRange map
func (m *Map) RLockRange(f func(interface{}, interface{})) {
	m.RLock()
	defer m.RUnlock()
	m.UnsafeRange(f)
}

// LockRange map
func (m *Map) LockRange(f func(interface{}, interface{})) {
	m.Lock()
	defer m.Unlock()
	m.UnsafeRange(f)
}
