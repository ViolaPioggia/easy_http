package easy_http

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"net/url"
	"strings"
	"testing"
	"time"
)

func runHertz(t *testing.T) {
	hertz := server.Default(server.WithHostPorts("127.0.0.1:6666"))
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
		ctx.JSON(200, map[string]string{
			"q1": strings.Join(req.Q1, ","),
			"q2": req.Q2,
			"q3": req.Q3,
		})
	})

	go hertz.Spin()
	time.Sleep(100 * time.Millisecond)
}

func TestSetQueryParam(t *testing.T) {
	runHertz(t)
	//c := MustNewClient().UseMiddleware().SetBaseURL("http://127.0.0.1:6666")
	res := make(map[string]string)
	value := url.Values{}
	value.Set("q3", "q3")
	resp, err := R().
		SetResult(res).
		AddQueryParam("q1", "q1").
		AddQueryParams(map[string]string{"q2": "q2", "q1": "q11"}).
		AddQueryParamsFromValues(value).
		Get(context.Background(), "http://127.0.0.1:6666/test_query")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(resp.BodyString())
	fmt.Println(resp.Result())
	fmt.Println(res)

}
