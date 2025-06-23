package commandutilities

import (
	"sync"
	"sync/atomic"
)

var (
	globalConnCounter int64 = 0
	counterMutex      sync.Mutex
)

/*
Increments the connection counter by 1.

This function is thread-safe.
*/
func IncrementConnCounter() {
	counterMutex.Lock()
	atomic.AddInt64(&globalConnCounter, 1)
	counterMutex.Unlock()
}

/*
Decrements the connection counter by 1.

This function is thread-safe.
*/
func DecrementConnCounter() {
	counterMutex.Lock()
	if atomic.LoadInt64(&globalConnCounter) > 0 {
		atomic.AddInt64(&globalConnCounter, -1)
	}
	counterMutex.Unlock()
}

/*
Returns the current connection counter.

This function is thread-safe.
*/
func GetConnCounter() int64 {
	counterMutex.Lock()
	defer counterMutex.Unlock()
	return atomic.LoadInt64(&globalConnCounter)
}

/*
Initializes the connection counter to 0.

This function is thread-safe.
*/
func InitializeCounter() {
	counterMutex.Lock()
	atomic.StoreInt64(&globalConnCounter, 0)
	counterMutex.Unlock()
}

/*
Initializes the connection counter to the specified value.

This function is thread-safe.
*/
func InitializeCounterWithValue(value int64) {
	counterMutex.Lock()
	atomic.StoreInt64(&globalConnCounter, value)
	counterMutex.Unlock()
}
