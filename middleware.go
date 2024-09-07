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
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

func parseRequestURL(c *Client, r *Request) error {
	if len(r.PathParams) > 0 {
		for p, v := range r.PathParams {
			if strings.HasSuffix(r.URL, "*"+p) { // "*" must be at end of route
				r.URL = strings.Replace(r.URL, "*"+p, url.PathEscape(v), 1)
				continue
			}
			r.URL = strings.Replace(r.URL, ":"+p, url.PathEscape(v), -1)
		}
	}

	// Parsing request URL
	reqURL, err := url.Parse(r.URL)
	if err != nil {
		return err
	}

	// If Request.URL is relative path then added c.HostURL into
	// the request URL otherwise Request.URL will be used as-is
	if !reqURL.IsAbs() {
		r.URL = reqURL.String()
		if len(r.URL) > 0 && r.URL[0] != '/' {
			r.URL = "/" + r.URL
		}
		reqURL, err = url.Parse(c.baseURL + r.URL)
		if err != nil {
			return err
		}
	}

	// Adding Query Param
	if len(r.QueryParam) > 0 {
		if len(r.QueryParam) > 0 {
			if len(strings.TrimSpace(reqURL.RawQuery)) == 0 {
				reqURL.RawQuery = r.QueryParam.Encode()
			} else {
				reqURL.RawQuery = reqURL.RawQuery + "&" + r.QueryParam.Encode()
			}
		}
	}

	r.URL = reqURL.String()

	return nil
}

func parseRequestHeader(c *Client, r *Request) error {
	hdr := make(http.Header)
	if c.header != nil {
		for k := range c.header {
			hdr[k] = append(hdr[k], c.header[k]...)
		}
	}

	for k := range r.Header {
		hdr.Del(k)
		hdr[k] = append(hdr[k], r.Header[k]...)
	}

	if len(r.FormData) != 0 {
		hdr.Add(hdrContentTypeKey, formContentType)
	}

	r.Header = hdr

	return nil
}

func parseRequestBody(c *Client, r *Request) error {
	switch {
	case r.RawRequest.HasMultipartForm(): // Handling Multipart
		handleMultipart(c, r)
	case len(c.FormData) > 0 || len(r.FormData) > 0: // Handling Form Data
		handleFormData(c, r)
		//case r.RawRequest.Body() != nil: // Handling Request body
		//	handleContentType(c, r)
	}

	return nil
}

func handleMultipart(c *Client, r *Request) {
	r.RawRequest.SetMultipartFormData(c.FormData)

	r.Header.Set(hdrContentTypeKey, formDataContentType)
}

func handleFormData(c *Client, r *Request) {
	r.RawRequest.SetFormData(c.FormData)

	r.Header.Set(hdrContentTypeKey, formContentType)
}

//func handleContentType(c *Client, r *Request) {
//	contentType := r.Header.Get(hdrContentTypeKey)
//	if len(strings.TrimSpace(contentType)) == 0 {
//		contentType = DetectContentType(r.RawRequest.Body())
//		r.Header.Set(hdrContentTypeKey, contentType)
//	}
//}

func DetectContentType(body interface{}) string {
	contentType := plainTextType
	kind := reflect.Indirect(reflect.ValueOf(body)).Kind()
	switch kind {
	case reflect.Struct, reflect.Map:
		contentType = jsonContentType
	case reflect.String:
		contentType = plainTextType
	default:
		if b, ok := body.([]byte); ok {
			contentType = http.DetectContentType(b)
		} else if kind == reflect.Slice {
			contentType = jsonContentType
		}
	}

	return contentType
}
