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
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/protocol"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type Client struct {
	baseURL string
	header  http.Header

	beforeRequest       []RequestMiddleware
	udBeforeRequest     []RequestMiddleware
	afterResponse       []ResponseMiddleware
	afterResponseLock   *sync.RWMutex
	udBeforeRequestLock *sync.RWMutex

	enableDiscovery bool

	client  *client.Client
	options []config.ClientOption
}

type (
	RequestMiddleware  func(*Client, *Request) error
	ResponseMiddleware func(*Client, *Response) error
)

var (
	hdrContentTypeKey = http.CanonicalHeaderKey("Content-Type")

	plainTextType       = "text/plain; charset=utf-8"
	jsonContentType     = "application/json"
	formContentType     = "application/x-www-form-urlencoded"
	formDataContentType = "multipart/form-data"
)

func createClient(cc *client.Client, opts ...config.ClientOption) *Client {
	c := &Client{
		udBeforeRequestLock: &sync.RWMutex{},
		afterResponseLock:   &sync.RWMutex{},

		client:  cc,
		options: opts,
	}

	c.beforeRequest = []RequestMiddleware{
		parseRequestURL,
		parseRequestHeader,
		parseRequestBody,
	}

	c.udBeforeRequest = []RequestMiddleware{}

	c.afterResponse = []ResponseMiddleware{}

	return c
}

func (c *Client) R() *Request {
	r := &Request{
		QueryParam: url.Values{},
		Header:     http.Header{},
		PathParams: map[string]string{},
		RawRequest: &protocol.Request{},

		client: c,
	}
	return r
}

func (c *Client) EnableServiceDiscovery() *Client {
	c.enableDiscovery = true

	return c
}

func (c *Client) GetClient() *client.Client {
	return c.client
}

func (c *Client) SetBaseURL(url string) *Client {
	c.baseURL = strings.TrimRight(url, "/")
	return c
}

func (c *Client) NewRequest() *Request {
	return c.R()
}

func (c *Client) execute(req *Request) (*Response, error) {
	// Lock the user-defined pre-request hooks.
	c.udBeforeRequestLock.RLock()
	defer c.udBeforeRequestLock.RUnlock()

	// Lock the post-request hooks.
	c.afterResponseLock.RLock()
	defer c.afterResponseLock.RUnlock()

	// Apply Request middleware
	var err error

	// user defined on before request methods
	// to modify the *resty.Request object
	for _, f := range c.udBeforeRequest {
		if err = f(c, req); err != nil {
			return nil, err
		}
	}

	for _, f := range c.beforeRequest {
		if err = f(c, req); err != nil {
			return nil, err
		}
	}

	if hostHeader := req.Header.Get("Host"); hostHeader != "" {
		req.RawRequest.SetHost(hostHeader)
	}

	resp := &protocol.Response{}
	err = c.client.Do(context.Background(), req.RawRequest, resp)

	response := &Response{
		Request:     req,
		RawResponse: resp,
	}

	if err != nil {
		return response, err
	}

	// Apply Response middleware
	for _, f := range c.afterResponse {
		if err = f(c, response); err != nil {
			break
		}
	}

	return response, err
}
