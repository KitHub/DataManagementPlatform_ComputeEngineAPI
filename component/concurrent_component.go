package component

import "sync"

type SyncMap[K comparable, V any] struct {
	m sync.Map
}

func (s *SyncMap[K, V]) Store(k K, v V) {
	s.m.Store(k, v)
}

func (s *SyncMap[K, V]) Load(k K) (V, bool) {
	val, ok := s.m.Load(k)
	if !ok {
		var zero V
		return zero, false
	}
	return val.(V), true
}
