// Copyright 2020 LINE Corporation
//
// LINE Corporation licenses this file to you under the Apache License,
// version 2.0 (the "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at:
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package linebot

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
)

// SetWebhookEndpointURLCall type
type SetWebhookEndpointURLCall struct {
	c   *Client
	ctx context.Context

	endpoint string
}

// SetWebhookEndpointURL method
func (client *Client) SetWebhookEndpointURL(webhookEndpoint string) *SetWebhookEndpointURLCall {
	return &SetWebhookEndpointURLCall{
		c:        client,
		endpoint: webhookEndpoint,
	}
}

// WithContext method
func (call *SetWebhookEndpointURLCall) WithContext(ctx context.Context) *SetWebhookEndpointURLCall {
	call.ctx = ctx
	return call
}

func (call *SetWebhookEndpointURLCall) encodeJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(&struct {
		Endpoint string `json:"endpoint"`
	}{
		Endpoint: call.endpoint,
	})
}

// Do method
func (call *SetWebhookEndpointURLCall) Do() (*BasicResponse, error) {
	var buf bytes.Buffer
	if err := call.encodeJSON(&buf); err != nil {
		return nil, err
	}
	res, err := call.c.put(call.ctx, APIEndpointWebhookSettings, &buf)
	if err != nil {
		return nil, err
	}
	defer closeResponse(res)
	return decodeToBasicResponse(res)
}

// GetWebhookEndpointInfoCall type
type GetWebhookEndpointInfoCall struct {
	c   *Client
	ctx context.Context
}

// GetWebhookEndpointInfo method
func (client *Client) GetWebhookEndpointInfo() *GetWebhookEndpointInfoCall {
	return &GetWebhookEndpointInfoCall{
		c: client,
	}
}

// WithContext method
func (call *GetWebhookEndpointInfoCall) WithContext(ctx context.Context) *GetWebhookEndpointInfoCall {
	call.ctx = ctx
	return call
}

// Do method
func (call *GetWebhookEndpointInfoCall) Do() (*WebhookEndpointInfoResponse, error) {
	res, err := call.c.get(call.ctx, call.c.endpointBase, APIEndpointWebhookSettings, nil)
	if err != nil {
		return nil, err
	}
	defer closeResponse(res)
	return decodeToWebhookEndpointInfoResponse(res)
}
