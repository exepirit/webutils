package webutils

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/maxence-charriere/go-app/v8/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
)

// Request is a structure of a client request, that sends to server.
type Request struct {
	Result  interface{}
	Body    interface{}
	Method  string
	Url     string
	Headers http.Header

	client *http.Client
	ctx    context.Context
}

// Req returns new Request object with default HTTP client.
func Req() *Request {
	return &Request{
		client: http.DefaultClient,
	}
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
		request.Headers = make(http.Header)
	}
	request.Headers.Set(key, value)
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

	req, err := request.prepareRequest()
	if err != nil {
		return nil, errors.New("creating request failed").Wrap(err)
	}

	res, err := request.client.Do(req)
	if err != nil {
		return nil, errors.New("getting document failed").Wrap(err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return nil, errors.New(res.Status).Tag("url", url)
	}

	if err := request.decodeResult(res); err != nil {
		return nil, errors.New("decode document failed").Wrap(err)
	}

	return request.Result, nil
}

func (request Request) prepareRequest() (*http.Request, error) {
	body, err := request.makeBodyReader()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(request.Context(), request.Method, request.Url, body)
	if err != nil {
		return nil, err
	}

	if request.Headers != nil {
		req.Header = request.Headers
	}

	return req, err
}

func (request Request) makeBodyReader() (io.Reader, error) {
	if request.Body == nil {
		return nil, nil
	}

	if r, ok := request.Body.(io.Reader); ok {
		return r, nil
	}

	switch request.Body.(type) {
	case []byte:
		return bytes.NewReader(request.Body.([]byte)), nil
	default:
		encoded, err := json.Marshal(request.Body)
		return bytes.NewReader(encoded), err
	}
}

func (request *Request) decodeResult(resp *http.Response) error {
	if request.Result == nil {
		return nil
	}

	switch request.Result.(type) {
	case []byte:
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		copy(request.Result.([]byte), data)
		return nil
	default:
		request.SetHeader("Content-Type", "application/json")
		decoder := json.NewDecoder(resp.Body)
		return decoder.Decode(request.Result)
	}
}
