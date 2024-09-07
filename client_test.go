package easy_http

import (
	"context"
	"fmt"
	"testing"
)

func TestSetQueryParam(t *testing.T) {
	c := MustNewClient().EnableServiceDiscovery().UseMiddleware().SetBaseURL("https://example.com")

	resp, err := c.R().AddQueryParam("", "").AddHeader("", "").SetBody("").Get(context.Background(), "/a")
	fmt.Println(err)
	fmt.Println(string(resp.Body()))
	fmt.Println(resp.StatusCode())
	fmt.Println(resp.ToRawHTTPResponse())
	fmt.Println(resp.Request.ToCurl())
}
