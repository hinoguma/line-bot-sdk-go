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

// SetWebhookEndpointCall type
type SetWebhookEndpointCall struct {
	c   *Client
	ctx context.Context

	endpoint string
}

// SetWebhookEndpoint method
func (client *Client) SetWebhookEndpoint(webhookEndpoint string) *SetWebhookEndpointCall {
	return &SetWebhookEndpointCall{
		c:        client,
		endpoint: webhookEndpoint,
	}
}

// WithContext method
func (call *SetWebhookEndpointCall) WithContext(ctx context.Context) *SetWebhookEndpointCall {
	call.ctx = ctx
	return call
}

func (call *SetWebhookEndpointCall) encodeJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(&struct {
		Endpoint string `json:"endpoint"`
	}{
		Endpoint: call.endpoint,
	})
}

// Do method
func (call *SetWebhookEndpointCall) Do() (*BasicResponse, error) {
	var buf bytes.Buffer
	if err := call.encodeJSON(&buf); err != nil {
		return nil, err
	}
	res, err := call.c.put(call.ctx, APIEndpointSetWebhookEndpoint, &buf)
	if err != nil {
		return nil, err
	}
	defer closeResponse(res)
	return decodeToBasicResponse(res)
}
