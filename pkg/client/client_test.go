/*
 * Copyright 2026 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-selection/v2/pkg/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// recorder answers every request with an empty json list and keeps the last request
type recorder struct {
	mux     sync.Mutex
	request *http.Request
}

func (this *recorder) server(t *testing.T) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		this.mux.Lock()
		this.request = r
		this.mux.Unlock()
		_, _ = w.Write([]byte("[]"))
	}))
	t.Cleanup(server.Close)
	return server
}

func (this *recorder) last() *http.Request {
	this.mux.Lock()
	defer this.mux.Unlock()
	return this.request
}

func TestContextCarriesTheTraceToTheService(t *testing.T) {
	previous := otel.GetTextMapPropagator()
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() { otel.SetTextMapPropagator(previous) })
	traceId, _ := trace.TraceIDFromHex("0af7651916cd43dd8448eb211c80319c")
	spanId, _ := trace.SpanIDFromHex("b7ad6b7169203331")
	ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceId, SpanID: spanId, TraceFlags: trace.FlagsSampled}))
	rec := &recorder{}
	c := NewClient(rec.server(t).URL, Options{})

	calls := map[string]func() error{
		"GetSelectables": func() error {
			_, _, err := c.GetSelectables(ctx, "token", nil, nil)
			return err
		},
		"GetBulkSelectablesV2": func() error {
			_, _, err := c.GetBulkSelectablesV2(ctx, "token", model.BulkRequestV2{}, nil)
			return err
		},
		"DeviceGroupHelper": func() error {
			// the answer is no list, the request was sent before decoding fails
			_, _, _ = c.DeviceGroupHelper(ctx, "token", nil, nil)
			return nil
		},
	}
	for name, call := range calls {
		if err := call(); err != nil {
			t.Fatal(name, err)
		}
		if got := rec.last().Header.Get("traceparent"); got != "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01" {
			t.Errorf("%v sent traceparent %q", name, got)
		}
	}
}

func TestContextCancelsTheRequest(t *testing.T) {
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer server.Close()
	defer close(release)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		_, _, err := NewClient(server.URL, Options{}).GetSelectables(ctx, "token", nil, nil)
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Error("a request past the deadline of its context has to fail")
		}
	case <-time.After(5 * time.Second):
		t.Error("the request was not bounded by its context")
	}
}

func TestOptionsTimeoutBoundsARequest(t *testing.T) {
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer server.Close()
	defer close(release)

	done := make(chan error, 1)
	go func() {
		_, _, err := NewClient(server.URL, Options{Timeout: 50 * time.Millisecond}).GetSelectables(context.Background(), "token", nil, nil)
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Error("a request past the timeout has to fail")
		}
	case <-time.After(5 * time.Second):
		t.Error("the request was not bounded by the timeout")
	}
}

type countingTransport struct {
	mux   sync.Mutex
	count int
}

func (this *countingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	this.mux.Lock()
	this.count++
	this.mux.Unlock()
	return http.DefaultTransport.RoundTrip(req)
}

func TestOptionsHttpClientSendsTheRequestsAndKeepsItsTimeout(t *testing.T) {
	rec := &recorder{}
	transport := &countingTransport{}
	given := &http.Client{Transport: transport, Timeout: time.Hour}

	if _, _, err := NewClient(rec.server(t).URL, Options{HttpClient: given, Timeout: time.Minute}).GetSelectables(context.Background(), "token", nil, nil); err != nil {
		t.Fatal(err)
	}
	if transport.count != 1 {
		t.Errorf("the request has to go through the given client, got %v requests", transport.count)
	}
	if given.Timeout != time.Hour {
		t.Errorf("the timeout is set on a copy, the given client changed to %v", given.Timeout)
	}
}

func TestGetSelectablesOptionsImportPathTrimFirstElement(t *testing.T) {
	rec := &recorder{}
	c := NewClient(rec.server(t).URL, Options{})

	if _, _, err := c.GetSelectables(context.Background(), "token", nil, &GetSelectablesOptions{IncludeImports: true, ImportPathTrimFirstElement: true}); err != nil {
		t.Fatal(err)
	}
	if got := rec.last().URL.Query().Get("import_path_trim_first_element"); got != "true" {
		t.Errorf("import_path_trim_first_element=%q", got)
	}

	// left out when not asked for, as before the option existed
	if _, _, err := c.GetSelectables(context.Background(), "token", nil, &GetSelectablesOptions{IncludeImports: true}); err != nil {
		t.Fatal(err)
	}
	if _, ok := rec.last().URL.Query()["import_path_trim_first_element"]; ok {
		t.Errorf("unexpected parameter in %v", rec.last().URL.RawQuery)
	}
	if got := rec.last().URL.Query(); got.Get("include_imports") != "true" || got.Get("include_devices") != "false" {
		t.Errorf("query %v", got.Encode())
	}
}
