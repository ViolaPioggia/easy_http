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

func (r *Request) SetQueryParam(param, value string) *Request {
	r.QueryParam.Set(param, value)
	return r
}

func (r *Request) SetQueryParams(params map[string]string) *Request {
	for p, v := range params {
		r.SetQueryParam(p, v)
	}
	return r
}

func (r *Request) SetQueryParamsFromValues(params url.Values) *Request {
	for p, v := range params {
		for _, pv := range v {
			// use 'add' to avoid slice case
			r.QueryParam.Add(p, pv)
		}
	}
	return r
}

func (r *Request) SetQueryString(query string) *Request {
	q, err := url.ParseQuery(strings.TrimSpace(query))
	if err != nil {
		r.Error = err
		return r
	}
	r.SetQueryParamsFromValues(q)
	return r
}

func (r *Request) AddQueryParam(params, value string) *Request {
	r.QueryParam.Add(params, value)
	return r
}

func (r *Request) AddQueryParams(params map[string]string) *Request {
	for k, v := range params {
		r.AddQueryParam(k, v)
	}
	return r
}

func (r *Request) AddQueryParamsFromValues(params url.Values) *Request {
	for p, v := range params {
		for _, pv := range v {
			r.QueryParam.Add(p, pv)
		}
	}
	return r
}

func (r *Request) SetPathParam(param, value string) *Request {
	r.PathParams[param] = value
	return r
}

func (r *Request) SetPathParams(params map[string]string) *Request {
	for p, v := range params {
		r.SetPathParam(p, v)
	}
	return r
}

func (r *Request) SetHeader(header, value string) *Request {
	r.Header.Set(header, value)
	return r
}

func (r *Request) SetHeaders(headers map[string]string) *Request {
	for h, v := range headers {
		r.SetHeader(h, v)
	}
	return r
}

func (r *Request) SetHTTPHeader(header http.Header) *Request {
	r.Header = header
	return r
}

func (r *Request) AddHeader(header, value string) *Request {
	r.Header.Add(header, value)
	return r
}

func (r *Request) AddHeaders(headers map[string]string) *Request {
	for k, v := range headers {
		r.AddHeader(k, v)
	}
	return r
}

func (r *Request) AddHTTPHeader(header http.Header) *Request {
	for key, value := range header {
		for _, v := range value {
			r.Header.Add(key, v)
		}
	}
	return r
}

func (r *Request) SetContentType(ct string) *Request {
	r.Header.Add(consts.HeaderContentType, ct)
	return r
}

func (r *Request) SetContentTypeJSON() *Request {
	r.Header.Add(consts.HeaderContentType, consts.MIMEApplicationJSON)
	return r
}

func (r *Request) SetContentTypeFormData() *Request {
	r.Header.Add(consts.HeaderContentType, consts.MIMEMultipartPOSTForm)
	return r
}

func (r *Request) SetContentTypeUrlEncode() *Request {
	r.Header.Add(consts.HeaderContentType, consts.MIMEApplicationHTMLForm)
	return r
}

// todo 确认 cookie 的具体实现
func (r *Request) SetCookie(hc *http.Cookie) *Request {
	r.Cookie = append(r.Cookie, hc)
	return r
}

func (r *Request) SetCookies(rs []*http.Cookie) *Request {
	r.Cookie = append(r.Cookie, rs...)
	return r
}

func (r *Request) SetBody(body interface{}) *Request {
	r.Body = body
	return r
}

func (r *Request) SetFormData(data map[string]string) *Request {
	for k, v := range data {
		r.FormData.Set(k, v)
	}
	return r
}

func (r *Request) SetFormDataFromValues(data url.Values) *Request {
	for key, value := range data {
		for _, v := range value {
			r.FormData.Add(key, v)
		}
	}
	return r
}

func (r *Request) SetMultipartFormData(data map[string]string) *Request {
	r.isMultiPart = true
	r.MultipartFormParams = data
	return r
}

func (r *Request) SetFile(filename, filepath string) *Request {
	r.isMultiPart = true
	r.File[filename] = filepath
	return r
}

func (r *Request) SetFiles(files map[string]string) *Request {
	r.isMultiPart = true
	r.File = files
	return r
}

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

func (r *Request) withContext(ctx context.Context) *Request {
	r.Ctx = ctx
	return r
}

func (r *Request) WithDC() *Request {
	return r
}

func (r *Request) WithCluster() *Request {
	return r
}

func (r *Request) WithEnv() *Request {
	return r
}

func (r *Request) WithRequestTimeout(t time.Duration) *Request {
	r.RawRequest.SetOptions(config.WithRequestTimeout(t))
	return r
}

func (r *Request) Get(ctx context.Context, url string) (*Response, error) {
	r.withContext(ctx)
	return r.Execute(consts.MethodGet, url)
}

func (r *Request) Head(ctx context.Context, url string) (*Response, error) {
	r.withContext(ctx)
	return r.Execute(consts.MethodHead, url)
}

func (r *Request) Post(ctx context.Context, url string) (*Response, error) {
	r.withContext(ctx)
	return r.Execute(consts.MethodPost, url)
}

func (r *Request) Put(ctx context.Context, url string) (*Response, error) {
	r.withContext(ctx)
	return r.Execute(consts.MethodPut, url)
}

func (r *Request) Delete(ctx context.Context, url string) (*Response, error) {
	r.withContext(ctx)
	return r.Execute(consts.MethodDelete, url)
}

func (r *Request) Options(ctx context.Context, url string) (*Response, error) {
	r.withContext(ctx)
	return r.Execute(consts.MethodOptions, url)
}

func (r *Request) Patch(ctx context.Context, url string) (*Response, error) {
	r.withContext(ctx)
	return r.Execute(consts.MethodPatch, url)
}

func (r *Request) Send(ctx context.Context) (*Response, error) {
	r.withContext(ctx)
	return r.Execute(r.Method, r.URL)
}

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

func (r *Request) Execute(method, url string) (*Response, error) {
	r.Method = method
	r.URL = url
	if r.Error != nil {
		return nil, r.Error
	}

	return r.client.execute(r)
}
