package easy_http

import (
	"net/http"

	"github.com/cloudwego/hertz/pkg/protocol"
)
type Response struct {
   request     *Request // 上面的 Request 结构体
   rawResponse *protocol.Response

//    bodyByte []byte
   size     int64
}
func (r *Response) Body() []byte {
	if r.rawResponse== nil {
		return []byte{}
	}
	return r.rawResponse.Body()
}
func (r *Response) BodyString() string {
	if r.rawResponse == nil {
		return ""
	}
	return string(r.Body())
}
func (r *Response) Status() string {
	if r.rawResponse == nil {
		return ""
	}
	return ""   // temporary to set ""
}
func (r *Response) Result() interface{} {
	return r.request.Result
}
func (r *Response) Error() interface{} {
	return r.request.Error
}
func (r *Response) Header() http.Header {
	if r.rawResponse== nil {
		return http.Header{}
	}
	header := make(http.Header)
	r.request.RawRequest.Header.VisitAll(func(key, value []byte) {
		keyStr := string(key)
		values := header[keyStr]
		values = append(values, string(value))
		header[keyStr] = values
	})
	return header
}

