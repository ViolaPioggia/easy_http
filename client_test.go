package easy_http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func runHertz(t *testing.T, port string) {
	hertz := server.Default(server.WithHostPorts(fmt.Sprintf("127.0.0.1:%s", port)))
	hertz.GET("/test_query", func(c context.Context, ctx *app.RequestContext) {
		type Query struct {
			Q1 []string `query:"q1"`
			Q2 string   `query:"q2"`
			Q3 string   `query:"q3"`
		}
		var req Query
		err := ctx.BindQuery(&req)
		if err != nil {
			t.Fatal(err)
		}
		if len(req.Q1) != 2 {
			t.Errorf("expected q1 has 2 element, but get %d", len(req.Q1))
		}
		if req.Q2 != "q2" {
			t.Errorf("expected q2, but get %s", req.Q2)
		}
		if req.Q3 != "q3" {
			t.Errorf("expected q3, but get %s", req.Q3)
		}
		ctx.JSON(200, req)
	})

	hertz.GET("/test_path/:id/*action", func(c context.Context, ctx *app.RequestContext) {
		type Path struct {
			ID     string `path:"id"`
			Action string `path:"action""`
		}
		var req Path
		err := ctx.Bind(&req)
		if err != nil {
			t.Fatal(err)
		}
		if req.ID != "id" {
			t.Errorf("expected id, but get %s", req.ID)
		}
		if req.Action != "action" {
			t.Errorf("expected action, but get %s", req.Action)
		}
		ctx.JSON(200, req)
	})

	hertz.GET("/test_header", func(c context.Context, ctx *app.RequestContext) {
		type Header struct {
			Header1 []string `header:"Header1"`
			Header2 string   `header:"Header2"`
			Header3 string   `header:"Header3"`
			Header4 string   `header:"Header4"`
		}
		var req Header
		err := ctx.Bind(&req)
		if err != nil {
			t.Fatal(err)
		}
		if len(req.Header1) != 2 {
			t.Errorf("expected Header1 has 2 element, but get %d", len(req.Header1))
		}
		if req.Header2 != "header2" {
			t.Errorf("expected header2, but get %s", req.Header2)
		}
		if req.Header3 != "header3" {
			t.Errorf("expected header3, but get %s", req.Header3)
		}
		if len(req.Header4) != 0 && req.Header4 != "header44" {
			t.Errorf("expected header44, but get %s", req.Header4)
		}

		ctx.JSON(200, req)
	})

	hertz.GET("/test_client_header", func(c context.Context, ctx *app.RequestContext) {
		type Header struct {
			ClientHeader string `header:"Client-Header"`
		}
		var req Header
		err := ctx.Bind(&req)
		if err != nil {
			t.Fatal(err)
		}
		if req.ClientHeader != "client-header" {
			t.Errorf("expected client-header, but get %s", req.ClientHeader)
		}

		ctx.JSON(200, req)
	})

	hertz.POST("/test_content_type", func(c context.Context, ctx *app.RequestContext) {
		type Query struct {
			CT string `query:"ct"`
		}
		var req Query
		err := ctx.Bind(&req)
		if err != nil {
			t.Fatal(err)
		}
		switch req.CT {
		case "json":
			if string(ctx.Request.Header.ContentType()) != consts.MIMEApplicationJSON {
				t.Errorf("expected %s, but get %s", consts.MIMEApplicationJSON, string(ctx.Request.Header.ContentType()))
			}
		case "url":
			if string(ctx.Request.Header.ContentType()) != consts.MIMEApplicationHTMLForm {
				t.Errorf("expected %s, but get %s", consts.MIMEApplicationHTMLForm, string(ctx.Request.Header.ContentType()))
			}
		case "form-data":
			if string(ctx.Request.Header.ContentType()) != consts.MIMEMultipartPOSTForm {
				t.Errorf("expected %s, but get %s", consts.MIMEMultipartPOSTForm, string(ctx.Request.Header.ContentType()))
			}
		}

		ctx.JSON(200, req)
	})

	hertz.POST("/test_body", func(c context.Context, ctx *app.RequestContext) {
		type Query struct {
			CT string `query:"ct"`
		}
		var req Query
		err := ctx.Bind(&req)
		if err != nil {
			t.Fatal(err)
		}
		switch req.CT {
		case "json":
			if string(ctx.Request.Header.ContentType()) != consts.MIMEApplicationJSON {
				t.Errorf("expected %s, but get %s", consts.MIMEApplicationJSON, string(ctx.Request.Header.ContentType()))
			}
			mp := make(map[string]string)
			err := json.Unmarshal(ctx.Request.Body(), &mp)
			if err != nil {
				t.Fatal(err)
			}
			if len(mp) == 0 {
				t.Errorf("expected hertz:hertz, but get nil")
			}
			if value, exist := mp["hertz"]; exist {
				if value != "hertz" {
					t.Errorf("expected hertz, but get %s", value)
				}
			} else {
				t.Errorf("expected hertz exists, but get nil")
			}
		case "url":
			if string(ctx.Request.Header.ContentType()) != consts.MIMEApplicationHTMLForm {
				t.Errorf("expected %s, but get %s", consts.MIMEApplicationHTMLForm, string(ctx.Request.Header.ContentType()))
			}
			if string(ctx.Request.PostArgs().Peek("hertz")) != "hertz" {
				t.Errorf("expected hertz, but get %s", string(ctx.Request.PostArgs().Peek("hertz")))
			}
		case "form-data":
			if !strings.HasPrefix(string(ctx.Request.Header.ContentType()), consts.MIMEMultipartPOSTForm) {
				t.Errorf("expected %s, but get %s", consts.MIMEMultipartPOSTForm, string(ctx.Request.Header.ContentType()))
			}
			if len(ctx.Request.Header.MultipartFormBoundary()) == 0 {
				t.Fatal("expected multipart form boundary, but get nil")
			}
			multipart, err := ctx.Request.MultipartForm()
			if err != nil {
				t.Fatal(err)
			}
			if value, exist := multipart.Value["hertz"]; exist {
				if len(value) != 1 {
					t.Fatalf("expected get multipart form length 1, but get %d", len(value))
				} else {
					if value[0] != "hertz" {
						t.Errorf("expected get multipart form 'hertz', but get '%s'", value[0])
					}
				}
			} else {
				t.Errorf("expected get key hertz, but get nil")
			}
		case "file":
			if !strings.HasPrefix(string(ctx.Request.Header.ContentType()), consts.MIMEMultipartPOSTForm) {
				t.Errorf("expected %s, but get %s", consts.MIMEMultipartPOSTForm, string(ctx.Request.Header.ContentType()))
			}
			if len(ctx.Request.Header.MultipartFormBoundary()) == 0 {
				t.Fatal("expected multipart form boundary, but get nil")
			}
			multipart, err := ctx.Request.MultipartForm()
			if err != nil {
				t.Fatal(err)
			}
			if hertz1, exist := multipart.File["hertz1.txt"]; exist {
				if len(hertz1) != 1 {
					t.Fatalf("expected 1 file, but get %d", len(hertz1))
				} else {
					if hertz1[0].Filename != "client.go" {
						t.Errorf("expected file 'client.go', but get '%s'", hertz1[0].Filename)
					}
				}
			} else {
				t.Fatal("expected hertz1.txt, but get nil")
			}
			if hertz2, exist := multipart.File["hertz2.txt"]; exist {
				if len(hertz2) != 1 {
					t.Fatalf("expected 1 file, but get %d", len(hertz2))
				} else {
					if hertz2[0].Filename != "client.go" {
						t.Errorf("expected file 'client.go', but get '%s'", hertz2[0].Filename)
					}
				}
			} else {
				t.Fatal("expected hertz2.txt, but get nil")
			}
		case "":
			if string(ctx.Request.Body()) != "hertz" {
				t.Errorf("expected hertz, but get %s", string(ctx.Request.Body()))
			}
			if ctx.Request.Header.ContentLength() != 5 {
				t.Errorf("expected content-length 5, but get %d", ctx.Request.Header.ContentLength())
			}

		}

		ctx.JSON(200, req)
	})

	go hertz.Spin()
	time.Sleep(100 * time.Millisecond)
}

func TestQueryParam(t *testing.T) {
	runHertz(t, "6666")
	c := MustNewClient().UseMiddleware().SetBaseURL("http://127.0.0.1:6666")
	res := make(map[string]string)
	value := url.Values{}
	value.Set("q3", "q3")
	// test add query parameters
	_, err := c.R().
		SetResult(res).
		AddQueryParam("q1", "q1").
		AddQueryParams(map[string]string{"q2": "q2", "q1": "q11"}).
		AddQueryParamsFromValues(value).
		Get(context.Background(), "http://127.0.0.1:6666/test_query")
	if err != nil {
		t.Fatal(err)
	}

	value = url.Values{}
	value.Set("q1", "q11")
	res = make(map[string]string)
	// test set query parameters
	_, err = c.R().
		SetResult(res).
		SetQueryString("q1=q1&q2=q2").
		SetQueryParam("q3", "q3").
		SetQueryParamsFromValues(value).
		Get(context.Background(), "http://127.0.0.1:6666/test_query")
	if err != nil {
		t.Fatal(err)
	}

	// test set query string error
	_, err = c.R().SetQueryString("123;").Get(context.Background(), "http://127.0.0.1:6666/test_query")
	if err == nil {
		t.Fatal("expected an error, but get nil")
	}
}

func TestPathParam(t *testing.T) {
	runHertz(t, "7777")
	c := MustNewClient().UseMiddleware().SetBaseURL("http://127.0.0.1:7777")
	// test route path parameter
	_, err := c.R().
		SetPathParam("id", "id").
		SetPathParam("action", "action").
		Get(context.Background(), "http://127.0.0.1:7777/test_path/:id/*action")
	if err != nil {
		t.Fatal(err)
	}

	// test route path parameters
	_, err = c.R().
		SetPathParams(map[string]string{
			"id":     "id",
			"action": "action",
		}).
		Get(context.Background(), "http://127.0.0.1:7777/test_path/:id/*action")
	if err != nil {
		t.Fatal(err)
	}
}

func TestHeaderParam(t *testing.T) {
	runHertz(t, "9999")
	c := MustNewClient().UseMiddleware().SetBaseURL("http://127.0.0.1:9999")
	hc := MustNewClient().UseMiddleware().SetBaseURL("http://127.0.0.1:9999").AddHeader("client-header", "client-header")

	// test client header
	_, err := hc.R().Get(context.Background(), "/test_client_header")
	if err != nil {
		t.Fatal(err)
	}

	h := http.Header{}
	h.Add("header3", "header3")
	// test add header parameter
	_, err = c.R().
		AddHeaders(map[string]string{
			"header1": "header1",
			"header2": "header2",
		}).
		AddHeader("header1", "header11").
		AddHTTPHeader(h).
		Get(context.Background(), "/test_header")
	if err != nil {
		t.Fatal(err)
	}

	h = http.Header{}
	h.Add("header3", "header3")
	// test set header parameter
	_, err = c.R().
		SetHeaders(map[string]string{
			"header1": "header1",
			"header2": "header2",
		}).
		AddHeader("header1", "header11").
		SetHTTPHeader(h).
		SetHeader("header4", "header4").
		SetHeader("header4", "header44").
		Get(context.Background(), "/test_header")
	if err != nil {
		t.Fatal(err)
	}

	// test content-type
	_, err = c.R().
		AddQueryParam("ct", "form-data").
		SetContentTypeFormData().
		Post(context.Background(), "/test_content_type")

	_, err = c.R().
		AddQueryParam("ct", "json").
		SetContentTypeJSON().
		Post(context.Background(), "/test_content_type")

	_, err = c.R().
		AddQueryParam("ct", "url").
		SetContentTypeUrlEncode().
		Post(context.Background(), "/test_content_type")
}

func TestBodyParam(t *testing.T) {
	runHertz(t, "11111")
	c := MustNewClient().UseMiddleware().SetBaseURL("http://127.0.0.1:11111")

	// test text body
	_, err := c.R().SetBody("hertz").Post(context.Background(), "/test_body")
	if err != nil {
		t.Fatal(err)
	}

	// test text body
	_, err = c.R().SetBody([]byte("hertz")).Post(context.Background(), "/test_body")
	if err != nil {
		t.Fatal(err)
	}

	// test json
	_, err = c.R().
		SetContentTypeJSON().
		SetQueryParam("ct", "json").
		SetBody(map[string]string{"hertz": "hertz"}).
		Post(context.Background(), "/test_body")
	if err != nil {
		t.Fatal(err)
	}

	// test form
	u := url.Values{}
	u.Add("hertz", "hertz")
	_, err = c.R().
		SetQueryParam("ct", "url").
		SetFormDataFromValues(u).
		SetContentTypeUrlEncode().
		Post(context.Background(), "/test_body")

	// test form-data
	_, err = c.R().
		SetQueryParam("ct", "form-data").
		SetContentTypeFormData().
		SetMultipartFormData(map[string]string{
			"hertz": "hertz",
		}).
		Post(context.Background(), "/test_body")

	// test file upload
	_, err = c.R().
		SetQueryParam("ct", "file").
		SetContentTypeFormData().
		SetFile("hertz1.txt", "./client.go").
		SetFiles(map[string]string{
			"hertz2.txt": "./client.go",
		}).
		Post(context.Background(), "/test_body")
}
