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
	"slices"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-selection/v2/pkg/configuration"
)

// functionsRepository answers the two function lists of device-repository; the measuring
// functions come late, so a read that does not wait for both returns without them, and with
// failMeasuring set they fail
func functionsRepository(t *testing.T, failMeasuring bool) string {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/controlling-functions":
			_, _ = w.Write([]byte(`[{"id":"controlling"}]`))
		case "/measuring-functions":
			time.Sleep(100 * time.Millisecond)
			if failMeasuring {
				http.Error(w, "unavailable", http.StatusInternalServerError)
				return
			}
			_, _ = w.Write([]byte(`[{"id":"measuring"}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	return server.URL
}

func TestGetFunctionsWaitsForBothLists(t *testing.T) {
	ctx := context.Background()
	ctrl, err := New(ctx, &configuration.ConfigStruct{LogLevel: "error", DeviceRepoUrl: functionsRepository(t, false)})
	if err != nil {
		t.Fatal(err)
	}
	functions, err := ctrl.GetFunctions(ctx, "Bearer token")
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{}
	for _, f := range functions {
		ids = append(ids, f.Id)
	}
	slices.Sort(ids)
	if !slices.Equal(ids, []string{"controlling", "measuring"}) {
		t.Errorf("functions %v", ids)
	}
}

func TestGetFunctionsReportsAFailedList(t *testing.T) {
	ctx := context.Background()
	ctrl, err := New(ctx, &configuration.ConfigStruct{LogLevel: "error", DeviceRepoUrl: functionsRepository(t, true)})
	if err != nil {
		t.Fatal(err)
	}
	if functions, err := ctrl.GetFunctions(ctx, "Bearer token"); err == nil {
		t.Errorf("a failed list has to fail the read, got %v", functions)
	}
}
