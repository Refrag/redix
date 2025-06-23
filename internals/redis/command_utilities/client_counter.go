package commandutilities

import (
	"sync"
	"sync/atomic"
)

var (
	globalConnCounter int64 = 0
	counterMutex      sync.Mutex
)

func IncrementConnCounter() {
	counterMutex.Lock()
	atomic.AddInt64(&globalConnCounter, 1)
	counterMutex.Unlock()
}

func DecrementConnCounter() {
	counterMutex.Lock()
	atomic.AddInt64(&globalConnCounter, -1)
	counterMutex.Unlock()
}

func GetConnCounter() int64 {
	counterMutex.Lock()
	defer counterMutex.Unlock()
	return atomic.LoadInt64(&globalConnCounter)
}
