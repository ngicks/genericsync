package genericsync

import (
	"sync"
)

type Map[K, V comparable] struct {
	inner sync.Map
}

func (m *Map[K, V]) CompareAndDelete(key K, old V) (deleted bool) {
	return m.inner.CompareAndDelete(key, old)
}
func (m *Map[K, V]) CompareAndSwap(key K, old, new V) bool {
	return m.inner.CompareAndSwap(key, old, new)
}
func (m *Map[K, V]) Delete(key K) {
	m.inner.Delete(key)
}
func (m *Map[K, V]) Load(key K) (value V, ok bool) {
	if vRaw, ok := m.inner.Load(key); ok {
		return vRaw.(V), true
	} else {
		var zero V
		return zero, false
	}
}
func (m *Map[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	vRaw, loaded := m.inner.LoadAndDelete(key)
	if vRaw != nil {
		value = vRaw.(V)
	}
	return value, loaded
}
func (m *Map[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	vRaw, loaded := m.inner.LoadOrStore(key, value)
	if vRaw != nil {
		actual = vRaw.(V)
	}
	return actual, loaded
}
func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	m.inner.Range(func(key, value any) bool {
		return f(key.(K), value.(V))
	})
}
func (m *Map[K, V]) Store(key K, value V) {
	m.inner.Store(key, value)
}
func (m *Map[K, V]) Swap(key K, value V) (previous V, loaded bool) {
	if vRaw, loaded := m.inner.Swap(key, value); loaded {
		return vRaw.(V), true
	} else {
		var zero V
		return zero, false
	}
}
