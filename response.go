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
	"bytes"
	"io"
	"net/http"

	"github.com/cloudwego/hertz/pkg/protocol"
)

type Response struct {
	Request     *Request
	RawResponse *protocol.Response

	bodyByte []byte
	size     int
}

// Body method returns the response body content from the Response object.
// If the RawResponse is nil, it returns an empty byte slice.
//
// For Example:
//
//	resp := client.Get("/endpoint")
//	body := resp.Body()
//
// Note: Ensure RawResponse is not nil before calling this method.
func (r *Response) Body() []byte {
	if r.RawResponse == nil {
		return []byte{}
	}
	return r.RawResponse.Body()
}

// BodyString returns the response body as a string.
// If RawResponse is nil, it returns an empty string.
//
// For Example:
//
//	resp := &Response{RawResponse: someHTTPResponse}
//	fmt.Println(resp.BodyString())
func (r *Response) BodyString() string {
	if r.RawResponse == nil {
		return ""
	}
	return string(r.RawResponse.Body())
}

// StatusCode returns the HTTP response status code.
// If RawResponse is nil, it returns 0.
//
// For Example:
//
//	resp := client.R().Get("/endpoint")
//	code := resp.StatusCode()
//
// Note: Ensure RawResponse is not nil before calling.
func (r *Response) StatusCode() int {
	if r.RawResponse == nil {
		return 0
	}
	return r.RawResponse.StatusCode()
}

// Result method returns the result of the request from the Response object.
// The result type is an interface{}, which can be any type depending on the request.
//
// For Example:
//
//	resp := client.R().Get("/endpoint")
//	result := resp.Result()
//
// Note: Ensure to type assert the result to the expected type.
func (r *Response) Result() interface{} {
	return r.Request.Result
}

// GetRequest returns the Request instance associated with the Response.
//
// For Example:
//
//	resp := &Response{}
//	req := resp.GetRequest()
func (r *Response) GetRequest() *Request {
	return r.Request
}

// GetRawResponse returns the raw protocol response from the Response struct.
//
// For Example:
//
//	resp := &Response{RawResponse: &protocol.Response{...}}
//	rawResp := resp.GetRawResponse()
func (r *Response) GetRawResponse() *protocol.Response {
	return r.RawResponse
}

// Error method retrieves the error from the associated Request object.
// It returns an error indicating any issues encountered during the Request.
//
// For Example:
//
//	err := response.Error()
//	if err != nil {
//		fmt.Println("Error:", err)
//	}
func (r *Response) Error() error {
	return r.Request.Error
}

// Header method extracts HTTP headers from the Response object.
// If RawResponse is nil, it returns an empty http.Header.
//
// For Example:
//
//	resp := client.Get("/example")
//	headers := resp.Header()
//
// Note: This method does not modify the Response object.
func (r *Response) Header() http.Header {
	if r.RawResponse == nil {
		return http.Header{}
	}
	header := make(http.Header)
	r.RawResponse.Header.VisitAll(func(key, value []byte) {
		header.Add(string(key), string(value))
	})
	return header
}

// Cookies method extracts all cookies from the HTTP response.
// It returns a slice of http.Cookie.
//
// For Example:
//
//	cookies := response.Cookies()
//	for _, cookie := range cookies {
//		fmt.Println(cookie.Name, cookie.Value)
//	}
//
// Note: Returns an empty slice if RawResponse is nil.
func (r *Response) Cookies() []*http.Cookie {
	if r.RawResponse == nil {
		return make([]*http.Cookie, 0)
	}
	var cookies []*http.Cookie
	r.RawResponse.Header.VisitAllCookie(func(key, value []byte) {
		cookies = append(cookies, &http.Cookie{
			Name:  string(key),
			Value: string(value),
		})
	})

	return cookies
}

// ToRawHTTPResponse converts Response object to raw HTTP response string.
// It sets StatusCode, Header, and Body from Response object.
//
// For Example:
//
//	resp := &Response{}
//	rawHTTP := resp.ToRawHTTPResponse()
//
// Note: Ensure Response object is properly initialized.
func (r *Response) ToRawHTTPResponse() string {
	resp := &http.Response{
		StatusCode: r.StatusCode(),
		Header:     r.Header(),
		Body:       io.NopCloser(bytes.NewReader(r.Body())),
	}
	for _, cookie := range r.Cookies() {
		resp.Header.Add("Set-Cookie", cookie.String())
	}
	var buffer bytes.Buffer
	resp.Write(&buffer)

	return buffer.String()
}

// IsSuccess checks if the HTTP response is successful.
// It returns true if the status code is between 200 and 299.
//
// For Example:
//
//	resp := client.R().Get("/endpoint")
//	if resp.IsSuccess() {
//		fmt.Println("Request was successful")
//	}
func (r *Response) IsSuccess() bool {
	return r.StatusCode() > 199 && r.StatusCode() < 300
}

// IsError checks if the HTTP response indicates an error.
// It returns true if the status code is greater than 399.
//
// Example:
//
//	resp := client.Get("/endpoint")
//	if resp.IsError() {
//		fmt.Println("Error in response")
//	}
func (r *Response) IsError() bool {
	return r.StatusCode() > 399
}
