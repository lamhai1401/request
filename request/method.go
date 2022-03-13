package request

func (a *API) checkClose() bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.isClosed
}

func (a *API) setClose(state bool) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.isClosed = state
}
