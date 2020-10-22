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

func TestSetWebhookEndpointURL(t *testing.T) {
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
				URLPath:     APIEndpointWebhookSettings,
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
				URLPath:     APIEndpointWebhookSettings,
				RequestBody: []byte(`{"endpoint":"https://example.com/abcdefghijklmn"}` + "\n"),
				Error: &APIError{
					Code: 500,
				},
			},
		},
		{
			Label:        "Invalid webhook URL error:not https",
			Endpoint:     "http://example.com/not/https",
			ResponseCode: 400,
			Response:     []byte(`{"message":"Invalid webhook endpoint URL"}`),
			Want: want{
				URLPath:     APIEndpointWebhookSettings,
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
			Label:        "Invalid webhook URL error:more 500 characters",
			Endpoint:     "https://example.com/exceed/500/characters",
			ResponseCode: 500,
			Response:     []byte(`{"message":"The request body has 1 error(s)","details":[{"message":"Size must be between 0 and 500","property":"endpoint"}]}`),
			Want: want{
				URLPath:     APIEndpointWebhookSettings,
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
			res, err := client.SetWebhookEndpointURL(tc.Endpoint).Do()
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

func TestSetWebhookEndpointURLWithContext(t *testing.T) {
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
	_, err = client.SetWebhookEndpointURL("https://example.com/abcdefghijklmn").WithContext(ctx).Do()
	expectCtxDeadlineExceed(ctx, err, t)
}

func BenchmarkSetWebhookEndpointURL(b *testing.B) {
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
		client.SetWebhookEndpointURL("https://example.com/abcdefghijklmn").Do()
	}
}

func TestGetWebhookEndpointInfo(t *testing.T) {
	type want struct {
		URLPath     string
		RequestBody []byte
		Response    *WebhookEndpointInfoResponse
		Error       error
	}
	var testCases = []struct {
		Label        string
		ResponseCode int
		Response     []byte
		Want         want
	}{
		{
			Label:        "Success:endpoint is active",
			ResponseCode: 200,
			Response:     []byte(`{"endpoint":"https://example.com/abcdefghijklmn","active":true}`),
			Want: want{
				URLPath:     APIEndpointWebhookSettings,
				RequestBody: []byte(``),
				Response: &WebhookEndpointInfoResponse{
					Endpoint: "https://example.com/abcdefghijklmn",
					Active:   true,
				},
			},
		},
		{
			Label:        "Success:endpoint is not active",
			ResponseCode: 200,
			Response:     []byte(`{"endpoint":"https://example.com/abcdefghijklmn","active":false}`),
			Want: want{
				URLPath:     APIEndpointWebhookSettings,
				RequestBody: []byte(``),
				Response: &WebhookEndpointInfoResponse{
					Endpoint: "https://example.com/abcdefghijklmn",
					Active:   false,
				},
			},
		},
		{
			Label:        "Internal server error",
			ResponseCode: 500,
			Response:     []byte("500 Internal server error"),
			Want: want{
				URLPath:     APIEndpointWebhookSettings,
				RequestBody: []byte(``),
				Error: &APIError{
					Code: 500,
				},
			},
		},
		{
			Label:        "Not set webhookURL error",
			ResponseCode: 404,
			Response:     []byte(`{"message":"Webhook endpoint not found"}`),
			Want: want{
				URLPath:     APIEndpointWebhookSettings,
				RequestBody: []byte(``),
				Error: &APIError{
					Code: 404,
					Response: &ErrorResponse{
						Message: "Webhook endpoint not found",
					},
				},
			},
		},
	}

	var currentTestIdx int
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		tc := testCases[currentTestIdx]
		if r.Method != http.MethodGet {
			t.Errorf("Method %s; want %s", r.Method, http.MethodGet)
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
			res, err := client.GetWebhookEndpointInfo().Do()
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

func TestGetWebhookEndpointInfoWithContext(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		time.Sleep(10 * time.Millisecond)
		w.Write([]byte(`{"endpoint":"https://example.com/abcdefghijklmn","active":true}`))
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
	_, err = client.GetWebhookEndpointInfo().WithContext(ctx).Do()
	expectCtxDeadlineExceed(ctx, err, t)
}

func BenchmarkGetWebhookEndpointInfo(b *testing.B) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Write([]byte(`{"endpoint":"https://example.com/abcdefghijklmn","active":true}`))
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
		client.GetWebhookEndpointInfo().Do()
	}
}
