package fibonacci

var memo = make(map[uint]uint64)

func Calculate(n uint) uint64 {
	if n <= 1 {
		return uint64(n)
	}

	if val, exists := memo[n]; exists {
		return val
	}

	result := Calculate(n-1) + Calculate(n-2)

	memo[n] = result

	return result
}
