package main

import (
	"fmt"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/hertz-contrib/easy_http"
)

func main() {
	opts1 := easy_http.NewOption().WithDialTimeout(10).WithWriteTimeout(10)
	hertzClient, _ := client.NewClient(client.WithDialTimeout(10), client.WithWriteTimeout(10))
	opts2 := easy_http.NewOption().WithHertzRawOption(hertzClient)

	c1, _ := easy_http.NewClient(opts1)
	c2 := easy_http.MustNewClient(opts2)
	c3 := easy_http.MustNewClient(&easy_http.Option{})

	res, err := c1.SetHeader("test", "test").SetQueryParam("test1", "test1").R().Get("http://www.baidu.com")
	if err != nil {
		panic(err)
	}
	fmt.Println(res)

	res, err = c2.SetHeader("test", "test").SetQueryParam("test1", "test1").R().Get("http://www.baidu.com")
	if err != nil {
		panic(err)
	}

	res, err = c3.SetHeader("test", "test").SetQueryParam("test1", "test1").R().Get("http://www.baidu.com")
	if err != nil {
		panic(err)
	}

	fmt.Println(res)
}
