package easy_http

import (
	"github.com/cloudwego/hertz/pkg/app/client"
	"net/http"
	"net/url"
)

type Client struct {
	QueryParam url.Values
	PathParams map[string]string
	Header     http.Header
	Cookies    []*http.Cookie

	client *client.Client
}

func New() *Client {
	c, _ := client.NewClient()
	return &Client{client: c}
}

func NewWithHertzClient(c *client.Client) *Client {
	return createClient(c)
}

func createClient(c *client.Client) *Client {
	return &Client{client: c}
}

func (c *Client) SetQueryParam(param, value string) *Client {
	c.QueryParam.Set(param, value)
	return c
}

func (c *Client) SetQueryParams(params map[string]string) *Client {
	for k, v := range params {
		c.QueryParam.Set(k, v)
	}
	return c
}

func (c *Client) SetQueryParamsFromValues(params url.Values) *Client {
	for k, v := range params {
		for _, v1 := range v {
			c.QueryParam.Add(k, v1)
		}
	}
	return c
}

func (c *Client) SetQueryString(query string) *Client {

	return c
}

func (c *Client) AddQueryParam(param, value string) *Client {
	c.QueryParam.Add(param, value)
	return c
}

func (c *Client) AddQueryParams(params map[string]string) *Client {
	for k, v := range params {
		c.QueryParam.Add(k, v)
	}
	return c
}

func (c *Client) SetPathParam(param, value string) *Client {
	c.PathParams[param] = value
	return c
}

func (c *Client) SetPathParams(params map[string]string) *Client {
	for k, v := range params {
		c.PathParams[k] = v
	}
	return c
}

func (c *Client) SetHeader(header, value string) *Client {
	c.Header.Set(header, value)
	return c
}

func (c *Client) SetHeaders(headers map[string]string) *Client {
	for k, v := range headers {
		c.Header.Set(k, v)
	}
	return c
}

func (c *Client) AddHeader(header, value string) *Client {
	c.Header.Add(header, value)
	return c
}

func (c *Client) AddHeaders(headers map[string]string) *Client {
	for k, v := range headers {
		c.Header.Add(k, v)
	}
	return c
}

func (c *Client) SetCookie(hc *http.Cookie) *Client {
	c.Cookies = append(c.Cookies, hc)
	return c
}

func (c *Client) SetCookies(hcs []*http.Cookie) *Client {
	c.Cookies = append(c.Cookies, hcs...)
	return c
}

func (c *Client) R() *Request {
	r := &Request{
		QueryParam: url.Values{},
		Header:     http.Header{},
		Cookies:    make([]*http.Cookie, 0),
		PathParams: map[string]string{},

		client: c,
	}
	return r
}

func (c *Client) NewRequest() *Request {
	return c.R()
}
