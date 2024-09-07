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
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/network/standard"
)

func NewClient(opts ...config.ClientOption) (*Client, error) {
	// 默认使用标准库以支持 https
	opts = append(opts, client.WithDialer(standard.NewDialer()))
	c, err := client.NewClient(opts...)
	return createClient(c, opts...), err
}

func MustNewClient(opts ...config.ClientOption) *Client {
	// 默认使用标准库以支持 https
	opts = append(opts, client.WithDialer(standard.NewDialer()))
	c, err := client.NewClient(opts...)
	if err != nil {
		panic(err)
	}
	return createClient(c, opts...)
}
