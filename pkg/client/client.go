/*
 * Copyright 2024 InfAI (CC SES)
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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/SENERGY-Platform/device-selection/v2/pkg/model"
	"github.com/SENERGY-Platform/models/go/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// InternalAdminToken is expired and invalid. but because this service does not validate the received tokens,
// it may be used by trusted internal services which are within the same network (kubernetes cluster).
// requests with this token may not be routed over an ingres with token validation
const InternalAdminToken = `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjEwMDAwMDAwMDAsImlhdCI6MTAwMDAwMDAwMCwiYXV0aF90aW1lIjoxMDAwMDAwMDAwLCJpc3MiOiJpbnRlcm5hbCIsImF1ZCI6W10sInN1YiI6ImRkNjllYTBkLWY1NTMtNDMzNi04MGYzLTdmNDU2N2Y4NWM3YiIsInR5cCI6IkJlYXJlciIsImF6cCI6ImZyb250ZW5kIiwicmVhbG1fYWNjZXNzIjp7InJvbGVzIjpbImFkbWluIiwiZGV2ZWxvcGVyIiwidXNlciJdfSwicmVzb3VyY2VfYWNjZXNzIjp7Im1hc3Rlci1yZWFsbSI6eyJyb2xlcyI6W119LCJCYWNrZW5kLXJlYWxtIjp7InJvbGVzIjpbXX0sImFjY291bnQiOnsicm9sZXMiOltdfX0sInJvbGVzIjpbImFkbWluIiwiZGV2ZWxvcGVyIiwidXNlciJdLCJuYW1lIjoiU2VwbCBBZG1pbiIsInByZWZlcnJlZF91c2VybmFtZSI6InNlcGwiLCJnaXZlbl9uYW1lIjoiU2VwbCIsImxvY2FsZSI6ImVuIiwiZmFtaWx5X25hbWUiOiJBZG1pbiIsImVtYWlsIjoic2VwbEBzZXBsLmRlIn0.HZyG6n-BfpnaPAmcDoSEh0SadxUx-w4sEt2RVlQ9e5I`

type Client interface {
	GetSelectables(ctx context.Context, token string, criteria []models.DeviceGroupFilterCriteria, options *GetSelectablesOptions) ([]model.Selectable, int, error)
	GetBulkSelectablesV2(ctx context.Context, token string, bulk model.BulkRequestV2, options *GetBulkSelectablesOptions) (model.BulkResult, int, error)
	DeviceGroupHelper(ctx context.Context, token string, deviceIds []string, options *DeviceGroupHelperOptions) (model.DeviceGroupHelperResult, int, error)
}

type ClientImpl struct {
	baseUrl string

	//sends every request; Options.Timeout is set on it
	httpClient *http.Client
}

// Options configure a client from NewClient. The zero value sends with http.DefaultClient, without a
// timeout of its own.
type Options struct {
	// Timeout bounds every request including reading its answer; 0 leaves the bound to the context
	// of the call and to HttpClient
	Timeout time.Duration

	// HttpClient sends the requests; nil means http.DefaultClient. Timeout is set on a copy, the
	// given client stays unchanged.
	HttpClient *http.Client
}

// Client only implements GetSelectables, GetBulkSelectablesV2 and DeviceGroupHelper. Client should be extended if more API functionality is required
func NewClient(baseUrl string, options Options) Client {
	httpClient := options.HttpClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if options.Timeout > 0 {
		bounded := *httpClient
		bounded.Timeout = options.Timeout
		httpClient = &bounded
	}
	return &ClientImpl{baseUrl: baseUrl, httpClient: httpClient}
}

// do adds the trace context of req to its headers, so the called service continues the trace of the caller.
// Without an OpenTelemetry setup in the calling service the global propagator is a no-op.
func do[T any](c *ClientImpl, req *http.Request) (result T, code int, err error) {
	otel.GetTextMapPropagator().Inject(req.Context(), propagation.HeaderCarrier(req.Header))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return result, http.StatusInternalServerError, err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		temp, _ := io.ReadAll(resp.Body) //read error response end ensure that resp.Body is read to EOF
		return result, resp.StatusCode, fmt.Errorf("unexpected statuscode %v: %v", resp.StatusCode, string(temp))
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		_, _ = io.ReadAll(resp.Body) //ensure resp.Body is read to EOF
		return result, http.StatusInternalServerError, err
	}
	//the naked return this replaced left code at its zero value, so every successful call
	//answered 0 while every failing one answered a real status
	return result, resp.StatusCode, nil
}
