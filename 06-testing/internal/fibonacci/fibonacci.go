package fibonacci

import "sync"

var (
	memo      = make(map[uint]uint64)
	memoMutex sync.Mutex
)

func Calculate(n uint) uint64 {
	if n <= 1 {
		return uint64(n)
	}

	memoMutex.Lock()
	if val, exists := memo[n]; exists {
		return val
	}
	memoMutex.Unlock()

	result := Calculate(n-1) + Calculate(n-2)

	memoMutex.Lock()
	memo[n] = result
	memoMutex.Unlock()

	return result
}
