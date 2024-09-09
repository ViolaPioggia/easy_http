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
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

var (
	jsonCheck = regexp.MustCompile(`(?i:(application|text)/(json|.*\+json|json\-.*)(; |$))`)
	xmlCheck  = regexp.MustCompile(`(?i:(application|text)/(xml|.*\+xml)(; |$))`)
)

// parseRequestURL parses and processes the request URL to ensure it is correctly formatted.
// It includes path and query parameters.
//
// For Example:
//
//	client := &Client{baseURL: "http://example.com"}
//	req := &Request{URL: "/api/:id", PathParams: map[string]string{"id": "123"}}
//	err := parseRequestURL(client, req)
//
// Note: This function handles both path and query parameters.
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

// parseRequestHeader merges client and request headers.
// It handles form data by adding content type header.
//
// For Example:
//
//	client := &Client{header: http.Header{"Key": {"Value"}}}
//	req := &Request{Header: http.Header{"AnotherKey": {"AnotherValue"}}, FormData: map[string][]string{}}
//	err := parseRequestHeader(client, req)
//
// Note: Always returns nil error.
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

// isPayloadSupported checks if the given HTTP method supports a request body.
// It returns true if the method is not HEAD, OPTIONS, GET, or DELETE.
//
// For Example:
//
//	isPayloadSupported("POST") // returns true
//	isPayloadSupported("GET")  // returns false
func isPayloadSupported(m string) bool {
	return !(m == consts.MethodHead || m == consts.MethodOptions || m == consts.MethodGet || m == consts.MethodDelete)
}

// isStringEmpty checks if a given string is empty.
// It trims spaces from both ends and checks if the length is zero.
//
// For Example:
//
//	isStringEmpty("   ") // returns true
//	isStringEmpty("hello") // returns false
func isStringEmpty(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}

// IsJSONType method is to check JSON content type or not
func isJSONType(ct string) bool {
	return jsonCheck.MatchString(ct)
}

// IsXMLType method is to check XML content type or not
func isXMLType(ct string) bool {
	return xmlCheck.MatchString(ct)
}

// detectContentType method is used to figure out "request.Body" content type for request header
func detectContentType(body interface{}) string {
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

// parseRequestBody parses HTTP request body and returns content type, body reader, and error.
// It checks if the request method supports payload, determines content type, and serializes body.
//
// For Example:
//
//	req := &Request{Method: "POST", Body: []byte("example")}
//	contentType, body, err := parseRequestBody(req)
//
// Note: Handles multipart, form data, and detects content type if not specified.
func parseRequestBody(r *Request) (contentType string, body io.Reader, err error) {
	if !isPayloadSupported(r.Method) {
		return
	}
	if r.isMultiPart {
		return formDataContentType, nil, nil
	} else if len(r.FormData) > 0 {
		return formContentType, nil, nil
	}
	var bodyBytes []byte
	contentType = r.Header.Get(hdrContentTypeKey)
	if isStringEmpty(contentType) {
		contentType = detectContentType(r.Body)
		r.AddHeader(hdrContentTypeKey, contentType)
	}

	switch bodyValue := r.Body.(type) {
	case []byte:
		bodyBytes = bodyValue
	case string:
		bodyBytes = []byte(bodyValue)
	default:
		contentType = r.Header.Get(hdrContentTypeKey)
		kind := reflect.Indirect(reflect.ValueOf(r.Body)).Kind()
		var err error
		if isJSONType(contentType) && (kind == reflect.Struct || kind == reflect.Map || kind == reflect.Slice) {
			bodyBytes, err = json.Marshal(r.Body)
		} else if isXMLType(contentType) && (kind == reflect.Struct) {
			bodyBytes, err = xml.Marshal(r.Body)
		}
		if err != nil {
			return "", nil, err
		}
	}

	return contentType, strings.NewReader(string(bodyBytes)), nil
}

// createHTTPRequest creates an HTTP request based on the given client and request.
// It sets the content type, body, and headers accordingly.
//
// For Example:
//
//	client := &Client{}
//	req := &Request{Method: "POST", URL: "https://example.com"}
//	err := createHTTPRequest(client, req)
//
// Note: Handles multipart and form data, sets cookies and options.
func createHTTPRequest(c *Client, r *Request) (err error) {
	contentType, body, err := parseRequestBody(r)
	if err != nil {
		return err
	}
	if !isStringEmpty(contentType) {
		r.Header.Set(hdrContentTypeKey, contentType)
	}

	r.RawRequest = protocol.NewRequest(r.Method, r.URL, body)
	if contentType == formDataContentType && isPayloadSupported(r.Method) {
		if r.RawRequest.IsBodyStream() {
			r.RawRequest.ResetBody()
		}
		r.RawRequest.SetMultipartFormData(r.MultipartFormParams)
		r.RawRequest.SetFiles(r.File)
	} else if contentType == formContentType && isPayloadSupported(r.Method) {
		r.RawRequest.SetFormDataFromValues(r.FormData)
	}

	for key, values := range r.Header {
		for _, val := range values {
			r.RawRequest.Header.Add(key, val)
		}
	}
	for _, cookie := range r.Cookie {
		r.RawRequest.SetCookie(cookie.Name, cookie.Value)
	}

	r.RawRequest.SetOptions(r.RequestOptions...)

	return nil
}

// parseResponseBody parses HTTP response body based on content type (JSON/XML).
// It handles error responses by encoding error info into JSON.
//
// For Example:
//
//	client := &Client{}
//	resp := &Response{}
//	err := parseResponseBody(client, resp)
//
// Note: Handles only JSON or XML content types.
func parseResponseBody(c *Client, resp *Response) (err error) {
	if resp.StatusCode() == http.StatusNoContent {
		return
	}
	resp.bodyByte = resp.Body()
	resp.size = resp.RawResponse.Header.ContentLength()
	// Handles only JSON or XML content type
	ct := resp.Header().Get(hdrContentTypeKey)
	isError := resp.IsError()
	if isError {
		jsonByte, jsonErr := json.Marshal(map[string]interface{}{
			"status_code": resp.RawResponse.StatusCode(),
			"body":        resp.BodyString(),
		})
		if jsonErr != nil {
			return jsonErr
		}
		err = fmt.Errorf(string(jsonByte))
	} else if resp.Request.Result != nil {
		if isJSONType(ct) || isXMLType(ct) {
			err = unmarshalContent(ct, resp.Body(), resp.Request.Result)
			return
		}
	}
	return
}

// unmarshalContent content into object from JSON or XML
func unmarshalContent(ct string, b []byte, d interface{}) (err error) {
	if isJSONType(ct) {
		err = json.Unmarshal(b, d)
	} else if isXMLType(ct) {
		err = xml.Unmarshal(b, d)
	}

	return
}
