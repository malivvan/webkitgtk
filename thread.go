package webkitgtk

import (
	"github.com/ebitengine/purego"
	"sync"
)

var mainThreadId uint64
var mainThreadFunctionStore = make(map[uint]func())
var mainThreadFunctionStoreLock sync.RWMutex

func isOnMainThread() bool {
	return mainThreadId == lib.g.ThreadSelf()
}

func generateFunctionStoreID() uint {
	startID := 0
	for {
		if _, ok := mainThreadFunctionStore[uint(startID)]; !ok {
			return uint(startID)
		}
		startID++
		if startID == 0 {
			fatal("Too many functions have been dispatched to the main thread")
		}
	}
}

func invokeSync(fn func()) {
	var wg sync.WaitGroup
	wg.Add(1)
	globalApplication.dispatchOnMainThread(func() {
		defer processPanicHandlerRecover()
		fn()
		wg.Done()
	})
	wg.Wait()
}

func processPanicHandlerRecover() {
	h := PanicHandler
	if h == nil {
		return
	}

	if err := recover(); err != nil {
		h(err)
	}
}
func (a *App) dispatchOnMainThread(fn func()) {
	// If we are on the main thread, just call the function
	if isOnMainThread() {
		fn()
		return
	}

	mainThreadFunctionStoreLock.Lock()
	id := generateFunctionStoreID()
	mainThreadFunctionStore[id] = fn
	mainThreadFunctionStoreLock.Unlock()

	// Call platform specific dispatch function
	dispatchOnMainThread(id)
}

func dispatchOnMainThread(id uint) {
	lib.g.IdleAdd(purego.NewCallback(func(ptr) int {
		executeOnMainThread(id)
		return gSourceRemove
	}))
}

func executeOnMainThread(callbackID uint) {
	mainThreadFunctionStoreLock.RLock()
	fn := mainThreadFunctionStore[callbackID]
	if fn == nil {
		fatal("dispatchCallback called with invalid id: %v", callbackID)
	}
	delete(mainThreadFunctionStore, callbackID)
	mainThreadFunctionStoreLock.RUnlock()
	fn()
}
