/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package easy_http

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/li-jin-gou/http2curl"
)

type Request struct {
	client              *Client
	URL                 string
	Method              string
	QueryParam          url.Values
	FormData            url.Values
	Header              http.Header
	Cookie              []*http.Cookie
	Body                interface{}
	PathParams          map[string]string
	MultipartFormParams map[string]string
	File                map[string]string
	RawRequest          *protocol.Request
	Ctx                 context.Context
	RequestOptions      []config.RequestOption
	Result              interface{}
	Error               error
	isMultiPart         bool
	hasCreate           bool
}

// SetQueryParam method sets single parameter and its value in the current request.
// It will be formed as query string for the request.
//
// For Example: `search=kitchen%20papers&size=large` in the URL after `?` mark.
//
//	client.R().
//		SetQueryParam("search", "kitchen papers").
//		SetQueryParam("size", "large")
//
// Note: it will overwrite the same query key.
func (r *Request) SetQueryParam(param, value string) *Request {
	r.QueryParam.Set(param, value)
	return r
}

// SetQueryParams sets multiple query parameters from a map.
// It iterates over the map and sets each key-value pair.
//
// For Example:
//
//	client.R().
//		SetQueryParams(map[string]string{
//			"search": "kitchen papers",
//			"size": "large",
//		})
//
// Note: it will overwrite existing parameters with the same keys.
func (r *Request) SetQueryParams(params map[string]string) *Request {
	for p, v := range params {
		r.SetQueryParam(p, v)
	}
	return r
}

// SetQueryParamsFromValues sets query parameters from url.Values.
// It iterates over each key-value pair and adds them to the request.
//
// For Example:
//
//	params := url.Values{}
//	params.Add("search", "kitchen papers")
//	params.Add("size", "large")
//	client.R().
//		SetQueryParamsFromValues(params)
//
// Note: it avoids slice case by using 'add'.
func (r *Request) SetQueryParamsFromValues(params url.Values) *Request {
	for p, v := range params {
		for _, pv := range v {
			// use 'add' to avoid slice case
			r.QueryParam.Add(p, pv)
		}
	}
	return r
}

// SetQueryString sets the query string for the request.
// It parses the input query string and sets the query parameters.
//
// For Example:
//
//	r := &Request{}
//	r.SetQueryString("param1=value1&param2=value2")
//
// Note: If parsing fails, it sets the error in the Request.
func (r *Request) SetQueryString(query string) *Request {
	q, err := url.ParseQuery(strings.TrimSpace(query))
	if err != nil {
		r.Error = err
		return r
	}
	r.SetQueryParamsFromValues(q)
	return r
}

// AddQueryParam method adds a query parameter to the current request.
// It accepts two strings: params (key) and value (value).
//
// For Example:
//
//	client.R().
//		AddQueryParam("key1", "value1").
//		AddQueryParam("key2", "value2")
//
// Returns the modified Request object for chaining.
func (r *Request) AddQueryParam(params, value string) *Request {
	r.QueryParam.Add(params, value)
	return r
}

// AddQueryParams adds multiple query parameters to the request.
// It accepts a map of string keys and values.
//
// For Example:
//
//	client.R().
//		AddQueryParams(map[string]string{"search": "kitchen papers", "size": "large"})
//
// Note: it will append to existing query parameters.
func (r *Request) AddQueryParams(params map[string]string) *Request {
	for k, v := range params {
		r.AddQueryParam(k, v)
	}
	return r
}

// AddQueryParamsFromValues method adds multiple query parameters from url.Values to the request.
// It iterates over the provided url.Values and adds each key-value pair to the request's query parameters.
//
// For Example:
//
//	params := url.Values{}
//	params.Add("key1", "value1")
//	params.Add("key2", "value2")
//	client.R().
//		AddQueryParamsFromValues(params)
//
// Note: This method supports chaining.
func (r *Request) AddQueryParamsFromValues(params url.Values) *Request {
	for p, v := range params {
		for _, pv := range v {
			r.QueryParam.Add(p, pv)
		}
	}
	return r
}

// SetPathParam sets a path parameter in the Request object.
// It takes two string arguments: param and value.
//
// Example:
//
//	r := &Request{}
//	r.SetPathParam("id", "123")
//
// This method modifies the Request object and returns it.
func (r *Request) SetPathParam(param, value string) *Request {
	r.PathParams[param] = value
	return r
}

// SetPathParams sets multiple path parameters in the request.
// It iterates over the provided map and sets each parameter.
//
// For Example:
//
//	r := &Request{}
//	r.SetPathParams(map[string]string{
//		"param1": "value1",
//		"param2": "value2",
//	})
//
// Note: This will overwrite existing parameters with the same name.
func (r *Request) SetPathParams(params map[string]string) *Request {
	for p, v := range params {
		r.SetPathParam(p, v)
	}
	return r
}

// SetHeader method sets a header field and its value in the current request.
// It supports chaining for easier method calls.
//
// For Example:
//
//	client.R().
//		SetHeader("Content-Type", "application/json").
//		SetHeader("Authorization", "Bearer token")
//
// Note: it will overwrite the same header key.
func (r *Request) SetHeader(header, value string) *Request {
	r.Header.Set(header, value)
	return r
}

// SetHeaders method sets multiple HTTP headers in the current request.
// It iterates over the provided map and sets each header individually.
//
// For Example:
//
//	client.R().
//		SetHeaders(map[string]string{
//			"Content-Type": "application/json",
//			"Authorization": "Bearer token",
//		})
//
// Note: it will overwrite existing headers with the same key.
func (r *Request) SetHeaders(headers map[string]string) *Request {
	for h, v := range headers {
		r.SetHeader(h, v)
	}
	return r
}

// SetHTTPHeader sets HTTP request headers.
// It accepts http.Header and returns the modified Request.
//
// For Example:
//
//	headers := http.Header{}
//	headers.Add("Content-Type", "application/json")
//	client.R().
//		SetHTTPHeader(headers)
func (r *Request) SetHTTPHeader(header http.Header) *Request {
	r.Header = header
	return r
}

// AddHeader adds a new HTTP header to the Request.
// It accepts header name and value as strings.
//
// For Example:
//
//	r := &Request{}
//	r.AddHeader("Content-Type", "application/json")
//
// Returns the updated Request object for chaining.
func (r *Request) AddHeader(header, value string) *Request {
	r.Header.Add(header, value)
	return r
}

// AddHeaders method adds multiple HTTP headers to the request.
// It iterates over the provided map and adds each header using AddHeader.
//
// For Example:
//
//	headers := map[string]string{
//		"Content-Type": "application/json",
//		"Authorization": "Bearer token",
//	}
//	client.R().
//		AddHeaders(headers)
//
// Note: Each header will be added individually.
func (r *Request) AddHeaders(headers map[string]string) *Request {
	for k, v := range headers {
		r.AddHeader(k, v)
	}
	return r
}

// AddHTTPHeader method adds HTTP headers to the request.
// It accepts a http.Header and appends each key-value pair to the request header.
//
// For Example:
//
//	headers := http.Header{}
//	headers.Add("Content-Type", "application/json")
//	client.R().
//		AddHTTPHeader(headers)
//
// Note: it will append to existing headers.
func (r *Request) AddHTTPHeader(header http.Header) *Request {
	for key, value := range header {
		for _, v := range value {
			r.Header.Add(key, v)
		}
	}
	return r
}

// SetContentType sets the Content-Type header for the HTTP request.
// It accepts a string parameter ct representing the Content-Type value.
//
// For Example:
//
//	client.R().
//		SetContentType("application/json")
//
// Returns the request object itself for chaining.
func (r *Request) SetContentType(ct string) *Request {
	r.Header.Add(consts.HeaderContentType, ct)
	return r
}

// SetContentTypeJSON sets the Content-Type header to application/json.
// It returns the Request instance for chaining.
//
// Example:
//
//	client.R().
//		SetContentTypeJSON()
func (r *Request) SetContentTypeJSON() *Request {
	r.Header.Add(consts.HeaderContentType, consts.MIMEApplicationJSON)
	return r
}

// SetContentTypeFormData sets the HTTP request content type to multipart/form-data.
// It adds the specific content type header to the request header.
//
// For Example:
//
//	req := client.R().
//		SetContentTypeFormData()
//
// Returns the modified request object for chaining.
func (r *Request) SetContentTypeFormData() *Request {
	r.Header.Add(consts.HeaderContentType, consts.MIMEMultipartPOSTForm)
	return r
}

// SetContentTypeUrlEncode sets the Content-Type header to application/x-www-form-urlencoded.
// This is useful for encoding form data as URL parameters.
//
// Example:
//
//	r := client.R().
//		SetContentTypeUrlEncode()
func (r *Request) SetContentTypeUrlEncode() *Request {
	r.Header.Add(consts.HeaderContentType, consts.MIMEApplicationHTMLForm)
	return r
}

// todo 确认 cookie 的具体实现
func (r *Request) SetCookie(hc *http.Cookie) *Request {
	r.Cookie = append(r.Cookie, hc)
	return r
}

// SetCookies adds a slice of HTTP cookies to the Request object.
// It appends the given cookies to the existing list.
//
// For Example:
//
//	cookies := []*http.Cookie{
//		{Name: "cookie1", Value: "value1"},
//		{Name: "cookie2", Value: "value2"},
//	}
//	client.R().SetCookies(cookies)
//
// Note: This method modifies the Request object in place.
func (r *Request) SetCookies(rs []*http.Cookie) *Request {
	r.Cookie = append(r.Cookie, rs...)
	return r
}

// SetBody sets the body of the HTTP request.
// It accepts an interface{} type and returns the request object.
//
// For Example:
//
//	req := &Request{}
//	req.SetBody(map[string]interface{}{"key": "value"})
func (r *Request) SetBody(body interface{}) *Request {
	r.Body = body
	return r
}

// SetFormData sets form data in the request.
// It accepts a map of form field names and values.
//
// For Example:
//
//	req := client.R().
//		SetFormData(map[string]string{
//			"name": "John",
//			"age": "30",
//		})
//
// Note: it will overwrite existing form data fields.
func (r *Request) SetFormData(data map[string]string) *Request {
	for k, v := range data {
		r.FormData.Set(k, v)
	}
	return r
}

// SetFormDataFromValues adds url.Values to the request's FormData.
// It iterates over the provided url.Values and adds each key-value pair.
//
// For Example:
//
//	data := url.Values{}
//	data.Add("key1", "value1")
//	data.Add("key2", "value2")
//	client.R().
//		SetFormDataFromValues(data)
//
// Note: it will append to existing FormData.
func (r *Request) SetFormDataFromValues(data url.Values) *Request {
	for key, value := range data {
		for _, v := range value {
			r.FormData.Add(key, v)
		}
	}
	return r
}

// SetMultipartFormData sets multipart form data for the request.
// It marks the request as multipart and sets the form parameters.
//
// For Example:
//
//	data := map[string]string{
//		"key1": "value1",
//		"key2": "value2",
//	}
//	client.R().
//		SetMultipartFormData(data)
//
// Note: This method supports chaining.
func (r *Request) SetMultipartFormData(data map[string]string) *Request {
	r.isMultiPart = true
	r.MultipartFormParams = data
	return r
}

// SetFile sets a file in the HTTP request.
// It marks the request as multipart and stores the file info.
//
// For Example:
//
//	r := &Request{}
//	r.SetFile("example.txt", "/path/to/example.txt")
//
// Note: This function does not check if the file exists.
func (r *Request) SetFile(filename, filepath string) *Request {
	r.isMultiPart = true
	r.File[filename] = filepath
	return r
}

// SetFiles sets HTTP request files and marks it as multipart form.
// It accepts a map where keys are filenames and values are file paths.
//
// For Example:
//
//	request.SetFiles(map[string]string{
//		"file1": "/path/to/file1",
//		"file2": "/path/to/file2",
//	})
//
// Returns the updated request object.
func (r *Request) SetFiles(files map[string]string) *Request {
	r.isMultiPart = true
	r.File = files
	return r
}

// SetResult sets the result value of the Request object.
// It accepts an interface{} and sets the Result field based on whether the input is a pointer.
//
// For Example:
//
//	req := &Request{}
//	req.SetResult(&MyStruct{}).
//		SetResult("stringValue")
//
// Note: If the input is not a pointer, a new pointer instance is created.
func (r *Request) SetResult(res interface{}) *Request {
	if res != nil {
		vv := reflect.ValueOf(res)
		if vv.Kind() == reflect.Ptr {
			r.Result = res
		} else {
			r.Result = reflect.New(vv.Type()).Interface()
		}
	}
	return r
}

// withContext method attaches a context to the Request instance.
// It accepts a context.Context and assigns it to the Ctx field.
//
// For Example:
//
//	req := &Request{}
//	req.withContext(context.Background())
//
// Note: This method does not handle context cancellation.
func (r *Request) withContext(ctx context.Context) *Request {
	r.Ctx = ctx
	return r
}

// WithDC returns the current request object pointer.
// It does not modify the request.
//
// For Example:
//
//	req := &Request{}
//	req = req.WithDC()
func (r *Request) WithDC() *Request {
	return r
}

// WithCluster method returns the current Request instance.
// It does not modify the receiver.
//
// For Example:
//
//	req := &Request{}
//	req = req.WithCluster()
func (r *Request) WithCluster() *Request {
	return r
}

// WithEnv returns the current Request instance.
// It does not modify the Request object.
//
// For Example:
//
//	req := &Request{}
//	req.WithEnv()
func (r *Request) WithEnv() *Request {
	return r
}

// WithRequestTimeout sets the request timeout duration.
// It modifies the request's timeout settings.
//
// For Example:
//
//	r := &Request{}
//	r.WithRequestTimeout(5 * time.Second)
//
// Note: This method modifies the request instance.
func (r *Request) WithRequestTimeout(t time.Duration) *Request {
	r.RawRequest.SetOptions(config.WithRequestTimeout(t))
	return r
}

// Get method executes an HTTP GET request in the given context.
// It associates the context with the request and then performs the GET request using the specified URL.
//
// For Example:
//
//	resp, err := client.R().
//		Get(ctx, "https://example.com/api/resource")
//
// Returns the response and any error encountered.
func (r *Request) Get(ctx context.Context, url string) (*Response, error) {
	r.withContext(ctx)
	return r.Execute(consts.MethodGet, url)
}

// Head method sends an HTTP HEAD request to the specified URL.
// It attaches the context to the request and executes it.
//
// For Example:
//
//	resp, err := client.R().
//		Head(ctx, "https://example.com")
//
// Note: Returns response and possible error.
func (r *Request) Head(ctx context.Context, url string) (*Response, error) {
	r.withContext(ctx)
	return r.Execute(consts.MethodHead, url)
}

// Post method executes HTTP POST request with given context and URL.
// It associates context with request, then performs POST action.
//
// For Example:
//
//	resp, err := client.R().
//		Post(ctx, "https://example.com/api")
//
// Note: Returns response and possible error.
func (r *Request) Post(ctx context.Context, url string) (*Response, error) {
	r.withContext(ctx)
	return r.Execute(consts.MethodPost, url)
}

// Put method sends a PUT request to the specified URL with the given context.
// It attaches the context to the request and executes it with the PUT method.
//
// For Example:
//
//	resp, err := client.R().
//		Put(ctx, "https://example.com/resource")
//
// Returns the response and any error encountered.
func (r *Request) Put(ctx context.Context, url string) (*Response, error) {
	r.withContext(ctx)
	return r.Execute(consts.MethodPut, url)
}

// Delete method executes an HTTP DELETE request in the given context.
// It attaches the context to the request and then performs the DELETE operation.
//
// For Example:
//
//	resp, err := client.R().
//		Delete(ctx, "https://example.com/resource")
//
// Note: Returns the response and any error encountered.
func (r *Request) Delete(ctx context.Context, url string) (*Response, error) {
	r.withContext(ctx)
	return r.Execute(consts.MethodDelete, url)
}

// Options method executes an HTTP OPTIONS request with the given context and URL.
// It attaches the context to the request and then calls Execute to perform the request.
//
// For Example:
//
//	resp, err := client.R().
//		Options(ctx, "https://example.com")
//
// Note: Returns the response and any error encountered.
func (r *Request) Options(ctx context.Context, url string) (*Response, error) {
	r.withContext(ctx)
	return r.Execute(consts.MethodOptions, url)
}

// Patch method performs a PATCH request to the specified URL with the given context.
// It associates the context with the request and then executes the PATCH request.
//
// For Example:
//
//	resp, err := client.R().
//		Patch(ctx, "https://example.com/api")
//
// Note: Ensure the context is properly managed to avoid goroutine leaks.
func (r *Request) Patch(ctx context.Context, url string) (*Response, error) {
	r.withContext(ctx)
	return r.Execute(consts.MethodPatch, url)
}

// Send sends an HTTP request with the given context.
// It associates the context with the request and executes it.
//
// For Example:
//
//	resp, err := client.R().
//		Send(ctx)
//
// Returns the response or an error if the request fails.
func (r *Request) Send(ctx context.Context) (*Response, error) {
	r.withContext(ctx)
	return r.Execute(r.Method, r.URL)
}

// ToCurl converts an HTTP request to a curl command string.
// It checks if the request has been created and returns an error if not.
//
// For Example:
//
//	req, _ := http.NewRequest("GET", "http://example.com", nil)
//	curlCmd, err := req.ToCurl()
//	if err != nil {
//		log.Fatalf("Error: %s", err)
//	}
//	fmt.Println(curlCmd)
//
// Note: Ensure the request is created before calling this method.
func (r *Request) ToCurl() (string, error) {
	if !r.hasCreate {
		return "", fmt.Errorf("request has not been create")
	}
	c, err := http2curl.GetCurlCommandHertz(r.RawRequest)
	if err != nil {
		return "", err
	}
	return c.String(), nil
}

// Execute method executes an HTTP request with the given method and URL.
// It sets the request method and URL, checks for any existing errors,
// and then calls the client's execute method to perform the request.
//
// For Example:
//
//	resp, err := client.R().
//		Execute("GET", "https://example.com/api")
//
// Returns the response and any error encountered.
func (r *Request) Execute(method, url string) (*Response, error) {
	r.Method = method
	r.URL = url
	if r.Error != nil {
		return nil, r.Error
	}

	return r.client.execute(r)
}
