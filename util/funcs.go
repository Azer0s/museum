package util

// IdentityF returns a function that returns the same value that was passed to it
func IdentityF[T any](t T) func() T {
	return func() T {
		return t
	}
}
