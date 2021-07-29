package redis

import (
	"math"
	"time"
)

func HashToFloat64(k interface{}) float64 {
	if k == nil {
		return 0
	}
	switch x := k.(type) {
	case string:
		return bytesHash([]byte(x))
	case []byte:
		return bytesHash(x)
	case bool:
		if x {
			return 0
		} else {
			return 1
		}
	case time.Time:
		return math.Float64frombits(uint64(x.UnixNano()))
	case int:
		return math.Float64frombits(uint64(x))
	case int8:
		return math.Float64frombits(uint64(x))
	case int16:
		return math.Float64frombits(uint64(x))
	case int32:
		return math.Float64frombits(uint64(x))
	case int64:
		return math.Float64frombits(uint64(x))
	case uint:
		return math.Float64frombits(uint64(x))
	case uint8:
		return math.Float64frombits(uint64(x))
	case uint16:
		return math.Float64frombits(uint64(x))
	case uint32:
		return math.Float64frombits(uint64(x))
	case uint64:
		return math.Float64frombits(uint64(x))
	case float32:
		return math.Float64frombits(uint64(x))
	case float64:
		return x
	case uintptr:
		return math.Float64frombits(uint64(x))
	}
	panic("unsupported key type.")
}

func bytesHash(bytes []byte) float64 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	keyLength := len(bytes)
	for i := 0; i < keyLength; i++ {
		hash *= prime32
		hash ^= uint32(bytes[i])
	}
	return math.Float64frombits(uint64(hash))
}
