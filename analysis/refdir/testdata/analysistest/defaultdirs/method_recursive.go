package defaultdirs

// Value receiver recursion with a base case: method.
type FooValue struct{}

func (f FooValue) Bar(n int) int {
	if n <= 0 {
		return 0
	}
	return 1 + f.Bar(n-1)
}

// Pointer receiver recursion with a base case: method.
type FooPtr struct{}

func (f *FooPtr) Bar(n int) int {
	if f == nil || n <= 0 {
		return 0
	}
	return 1 + f.Bar(n-1)
}

// Generic receiver type recursion with a base case: method.
type Box[T any] struct{}

func (b Box[T]) Beat(n int) int {
	if n <= 0 {
		return 0
	}
	return 1 + b.Beat(n-1)
}

// Method value recursion with a base case: ensure we also don't flag taking method values.
type RecMethodValue struct{}

func (r RecMethodValue) MethodVal(n int) int {
	f := r.MethodVal
	if n <= 0 {
		return 0
	}
	return 1 + f(n-1)
}
