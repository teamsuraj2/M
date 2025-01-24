package helpers

import "sync"

func LoadTyped[T any](cache *sync.Map, key string) (T, bool) {
	var zero T
	if val, ok := cache.Load(key); ok {
		if typed, ok := val.(T); ok {
			return typed, true
		}
	}
	return zero, false
}
