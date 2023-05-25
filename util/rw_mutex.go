package util

type RwErrMutex interface {
	RLock() error
	RUnlock() error
	Lock() error
	Unlock() error
}
