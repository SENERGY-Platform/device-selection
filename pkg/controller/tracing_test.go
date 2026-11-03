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

package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/SENERGY-Platform/device-selection/v2/pkg/configuration"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// traceRecorder is a stand-in for a downstream service that remembers the traceparent of each request
type traceRecorder struct {
	mux          sync.Mutex
	traceparents []string
}

func (this *traceRecorder) server(t *testing.T) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		this.mux.Lock()
		this.traceparents = append(this.traceparents, r.Header.Get("traceparent"))
		this.mux.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if strings.HasPrefix(r.URL.Path, "/import-types/") {
			_, _ = w.Write([]byte("{}"))
			return
		}
		_, _ = w.Write([]byte("[]"))
	}))
	t.Cleanup(server.Close)
	return server
}

func (this *traceRecorder) recorded() []string {
	this.mux.Lock()
	defer this.mux.Unlock()
	return append([]string{}, this.traceparents...)
}

// callerContext returns a context that carries the span context of an incoming request
func callerContext(t *testing.T) (context.Context, trace.TraceID) {
	previous := otel.GetTextMapPropagator()
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() { otel.SetTextMapPropagator(previous) })
	traceId, err := trace.TraceIDFromHex("0af7651916cd43dd8448eb211c80319c")
	if err != nil {
		t.Fatal(err)
	}
	spanId, err := trace.SpanIDFromHex("b7ad6b7169203331")
	if err != nil {
		t.Fatal(err)
	}
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceId, SpanID: spanId, TraceFlags: trace.FlagsSampled})
	return trace.ContextWithSpanContext(context.Background(), spanContext), traceId
}

func checkTraceId(t *testing.T, traceparents []string, expected trace.TraceID) {
	t.Helper()
	if len(traceparents) == 0 {
		t.Fatal("downstream service received no request")
	}
	for _, traceparent := range traceparents {
		carrier := propagation.MapCarrier{"traceparent": traceparent}
		received := trace.SpanContextFromContext(propagation.TraceContext{}.Extract(context.Background(), carrier))
		if !received.IsValid() || received.TraceID() != expected {
			t.Errorf("downstream request carries traceparent %q, expected trace id %v", traceparent, expected)
		}
	}
}

func TestTraceReachesDeviceRepository(t *testing.T) {
	ctx, traceId := callerContext(t)
	downstream := &traceRecorder{}
	config := &configuration.ConfigStruct{LogLevel: "error", DeviceRepoUrl: downstream.server(t).URL}
	ctrl, err := New(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	if _, err, _ = ctrl.GetProtocols(ctx, "Bearer token"); err != nil {
		t.Fatal(err)
	}
	checkTraceId(t, downstream.recorded(), traceId)
}

func TestTraceReachesImportServices(t *testing.T) {
	ctx, traceId := callerContext(t)
	importDeploy := &traceRecorder{}
	importRepo := &traceRecorder{}
	config := &configuration.ConfigStruct{
		LogLevel:        "error",
		ImportDeployUrl: importDeploy.server(t).URL,
		ImportRepoUrl:   importRepo.server(t).URL,
	}
	ctrl, err := New(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	if _, err, _ = ctrl.getImportsByTypes(ctx, "Bearer token", nil); err != nil {
		t.Fatal(err)
	}
	checkTraceId(t, importDeploy.recorded(), traceId)

	if _, err = ctrl.getFullImportType(ctx, "Bearer token", "type-id"); err != nil {
		t.Fatal(err)
	}
	checkTraceId(t, importRepo.recorded(), traceId)
}
