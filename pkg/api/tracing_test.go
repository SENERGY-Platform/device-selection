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

package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SENERGY-Platform/device-selection/v2/pkg/configuration"
	"github.com/SENERGY-Platform/gin-middleware/otelx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// recordSpans makes the global provider record into the returned recorder for the test
func recordSpans(t *testing.T) *tracetest.SpanRecorder {
	// otelx initializes the global provider only once per process; doing it here first keeps it
	// from replacing the recording provider below when withTracing calls it
	_, _ = otelx.HTTPOpenTelemetry(context.Background(), "localhost:0", "init", http.NotFoundHandler())
	previousProvider, previousPropagator := otel.GetTracerProvider(), otel.GetTextMapPropagator()
	t.Cleanup(func() {
		otel.SetTracerProvider(previousProvider)
		otel.SetTextMapPropagator(previousPropagator)
	})
	recorder := tracetest.NewSpanRecorder()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)))
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return recorder
}

func TestTracingContinuesTheTraceOfTheCaller(t *testing.T) {
	recorder := recordSpans(t)
	config := &configuration.ConfigStruct{LogLevel: "error"}
	handler := withTracing(config, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	request := httptest.NewRequest(http.MethodPost, "/v2/query/selectables", nil)
	request.Header.Set("traceparent", "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")
	handler.ServeHTTP(httptest.NewRecorder(), request)

	spans := recorder.Ended()
	if len(spans) == 0 {
		t.Fatal("no span recorded")
	}
	if got := spans[0].SpanContext().TraceID().String(); got != "0af7651916cd43dd8448eb211c80319c" {
		t.Errorf("trace id %v, expected the one of the caller", got)
	}
}

func TestTracingOffRecordsNoSpans(t *testing.T) {
	recorder := recordSpans(t)
	config := &configuration.ConfigStruct{OtelEndpoint: configuration.OtelEndpointOff, LogLevel: "error"}
	handler := withTracing(config, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v2/query/selectables", nil))

	if response.Code != http.StatusNoContent {
		t.Errorf("the request reaches the handler: %v", response.Code)
	}
	if spans := recorder.Ended(); len(spans) != 0 {
		t.Errorf("no spans with tracing off, got %v", len(spans))
	}
}
