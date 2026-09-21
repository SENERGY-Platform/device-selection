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

package bulk

import (
	"context"
	"net/http"
	"sync"
	"testing"

	"github.com/SENERGY-Platform/device-selection/v2/pkg/client"
	"github.com/SENERGY-Platform/device-selection/v2/pkg/model"
	"github.com/SENERGY-Platform/device-selection/v2/pkg/model/devicemodel"
	"github.com/SENERGY-Platform/device-selection/v2/pkg/tests/environment/legacy"
)

// The v2 bulk endpoint had no test of its own until the client gained a method for it. The
// test drives the client rather than a hand-built request, because that is what a consumer
// uses now and it covers both sides at once.
func TestClientBulkSelectablesV2(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, _, selectionurl, err := legacy.Testenv(ctx, wg)
	if err != nil {
		t.Fatal(err)
	}
	c := client.NewClient(selectionurl)

	matching := devicemodel.FilterCriteria{
		FunctionId:    devicemodel.MEASURING_FUNCTION_PREFIX + "_1",
		DeviceClassId: "dc1",
		AspectId:      "a1",
	}
	unmatched := devicemodel.FilterCriteria{
		FunctionId:    devicemodel.MEASURING_FUNCTION_PREFIX + "_1",
		DeviceClassId: "unknown",
		AspectId:      "a1",
	}
	request := model.BulkRequestV2{
		{Id: "matching", Criteria: model.FilterCriteriaAndSet{matching}, IncludeDevices: true},
		{Id: "unmatched", Criteria: model.FilterCriteriaAndSet{unmatched}, IncludeDevices: true},
	}

	result, code, err := c.GetBulkSelectablesV2(client.InternalAdminToken, request, &client.GetBulkSelectablesOptions{CompleteServices: true})
	if err != nil {
		t.Fatal(err, code)
	}
	if code != http.StatusOK {
		t.Fatal(code)
	}

	byId := map[string][]model.Selectable{}
	for _, element := range result {
		byId[element.Id] = element.Selectables
	}

	t.Run("answers every request element under the id it was given", func(t *testing.T) {
		if len(result) != len(request) {
			t.Fatal(len(result))
		}
		for _, element := range request {
			if _, ok := byId[element.Id]; !ok {
				t.Error("missing", element.Id)
			}
		}
	})

	//asserted against what the criteria ask for rather than against a fixed device count: the
	//v2 element has no protocol filter, so which devices match is a property of the test
	//environment, while what every match has to offer is a property of the endpoint
	t.Run("answers every selectable with the function and aspect that were asked for", func(t *testing.T) {
		selectables := byId["matching"]
		if len(selectables) == 0 {
			t.Fatal("expected the matching criteria to select something")
		}
		for _, selectable := range selectables {
			if selectable.Device == nil {
				t.Error("selectable without a device", selectable)
				continue
			}
			if len(selectable.Services) == 0 {
				t.Error(selectable.Device.Id, "has no service")
			}
			if len(selectable.ServicePathOptions) == 0 {
				t.Error(selectable.Device.Id, "has no path options")
			}
			for serviceId, options := range selectable.ServicePathOptions {
				for _, option := range options {
					if option.FunctionId != matching.FunctionId {
						t.Error(serviceId, "offers", option.FunctionId)
					}
					if option.AspectNode.Id != matching.AspectId {
						t.Error(serviceId, "offers aspect", option.AspectNode.Id)
					}
					if len(option.AspectNodes) != 1 || option.AspectNodes[0].Id != matching.AspectId {
						t.Error(serviceId, "aspect node list", option.AspectNodes)
					}
				}
			}
		}
	})

	t.Run("returns an empty selectable list for a criteria nothing matches", func(t *testing.T) {
		if len(byId["unmatched"]) != 0 {
			t.Error(byId["unmatched"])
		}
	})
}
