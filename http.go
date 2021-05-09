// +build js

package webutils

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"syscall/js"

	"github.com/maxence-charriere/go-app/v8/pkg/errors"
)

// Request is a structure of a client request, that sends to server.
type Request struct {
	Result  interface{}
	Body    interface{}
	Method  string
	Url     string
	Headers map[string]string

	ctx context.Context
}

// Req returns new Request object with default HTTP client.
func Req() *Request {
	return &Request{}
}

// Context returns current request context or create new one.
func (request *Request) Context() context.Context {
	if request.ctx == nil {
		request.ctx = context.Background()
	}
	return request.ctx
}

// SetResult method register response object for unmarshalling response data.
// Unmarshalling disabled if response type is `[]byte`, otherwise object fields
// will be unmarshalled from JSON.
//
// Result object must be pointer. Example: `r.SetResult(&UserData{})`.
func (request *Request) SetResult(v interface{}) *Request {
	request.Result = v
	return request
}

// SetContext applies context to request.
func (request *Request) SetContext(ctx context.Context) *Request {
	request.ctx = ctx
	return request
}

// SetHeader set client request header.
//
// This method will create header map if it does not exists.
func (request *Request) SetHeader(key, value string) *Request {
	if request.Headers == nil {
		request.Headers = make(map[string]string)
	}
	request.Headers[key] = value
	return request
}

// SetBody sets request body. Supported any bytes-like or JSON-serializable
// content.
//
// `[]byte` or `io.Reader` will send to server as it is. Slices, structs
// and other serializable objects will marshalled into JSON before sending.
func (request *Request) SetBody(body interface{}) *Request {
	request.Body = body
	return request
}

func (request *Request) Get(url string) (interface{}, error) {
	return request.Execute(http.MethodGet, url)
}

func (request *Request) Execute(method, url string) (interface{}, error) {
	request.Url = url
	request.Method = method

	requestOptions, err := request.prepareRequest()
	if err != nil {
		return nil, errors.New("creating request failed").Wrap(err)
	}

	// TODO: context handling
	resp, err := jsFetch(request.Url, requestOptions)
	if err != nil {
		return nil, err
	}

	if resp.statusCode >= 400 {
		return nil, errors.New(strconv.Itoa(resp.statusCode)).Tag("url", url)
	}

	if err := request.decodeResult(resp); err != nil {
		return nil, errors.New("decode document failed").Wrap(err)
	}

	return request.Result, nil
}

func (request Request) prepareRequest() (*fetchOptions, error) {
	body, err := request.makeRequestBody()
	if err != nil {
		return nil, err
	}

	opts := &fetchOptions{
		method:  request.Method,
		headers: request.Headers,
		body:    string(body),
	}

	return opts, nil
}

func (request Request) makeRequestBody() ([]byte, error) {
	if request.Body == nil {
		return []byte{}, nil
	}

	if r, ok := request.Body.(io.Reader); ok {
		return ioutil.ReadAll(r)
	}

	switch request.Body.(type) {
	case []byte:
		return request.Body.([]byte), nil
	default:
		request.Headers["Content-Type"] = "application/json"
		return json.Marshal(request.Body)
	}
}

func (request *Request) decodeResult(resp *response) error {
	if request.Result == nil {
		return nil
	}

	switch request.Result.(type) {
	case []byte:
		copy(request.Body.([]byte), []byte(resp.text))
		return nil
	default:
		return json.Unmarshal([]byte(resp.text), request.Result)
	}
}

type response struct {
	text       string
	statusCode int
	headers    map[string]string
}

type fetchOptions struct {
	method      string
	headers     map[string]string
	body        string
	credentials string
}

func (opt fetchOptions) toMap() map[string]interface{} {
	mp := map[string]interface{}{}
	mp["method"] = opt.method

	if len(opt.headers) > 0 {
		mp["headers"] = opt.headers
	}

	if opt.body != "" {
		mp["body"] = opt.body
	}

	if opt.credentials != "" {
		mp["credentials"] = opt.credentials
	}
	return mp
}

func jsFetch(url string, options *fetchOptions) (*response, error) {
	var r response
	var err error
	done := make(chan struct{})

	js.Global().Call("fetch", url, options.toMap()).
		Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			resp := args[0]
			headersContainer := resp.Get("headers").Call("entries")
			for {
				cond := headersContainer.Call("next")
				if cond.Get("done").Bool() {
					break
				}
				header := cond.Get("value")
				key, value := header.Index(0).String(), header.Index(1).String()
				r.headers[key] = value
			}
			r.statusCode = resp.Get("status").Int()

			resp.Call("text").Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				r.text = args[0].String()
				done <- struct{}{}
				return nil
			}))
			return nil
		})).
		Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			message := args[0].Get("message").String()
			err = errors.New(message)
			done <- struct{}{}
			return nil
		}))

	<-done
	return &r, err
}
