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
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestSetWebhookEndpoint(t *testing.T) {
	type want struct {
		URLPath     string
		RequestBody []byte
		Response    *BasicResponse
		Error       error
	}
	var testCases = []struct {
		Label        string
		Endpoint     string
		ResponseCode int
		Response     []byte
		Want         want
	}{
		{
			Label:        "Success",
			Endpoint:     "https://example.com/abcdefghijklmn",
			ResponseCode: 200,
			Response:     []byte(`{}`),
			Want: want{
				URLPath:     APIEndpointSetWebhookEndpoint,
				RequestBody: []byte(`{"endpoint":"https://example.com/abcdefghijklmn"}` + "\n"),
				Response:    &BasicResponse{},
			},
		},
		{
			Label:        "Internal server error",
			Endpoint:     "https://example.com/abcdefghijklmn",
			ResponseCode: 500,
			Response:     []byte("500 Internal server error"),
			Want: want{
				URLPath:     APIEndpointSetWebhookEndpoint,
				RequestBody: []byte(`{"endpoint":"https://example.com/abcdefghijklmn"}` + "\n"),
				Error: &APIError{
					Code: 500,
				},
			},
		},
		{
			Label:        "Invalid webhook URL error",
			Endpoint:     "http://example.com/not/https",
			ResponseCode: 400,
			Response:     []byte(`{"message":"Invalid webhook endpoint URL"}`),
			Want: want{
				URLPath:     APIEndpointSetWebhookEndpoint,
				RequestBody: []byte(`{"endpoint":"http://example.com/not/https"}` + "\n"),
				Error: &APIError{
					Code: 400,
					Response: &ErrorResponse{
						Message: "Invalid webhook endpoint URL",
					},
				},
			},
		},
		{
			Label:        "Invalid webhook URL error",
			Endpoint:     "https://example.com/exceed/500/characters",
			ResponseCode: 500,
			Response:     []byte(`{"message":"The request body has 1 error(s)","details":[{"message":"Size must be between 0 and 500","property":"endpoint"}]}`),
			Want: want{
				URLPath:     APIEndpointSetWebhookEndpoint,
				RequestBody: []byte(`{"endpoint":"https://example.com/exceed/500/characters"}` + "\n"),
				Error: &APIError{
					Code: 500,
					Response: &ErrorResponse{
						Message: "The request body has 1 error(s)",
						Details: []errorResponseDetail{
							{
								Message:  "Size must be between 0 and 500",
								Property: "endpoint",
							},
						},
					},
				},
			},
		},
	}

	var currentTestIdx int
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		tc := testCases[currentTestIdx]
		if r.Method != http.MethodPut {
			t.Errorf("Method %s; want %s", r.Method, http.MethodPut)
		}
		if r.URL.Path != tc.Want.URLPath {
			t.Errorf("URLPath %s; want %s", r.URL.Path, tc.Want.URLPath)
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(body, tc.Want.RequestBody) {
			t.Errorf("RequestBody %s; want %s", body, tc.Want.RequestBody)
		}
		w.WriteHeader(tc.ResponseCode)
		w.Write(tc.Response)
	}))
	defer server.Close()

	dataServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		t.Error("Unexpected Data API call")
		w.WriteHeader(404)
		w.Write([]byte(`{"message":"Not found"}`))
	}))
	defer dataServer.Close()

	client, err := mockClient(server, dataServer)
	if err != nil {
		t.Fatal(err)
	}
	for i, tc := range testCases {
		currentTestIdx = i
		t.Run(strconv.Itoa(i)+"/"+tc.Label, func(t *testing.T) {
			res, err := client.SetWebhookEndpoint(tc.Endpoint).Do()
			if tc.Want.Error != nil {
				if !reflect.DeepEqual(err, tc.Want.Error) {
					t.Errorf("Error %v; want %v", err, tc.Want.Error)
				}
			} else {
				if err != nil {
					t.Error(err)
				}
			}
			if !reflect.DeepEqual(res, tc.Want.Response) {
				t.Errorf("Response %v; want %v", res, tc.Want.Response)
			}
		})
	}
}

func TestSetWebhookEndpointWithContext(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		time.Sleep(10 * time.Millisecond)
		w.Write([]byte("{}"))
	}))
	defer server.Close()

	dataServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		t.Error("Unexpected Data API call")
		w.WriteHeader(404)
		w.Write([]byte(`{"message":"Not found"}`))
	}))
	defer dataServer.Close()

	client, err := mockClient(server, dataServer)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	_, err = client.SetWebhookEndpoint("https://example.com/abcdefghijklmn").WithContext(ctx).Do()
	expectCtxDeadlineExceed(ctx, err, t)
}

func BenchmarkSetWebhookEndpoint(b *testing.B) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	dataServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		b.Error("Unexpected Data API call")
		w.WriteHeader(404)
		w.Write([]byte(`{"message":"Not found"}`))
	}))
	defer dataServer.Close()

	client, err := mockClient(server, dataServer)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.SetWebhookEndpoint("https://example.com/abcdefghijklmn").Do()
	}
}
