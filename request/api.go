package request

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

var (
	errorCode = map[int]bool{
		http.StatusBadRequest: true,
		http.StatusConflict:   true,
		http.StatusForbidden:  true,
		http.StatusBadGateway: true,
	}
)

const (
	// PostRequest linter
	PostRequest = "POST"
	// GetRequest linter
	GetRequest = "GET"
)

// APIResp save request and resp for user
type APIResp struct {
	Request *http.Request // api request
	url     *string
	Body    interface{}
	header  map[string]string
	Resp    chan *http.Response // resp channel
	Kind    string              // get or post
}

// API to call api with timeout
type API struct {
	requestChann chan *APIResp // chann to handle request
	client       *http.Client  // handle call api
	timeout      int           // make request time out
	isClosed     bool          // check close
	ctx          context.Context
	cancel       context.CancelFunc
	mutex        sync.RWMutex
}

// NewAPI setting timeout in second, input = 0 for test timeout
func NewAPI(timeout int) *API {
	transport := &http.Transport{
		// TLSHandshakeTimeout: time.Duration(timeout) * time.Second,
		MaxConnsPerHost:     300,
		MaxIdleConnsPerHost: 300,
		MaxIdleConns:        300,
		IdleConnTimeout:     30 * time.Second,
		// WriteBufferSize:     maxMimumWriteBuffer,
		// ReadBufferSize:      maxMimumReadBuffer,
	}
	client := &http.Client{
		Transport: transport,
	}

	if timeout == 0 {
		client.Timeout = time.Duration(1) * time.Nanosecond
	} else {
		client.Timeout = time.Duration(timeout) * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())
	api := &API{
		client:       client,
		timeout:      timeout,
		requestChann: make(chan *APIResp, 1024),
		isClosed:     false,
		ctx:          ctx,
		cancel:       cancel,
	}

	go api.serve()
	return api
}

// Close linter
func (a *API) Close() {
	if !a.checkClose() {
		a.setClose(true)
		a.cancel()
	}
}

func (a *API) serve() {
	for {
		select {
		case request := <-a.requestChann:
			a.choosing(request)
		case <-a.ctx.Done():
			return
		}
	}
}

func (a *API) choosing(request *APIResp) {
	switch request.Kind {
	case PostRequest:
		go a.makePost(request)
	case GetRequest:
		go a.makeGet(request)
	default:
		return
	}
}

// ReadResponse linter
func (a *API) ReadResponse(resp *http.Response) ([]byte, error) {
	if errorCode[resp.StatusCode] {
		return nil, fmt.Errorf("statusCode is: %v err: %v", resp.StatusCode, resp.Status)
	}

	var err error
	var bytes []byte
	if resp.Body != nil {
		defer resp.Body.Close()
		bytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	}

	return bytes, nil
}

// POST make post request and return resp body
func (a *API) POST(url *string, header map[string]string, body interface{}) chan *http.Response {
	result := make(chan *http.Response, 1)
	request := &APIResp{
		url:    url,
		header: header,
		Body:   body,
		Resp:   result,
		Kind:   PostRequest,
	}
	a.requestChann <- request
	return result
}

func (a *API) makePost(request *APIResp) {
	// encode data
	postBody, err := json.Marshal(request.Body)
	if err != nil {
		request.Resp <- &http.Response{
			StatusCode: http.StatusBadRequest,
			Status:     err.Error(),
		}
		return
	}
	hashBody := bytes.NewBuffer(postBody)
	post, err := a.makeRequest(request.url, PostRequest, hashBody)
	if err != nil {
		request.Resp <- &http.Response{
			StatusCode: http.StatusBadRequest,
			Status:     err.Error(),
		}
		return
	}

	for k, v := range request.header {
		post.Header.Set(k, v)
	}
	request.Request = post
	go a.doRequest(request)
}

// GET make get request and return resp body
func (a *API) GET(url *string, header map[string]string) chan *http.Response {
	result := make(chan *http.Response, 1)
	request := &APIResp{
		url:    url,
		header: header,
		Resp:   result,
		Kind:   GetRequest,
	}
	a.requestChann <- request
	return result
}

func (a *API) makeGet(request *APIResp) {
	get, err := a.makeRequest(request.url, request.Kind, nil)
	if err != nil {
		request.Resp <- &http.Response{
			StatusCode: http.StatusBadRequest,
			Status:     err.Error(),
		}
		return
	}
	for k, v := range request.header {
		get.Header.Set(k, v)
	}
	request.Request = get
	// send to request chann
	go a.doRequest(request)
}

func (a *API) makeRequest(url *string, method string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, *url, body)
}

func (a *API) doRequest(r *APIResp) {
	if r.header == nil {
		r.header = make(map[string]string)
	}
	r.header["Content-Type"] = "application/json"
	resp, err := a.client.Do(r.Request)
	if err != nil {
		r.Resp <- &http.Response{
			StatusCode: http.StatusBadRequest,
			Status:     err.Error(),
		}
	} else {
		r.Resp <- resp
	}
}
