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
	"strings"
	"sync"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type Client struct {
	baseURL string
	header  http.Header

	beforeRequest     []RequestMiddleware
	afterResponse     []ResponseMiddleware
	afterResponseLock *sync.RWMutex

	enableDiscovery bool

	client  *client.Client
	options []config.ClientOption
}

type (
	RequestMiddleware  func(*Client, *Request) error
	ResponseMiddleware func(*Client, *Response) error
)

var (
	hdrContentTypeKey = http.CanonicalHeaderKey(consts.HeaderContentType)

	plainTextType       = consts.MIMETextPlainUTF8
	jsonContentType     = consts.MIMEApplicationJSON
	formContentType     = consts.MIMEApplicationHTMLForm
	formDataContentType = consts.MIMEMultipartPOSTForm
)

func createClient(cc *client.Client, opts ...config.ClientOption) *Client {
	c := &Client{
		afterResponseLock: &sync.RWMutex{},

		client:  cc,
		options: opts,
	}

	c.beforeRequest = []RequestMiddleware{
		parseRequestURL,
		parseRequestHeader,
		createHTTPRequest,
	}

	c.afterResponse = []ResponseMiddleware{
		parseResponseBody,
	}

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

func (c *Client) UseMiddleware(mws ...client.Middleware) *Client {
	c.client.Use(mws...)
	return c
}

func (c *Client) AddHeader(header, value string) *Client {
	c.header.Add(header, value)
	return c
}

func (c *Client) AddHeaders(headers map[string]string) *Client {
	for k, v := range headers {
		c.AddHeader(k, v)
	}
	return c
}

func (c *Client) GetClient() *client.Client {
	return c.client
}

func (c *Client) SetBaseURL(url string) *Client {
	c.baseURL = strings.TrimRight(url, "/")
	return c
}

func (c *Client) SetServiceName(name string) *Client {
	c.SetBaseURL(fmt.Sprintf("http://%s", name))
	return c
}

func (c *Client) NewRequest() *Request {
	return c.R()
}

func (c *Client) execute(req *Request) (*Response, error) {
	// Lock the post-request hooks.
	c.afterResponseLock.RLock()
	defer c.afterResponseLock.RUnlock()
	// Apply Request middleware
	var err error
	for _, f := range c.beforeRequest {
		if err = f(c, req); err != nil {
			return nil, err
		}
	}

	if hostHeader := req.Header.Get("Host"); hostHeader != "" {
		req.RawRequest.SetHost(hostHeader)
	}
	req.hasCreate = true

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
