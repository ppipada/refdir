package defaultdirs

// Direct recursion with a base case: function.
func RecursiveFunctionSafe(n int) int {
	if n <= 0 {
		return 0
	}
	return 1 + RecursiveFunctionSafe(n-1)
}

// Generic direct recursion with a base case: function.
func RecursiveGenericSafe[T any](n int, x T) int {
	if n <= 0 {
		return 0
	}
	return 1 + RecursiveGenericSafe[T](n-1, x)
}

// Mutual recursion: only this second call should be flagged (A is defined above).
func MutualA(n int) int {
	if n <= 0 {
		return 0
	}
	return MutualB(n - 1) // OK: call before MutualB's definition
}

func MutualB(n int) int {
	if n <= 0 {
		return 0
	}
	return MutualA(n - 1) // want "func reference MutualA is after definition"
}
