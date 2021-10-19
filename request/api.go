package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/lamhai1401/gologs/logs"
)

// Result handle request result
type Result struct {
	StatusCode int
	Err        error
	Body       []byte
}

// APIResp save request and resp for user
type APIResp struct {
	Request *http.Request    // api request
	Resp    chan interface{} // resp channel
}

// API to call api with timeout
type API struct {
	requestChann chan *APIResp // chann to handle request
	closeChann   chan int      // handle close
	client       *http.Client  // handle call api
	timeout      int           // make request time out
	isClosed     bool          // check close
	mutex        sync.RWMutex
}

// NewAPI setting timeout in second, input = 0 for test timeout
func NewAPI(timeout int) *API {
	transport := &http.Transport{
		TLSHandshakeTimeout: time.Duration(timeout) * time.Second,
		IdleConnTimeout:     time.Duration(timeout) * time.Second,
		MaxConnsPerHost:     getMaxConnsPerHost(),
		MaxIdleConnsPerHost: getMaxIdleConnsPerHost(),
	}
	client := &http.Client{
		Transport: transport,
	}

	if timeout == 0 {
		client.Timeout = time.Duration(1) * time.Nanosecond
	} else {
		client.Timeout = time.Duration(timeout) * time.Second
	}

	api := &API{
		client:       client,
		timeout:      timeout,
		requestChann: make(chan *APIResp, 1000),
		closeChann:   make(chan int, 1),
		isClosed:     false,
	}

	go api.serve()
	return api
}

// Close linter
func (a *API) Close() {
	a.pushClose()
	a.setClose(true)
}

// POST make post request and return resp body
func (a *API) POST(url *string, header map[string]string, body interface{}) *APIResp {
	resp := &APIResp{
		Resp: make(chan interface{}, 1),
	}
	// encode data
	postBody, err := json.Marshal(body)
	if err != nil {
		resp.Resp <- err
		return resp
	}
	hashBody := bytes.NewBuffer(postBody)
	post, err := a.makeRequest(url, "POST", hashBody)
	if err != nil {
		resp.Resp <- err
		return resp
	}
	for k, v := range header {
		post.Header.Set(k, v)
	}
	resp.Request = post
	a.pushRequest(resp)
	return resp
}

// GET make get request and return resp body
func (a *API) GET(url *string, header map[string]string) *APIResp {
	resp := &APIResp{
		Resp: make(chan interface{}, 1),
	}
	get, err := a.makeRequest(url, "GET", nil)
	if err != nil {
		resp.Resp <- err
		return resp
	}
	for k, v := range header {
		get.Header.Set(k, v)
	}
	resp.Request = get
	// send to request chann
	a.pushRequest(resp)
	return resp
}

func (a *API) serve() {
	for {
		select {
		case req := <-a.getRequestChann():
			go a.handleRequest(req) // for handle async request
		case <-a.getCloseChann():
			logs.Info("Api was closed")
			a.setRequestChann(nil)
			a.client.CloseIdleConnections()
			return
		}
	}
}

func (a *API) handleRequest(r *APIResp) {
	resp, err := a.client.Do(r.Request)
	if err != nil {
		r.Resp <- err
	} else {
		r.Resp <- resp
	}
}

// GetResult get result from request
func (a *API) GetResult(chann *APIResp) *Result {
	result := &Result{}
	resp := <-chann.Resp
	switch v := resp.(type) {
	case error:
		result.Err = v
	case *http.Response:
		defer v.Body.Close()

		// Read the response body
		body, err := ioutil.ReadAll(v.Body)
		if err != nil {
			result.Err = err
		} else {
			result.StatusCode = v.StatusCode
			result.Body = body
		}
	default:
		result.Err = fmt.Errorf("invalid result type, value %v", resp)
	}
	return result
}

func (a *API) makeRequest(url *string, method string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, *url, body)
}
