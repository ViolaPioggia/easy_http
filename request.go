package easy_http

import (
	"context"
	"net/http"
	"net/url"

	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/protocol"
)

type Request struct {
	client         *Client
	Url            string
	Method         string
	QueryParam     url.Values
	Header         http.Header
	Cookies        []*http.Cookie
	PathParams     map[string]string
	FormParams     map[string]string
	FileParams     map[string]string
	BodyParams     interface{}
	RawRequest     *protocol.Request
	Ctx            context.Context
	RequestOptions []config.RequestOption
	Result         interface{}
	Error          interface{}
}

const (
	// MethodGet HTTP method
	MethodGet = "GET"

	// MethodPost HTTP method
	MethodPost = "POST"

	// MethodPut HTTP method
	MethodPut = "PUT"

	// MethodDelete HTTP method
	MethodDelete = "DELETE"

	// MethodPatch HTTP method
	MethodPatch = "PATCH"

	// MethodHead HTTP method
	MethodHead = "HEAD"

	// MethodOptions HTTP method
	MethodOptions = "OPTIONS"
)


func (r *Request) Get(url string) (*Response, error) {
	return r.Execute(MethodGet, url)
}

func (r *Request) Head(url string) (*Response, error) {
	return r.Execute(MethodHead, url)
}

func (r *Request) Post(url string) (*Response, error) {
	return r.Execute(MethodPost, url)
}

func (r *Request) Put(url string) (*Response, error) {
	return r.Execute(MethodPut, url)
}

func (r *Request) Delete(url string) (*Response, error) {
	return r.Execute(MethodDelete, url)
}

func (r *Request) Options(url string) (*Response, error) {
	return r.Execute(MethodOptions, url)
}

func (r *Request) Patch(url string) (*Response, error) {
	return r.Execute(MethodPatch, url)
}

func (r *Request) Send() (*Response, error) {
	return r.Execute(r.Method, r.Url)
}

func (r *Request) Execute(method, url string) (*Response, error) {
	r.Method = method
	r.Url = url
	res := &Response{
		request: r,
		rawResponse: &protocol.Response{},
	}
	err :=r.client.client.Do(context.Background(),r.RawRequest,res.rawResponse)
	return res,err	
}