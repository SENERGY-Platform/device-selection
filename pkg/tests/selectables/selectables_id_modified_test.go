/*
 * Copyright 2020 InfAI (CC SES)
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

package selectables

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/device-selection/pkg/controller/idmodifier"
	"github.com/SENERGY-Platform/device-selection/pkg/model"
	"github.com/SENERGY-Platform/device-selection/pkg/model/devicemodel"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/helper"
	"net/url"
	"sync"
	"testing"
)

func TestIdModifiedSelectables(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dtId := "dtId"
	dId := "dId"
	deviceTypes := []devicemodel.DeviceType{
		{
			Id:            dtId,
			Name:          "dtsg",
			DeviceClassId: "toggle",
			ServiceGroups: []devicemodel.ServiceGroup{{Key: "sg1", Name: "sg1"}, {Key: "sg2", Name: "sg2"}},
			Services: []devicemodel.Service{
				{
					Id:              "plug1",
					ServiceGroupKey: "sg1",
					Interaction:     devicemodel.REQUEST,
					Outputs: []devicemodel.Content{
						{
							Id: "so1",
							ContentVariable: devicemodel.ContentVariable{
								Id:               "state1",
								Name:             "state",
								FunctionId:       devicemodel.MEASURING_FUNCTION_PREFIX + "getPlugState",
								AspectId:         "plug",
								CharacteristicId: "plug-state-characteristic",
							},
						},
					},
				},
				{
					Id:              "plug2",
					ServiceGroupKey: "sg2",
					Interaction:     devicemodel.REQUEST,
					Outputs: []devicemodel.Content{
						{
							Id: "so2",
							ContentVariable: devicemodel.ContentVariable{
								Id:               "state2",
								Name:             "state",
								FunctionId:       devicemodel.MEASURING_FUNCTION_PREFIX + "getPlugState",
								AspectId:         "plug",
								CharacteristicId: "plug-state-characteristic",
							},
						},
					},
				},
				{
					Id:          "plugs",
					Interaction: devicemodel.REQUEST,
					Outputs: []devicemodel.Content{
						{
							Id: "so3",
							ContentVariable: devicemodel.ContentVariable{
								Id:               "states",
								Name:             "states",
								FunctionId:       devicemodel.MEASURING_FUNCTION_PREFIX + "getPlugStates",
								AspectId:         "plug",
								CharacteristicId: "plug-state-list-characteristic",
							},
						},
					},
				},
			},
		},
	}
	aspects := []devicemodel.Aspect{{Id: "plug"}}
	functions := []devicemodel.Function{
		{Id: devicemodel.MEASURING_FUNCTION_PREFIX + "getPlugState"},
		{Id: devicemodel.MEASURING_FUNCTION_PREFIX + "getPlugStates"},
	}
	devices := []devicemodel.Device{{
		Id:           dId,
		LocalId:      dId,
		Name:         "sg_device",
		DeviceTypeId: dtId,
		OwnerId:      helper.JwtSubject,
	}}

	_, _, _, selectionurl, err := helper.EnvWithMetadata(ctx, wg, deviceTypes, devices, aspects, functions)
	if err != nil {
		t.Error(err)
		return
	}

	criteria := model.FilterCriteriaAndSet{
		{
			Interaction: string(devicemodel.REQUEST),
			FunctionId:  devicemodel.MEASURING_FUNCTION_PREFIX + "getPlugState",
			AspectId:    "plug",
		},
	}
	criteriaJson, err := json.Marshal(criteria)
	if err != nil {
		t.Error(err)
		return
	}
	criteriaQuery := url.QueryEscape(string(criteriaJson))

	expectedSelectionWithoutModify := model.Selectable{
		Device: &model.PermSearchDevice{
			Device:      devices[0],
			DisplayName: devices[0].Name,
			Permissions: model.Permissions{
				R: true,
				W: true,
				X: true,
				A: true,
			},
			Shared:  false,
			Creator: "dd69ea0d-f553-4336-80f3-7f4567f85c7b",
		},
		Services: deviceTypes[0].Services[:2],
		ServicePathOptions: map[string][]model.PathOption{
			"plug1": {
				{
					Path:             "state",
					CharacteristicId: "plug-state-characteristic",
					AspectNode: devicemodel.AspectNode{
						Id:            "plug",
						RootId:        "plug",
						ChildIds:      []string{},
						AncestorIds:   []string{},
						DescendentIds: []string{},
					},
					FunctionId:  "urn:infai:ses:measuring-function:getPlugState",
					Interaction: devicemodel.REQUEST,
				},
			},
			"plug2": {
				{
					Path:             "state",
					CharacteristicId: "plug-state-characteristic",
					AspectNode: devicemodel.AspectNode{
						Id:            "plug",
						RootId:        "plug",
						ChildIds:      []string{},
						AncestorIds:   []string{},
						DescendentIds: []string{},
					},
					FunctionId:  "urn:infai:ses:measuring-function:getPlugState",
					Interaction: devicemodel.REQUEST,
				},
			},
		},
	}

	expectedSg1Selection := model.Selectable{
		Device: &model.PermSearchDevice{
			Device:      testServiceGroupSelectModifyDevice(devices[0], "sg1", "sg1"),
			DisplayName: devices[0].Name + " sg1",
			Permissions: model.Permissions{
				R: true,
				W: true,
				X: true,
				A: true,
			},
			Shared:  false,
			Creator: "dd69ea0d-f553-4336-80f3-7f4567f85c7b",
		},
		Services: []devicemodel.Service{deviceTypes[0].Services[0]},
		ServicePathOptions: map[string][]model.PathOption{
			"plug1": {
				{
					Path:             "state",
					CharacteristicId: "plug-state-characteristic",
					AspectNode: devicemodel.AspectNode{
						Id:            "plug",
						RootId:        "plug",
						ChildIds:      []string{},
						AncestorIds:   []string{},
						DescendentIds: []string{},
					},
					FunctionId:  "urn:infai:ses:measuring-function:getPlugState",
					Interaction: devicemodel.REQUEST,
				},
			},
		},
	}
	expectedSg2Selection := model.Selectable{
		Device: &model.PermSearchDevice{
			Device:      testServiceGroupSelectModifyDevice(devices[0], "sg2", "sg2"),
			DisplayName: devices[0].Name + " sg2",
			Permissions: model.Permissions{
				R: true,
				W: true,
				X: true,
				A: true,
			},
			Shared:  false,
			Creator: "dd69ea0d-f553-4336-80f3-7f4567f85c7b",
		},
		Services: []devicemodel.Service{deviceTypes[0].Services[1]},
		ServicePathOptions: map[string][]model.PathOption{
			"plug2": {
				{
					Path:             "state",
					CharacteristicId: "plug-state-characteristic",
					AspectNode: devicemodel.AspectNode{
						Id:            "plug",
						RootId:        "plug",
						ChildIds:      []string{},
						AncestorIds:   []string{},
						DescendentIds: []string{},
					},
					FunctionId:  "urn:infai:ses:measuring-function:getPlugState",
					Interaction: devicemodel.REQUEST,
				},
			},
		},
	}

	t.Run("get selectables without modified", helper.TestRequest(selectionurl, "GET", "/v2/selectables?include_devices=true&json="+criteriaQuery, nil, 200, []model.Selectable{
		expectedSelectionWithoutModify,
	}))

	t.Run("query selectables without modified", helper.TestRequest(selectionurl, "POST", "/v2/query/selectables?include_devices=true", criteria, 200, []model.Selectable{
		expectedSelectionWithoutModify,
	}))

	t.Run("get selectables with modified", helper.TestRequest(selectionurl, "GET", "/v2/selectables?include_devices=true&include_id_modified=true&json="+criteriaQuery, nil, 200, []model.Selectable{
		expectedSelectionWithoutModify,
		expectedSg1Selection,
		expectedSg2Selection,
	}))

	t.Run("query selectables with modified", helper.TestRequest(selectionurl, "POST", "/v2/query/selectables?include_devices=true&include_id_modified=true", criteria, 200, []model.Selectable{
		expectedSelectionWithoutModify,
		expectedSg1Selection,
		expectedSg2Selection,
	}))
}

func testServiceGroupSelectModifyDevice(device devicemodel.Device, serviceGroupId string, serviceGroupName string) devicemodel.Device {
	result := device
	result.Id = result.Id + idmodifier.Seperator + idmodifier.EncodeModifierParameter(map[string][]string{"service_group_selection": {serviceGroupId}})
	result.Name = result.Name + " " + serviceGroupName
	result.DeviceTypeId = result.DeviceTypeId + idmodifier.Seperator + idmodifier.EncodeModifierParameter(map[string][]string{"service_group_selection": {serviceGroupId}})
	return result
}
