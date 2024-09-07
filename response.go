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
	"github.com/cloudwego/hertz/pkg/protocol"
	"io"
	"net/http"
)

// done

type Response struct {
	Request     *Request
	RawResponse *protocol.Response
}

func (r *Response) Body() []byte {
	if r.RawResponse == nil {
		return []byte{}
	}
	return r.RawResponse.Body()
}

func (r *Response) BodyString() string {
	if r.RawResponse == nil {
		return ""
	}
	return string(r.RawResponse.Body())
}

func (r *Response) StatusCode() int {
	if r.RawResponse == nil {
		return 0
	}
	return r.RawResponse.StatusCode()
}

func (r *Response) Result() interface{} {
	return r.Request.Result
}

func (r *Response) Error() interface{} {
	return r.Request.Error
}

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

func (r *Response) IsSuccess() bool {
	return r.StatusCode() > 199 && r.StatusCode() < 300
}

func (r *Response) IsError() bool {
	return r.StatusCode() > 399
}
