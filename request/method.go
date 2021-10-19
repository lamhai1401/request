package request

import "github.com/lamhai1401/gologs/logs"

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

func (a *API) getRequestChann() chan *APIResp {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.requestChann
}

func (a *API) setRequestChann(c chan *APIResp) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.requestChann = c
}

func (a *API) getCloseChann() chan int {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.closeChann
}

func (a *API) pushRequest(r *APIResp) {
	if chann := a.getRequestChann(); chann != nil && !a.checkClose() {
		chann <- r
	} else {
		logs.Info("Cannot push request. API was closed")
	}
}

func (a *API) pushClose() {
	if chann := a.getCloseChann(); chann != nil && !a.checkClose() {
		chann <- 1
	}
}
