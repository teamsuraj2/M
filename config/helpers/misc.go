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

const digitMap = "adefjtghkz" // 0=a, 1=d, 2=e, 3=f, 4=j, 5=t, 6=g, 7=h, 8=k, 9=z

func EncodeDigits(num int) string {
	s := strconv.Itoa(num)
	var sb strings.Builder
	for _, ch := range s {
		d := ch - '0'
		if d < 0 || d > 9 {
			// invalid digit in integer string, skip or error
			continue
		}
		sb.WriteByte(digitMap[d])
	}
	return sb.String()
}