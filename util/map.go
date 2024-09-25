package util

func Contains[K comparable](m []K, k K) bool {
	for _, v := range m {
		if v == k {
			return true
		}
	}
	return false
}
