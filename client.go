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
	hostHeader        = "Host"

	plainTextType       = consts.MIMETextPlainUTF8
	jsonContentType     = consts.MIMEApplicationJSON
	formContentType     = consts.MIMEApplicationHTMLForm
	formDataContentType = consts.MIMEMultipartPOSTForm
)

// createClient creates a new client instance with configured request and response middleware.
// It accepts a client.Client pointer and optional config.ClientOption parameters.
//
// For Example:
//
//	cc := &client.Client{}
//	opts := []config.ClientOption{...}
//	client := createClient(cc, opts...)
//
// Note: This function configures middleware for request processing and response parsing.
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

// R initializes and returns a new Request instance.
// It sets up QueryParam, Header, PathParams, and RawRequest fields.
//
// For Example:
//
//	client := &Client{}
//	req := client.R()
//
// Note: This method does not take any parameters.
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

// EnableServiceDiscovery enables service discovery for the client.
// It sets the enableDiscovery field to true and returns the modified client.
//
// Example:
//
//	client.EnableServiceDiscovery()
func (c *Client) EnableServiceDiscovery() *Client {
	c.enableDiscovery = true
	return c
}

// UseMiddleware adds one or more middleware to the client's request processing chain.
// It returns the client instance for chaining.
//
// For Example:
//
//	client.UseMiddleware(middleware1, middleware2)
func (c *Client) UseMiddleware(mws ...client.Middleware) *Client {
	c.client.Use(mws...)
	return c
}

// AddHeader method adds a custom HTTP header to the Client instance.
// It accepts header name and value as parameters.
//
// For Example:
//
//	client.AddHeader("Authorization", "Bearer token").
//		AddHeader("Content-Type", "application/json")
//
// Returns the updated Client instance for chaining.
func (c *Client) AddHeader(header, value string) *Client {
	c.header.Add(header, value)
	return c
}

// AddHeaders adds multiple HTTP headers to the client instance.
// It iterates over the provided map and calls AddHeader for each key-value pair.
//
// For Example:
//
//	client.AddHeaders(map[string]string{
//		"Authorization": "Bearer token",
//		"Content-Type": "application/json",
//	})
//
// Returns the updated client instance.
func (c *Client) AddHeaders(headers map[string]string) *Client {
	for k, v := range headers {
		c.AddHeader(k, v)
	}
	return c
}

// GetClient retrieves the underlying client.Client instance from the Client.
// It returns a pointer to the client.Client instance.
//
// For Example:
//
//	client := &Client{client: &client.Client{}}
//	underlyingClient := client.GetClient()
func (c *Client) GetClient() *client.Client {
	return c.client
}

// SetBaseURL sets the base URL for the client and trims trailing slashes.
// It returns the updated client instance.
//
// For Example:
//
//	client.SetBaseURL("https://example.com/")
//
// Note: trailing slashes are removed from the URL.
func (c *Client) SetBaseURL(url string) *Client {
	c.baseURL = strings.TrimRight(url, "/")
	return c
}

// SetServiceName sets the service name and updates the base URL.
// It formats the name as the base URL and updates the client.
//
// For Example:
//
//	client.SetServiceName("example.com")
//
// Note: This method does not check the validity of the URL.
func (c *Client) SetServiceName(name string) *Client {
	c.SetBaseURL(fmt.Sprintf("http://%s", name))
	return c
}

// NewRequest creates a new Request instance.
// It calls the R method of the Client.
//
// For Example:
//
//	req := client.NewRequest()
//
// Note: This method does not take any parameters.
func (c *Client) NewRequest() *Request {
	return c.R()
}

// execute method executes an HTTP request and processes the response.
// It locks the post-request hooks to prevent concurrency issues.
//
//	req := &Request{}
//	resp, err := client.execute(req)
//	if err != nil {
//		log.Fatalf("Request failed: %v", err)
//	}
//
// Note: Handles request and response middleware, and sets Host header.
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

	if hostHeader := req.Header.Get(hostHeader); hostHeader != "" {
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
