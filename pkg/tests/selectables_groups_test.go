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

package devices

import (
	"bytes"
	"context"
	"device-selection/pkg/configuration"
	"device-selection/pkg/controller"
	"device-selection/pkg/model"
	"device-selection/pkg/model/devicemodel"
	"device-selection/pkg/tests/environment"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"
)

func TestSelectableGroups(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deviceManagerUrl, semanticUrl, deviceRepoUrl, permSearchUrl, err := environment.New(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	c := &configuration.ConfigStruct{
		SemanticRepoUrl: semanticUrl,
		PermSearchUrl:   permSearchUrl,
		DeviceRepoUrl:   deviceRepoUrl,
		Debug:           true,
	}

	ctrl, err := controller.New(ctx, c)
	if err != nil {
		t.Error(err)
		return
	}

	deviceTypes := []devicemodel.DeviceType{
		{
			Id:            "lamp",
			Name:          "lamp",
			DeviceClassId: "lamp",
			Services: []devicemodel.Service{
				{Id: "s1", Name: "s1", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "s2", Name: "s2", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
				{Id: "s3", Name: "s3", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getState"}},
			},
		},
		{
			Id:            "event_lamp",
			Name:          "event_lamp",
			DeviceClassId: "lamp",
			Services: []devicemodel.Service{
				{Id: "se1", Name: "se1", Interaction: devicemodel.EVENT, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "se2", Name: "se2", Interaction: devicemodel.EVENT, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
				{Id: "se3", Name: "se3", Interaction: devicemodel.EVENT, AspectIds: []string{"device", "light"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getState"}},
			},
		},
		{
			Id:            "colorlamp",
			Name:          "colorlamp",
			DeviceClassId: "lamp",
			Services: []devicemodel.Service{
				{Id: "s4", Name: "s4", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "s5", Name: "s5", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
				{Id: "s6", Name: "s6", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getState"}},
				{Id: "s7", Name: "s7", Interaction: devicemodel.REQUEST, AspectIds: []string{"light"}, FunctionIds: []string{"setColor"}},
				{Id: "s8", Name: "s8", Interaction: devicemodel.REQUEST, AspectIds: []string{"light"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getColor"}},
			},
		},
		{
			Id:            "plug",
			Name:          "plug",
			DeviceClassId: "plug",
			Services: []devicemodel.Service{
				{Id: "s9", Name: "s9", Interaction: devicemodel.REQUEST, AspectIds: []string{"device"}, FunctionIds: []string{"setOn"}},
				{Id: "s10", Name: "s10", Interaction: devicemodel.REQUEST, AspectIds: []string{"device"}, FunctionIds: []string{"setOff"}},
				{Id: "s11", Name: "s11", Interaction: devicemodel.REQUEST, AspectIds: []string{"device"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getState"}},
			},
		},
	}

	devicesInstances := []devicemodel.Device{
		{
			Id:           "elamp",
			Name:         "elamp",
			LocalId:      "elamp",
			DeviceTypeId: "event_lamp",
		},
		{
			Id:           "lamp1",
			Name:         "lamp1",
			LocalId:      "lamp1",
			DeviceTypeId: "lamp",
		},
		{
			Id:           "lamp2",
			Name:         "lamp2",
			LocalId:      "lamp2",
			DeviceTypeId: "lamp",
		},
		{
			Id:           "colorlamp1",
			Name:         "colorlamp1",
			LocalId:      "colorlamp1",
			DeviceTypeId: "colorlamp",
		},
		{
			Id:           "colorlamp2",
			Name:         "colorlamp2",
			LocalId:      "colorlamp2",
			DeviceTypeId: "colorlamp",
		},
		{
			Id:           "plug1",
			Name:         "plug1",
			LocalId:      "plug1",
			DeviceTypeId: "plug",
		},
		{
			Id:           "plug2",
			Name:         "plug2",
			LocalId:      "plug2",
			DeviceTypeId: "plug",
		},
	}

	deviceGroups := []devicemodel.DeviceGroup{
		{
			Id:                 "dg_lamp",
			Name:               "dg_lamp",
			BlockedInteraction: devicemodel.EVENT,
			Criteria: []devicemodel.FilterCriteria{
				{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: ""},
				{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: ""},
				{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device"},
				{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light"},
			},
			DeviceIds: []string{"lamp1", "colorlamp1"},
		},
		{
			Id:                 "dg_colorlamp",
			Name:               "dg_colorlamp",
			BlockedInteraction: devicemodel.EVENT,
			Criteria: []model.FilterCriteria{
				{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: ""},
				{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: ""},
				{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: ""},
				{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device"},
				{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light"},
				{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light"},
			},
			DeviceIds: []string{"colorlamp1"},
		},
		{
			Id:                 "dg_plug",
			Name:               "dg_plug",
			BlockedInteraction: devicemodel.EVENT,
			Criteria: []model.FilterCriteria{
				{FunctionId: "setOn", DeviceClassId: "plug", AspectId: ""},
				{FunctionId: "setOff", DeviceClassId: "plug", AspectId: ""},
				{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device"},
			},
			DeviceIds: []string{"plug1", "plug2"},
		},
		{
			Id:                 "dg_event_lamp",
			Name:               "eventlamps",
			BlockedInteraction: devicemodel.REQUEST,
			Criteria: []model.FilterCriteria{
				{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light"},
				{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device"},
			},
			DeviceIds: []string{"elamp"},
		},
	}

	t.Run("create device-types", testCreateDeviceTypes(deviceManagerUrl, deviceTypes))
	t.Run("create devices", testCreateDevices(deviceManagerUrl, devicesInstances))
	t.Run("create devices-groups", testCreateDeviceGroups(deviceManagerUrl, deviceGroups))

	time.Sleep(5 * time.Second)

	t.Run("lamp on/off", testCheckSelection(ctrl, model.FilterCriteriaAndSet{
		{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: ""},
		{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: ""},
	}, devicemodel.EVENT, false, []model.Selectable{
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "colorlamp1",
					Name:         "colorlamp1",
					DeviceTypeId: "colorlamp",
				},
			},
			Services: []devicemodel.Service{
				{Id: "s4", Name: "s4", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "s5", Name: "s5", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
			},
		},
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				},
			},
			Services: []devicemodel.Service{
				{Id: "s4", Name: "s4", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "s5", Name: "s5", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
			},
		},
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "lamp1",
					Name:         "lamp1",
					DeviceTypeId: "lamp",
				},
			},
			Services: []devicemodel.Service{
				{Id: "s1", Name: "s1", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "s2", Name: "s2", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
			},
		},
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				},
			},
			Services: []devicemodel.Service{
				{Id: "s1", Name: "s1", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "s2", Name: "s2", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
			},
		},
	}))

	t.Run("lamp on/off with group", testCheckSelection(ctrl, model.FilterCriteriaAndSet{
		{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: ""},
		{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: ""},
	}, devicemodel.EVENT, true, []model.Selectable{
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "colorlamp1",
					Name:         "colorlamp1",
					DeviceTypeId: "colorlamp",
				},
			},
			Services: []devicemodel.Service{
				{Id: "s4", Name: "s4", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "s5", Name: "s5", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
			},
		},
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				},
			},
			Services: []devicemodel.Service{
				{Id: "s4", Name: "s4", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "s5", Name: "s5", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
			},
		},
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "lamp1",
					Name:         "lamp1",
					DeviceTypeId: "lamp",
				},
			},
			Services: []devicemodel.Service{
				{Id: "s1", Name: "s1", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "s2", Name: "s2", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
			},
		},
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				},
			},
			Services: []devicemodel.Service{
				{Id: "s1", Name: "s1", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "s2", Name: "s2", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
			},
		},
		{
			DeviceGroup: &model.DeviceGroup{
				Id:   "dg_colorlamp",
				Name: "dg_colorlamp",
			},
		},
		{
			DeviceGroup: &model.DeviceGroup{
				Id:   "dg_lamp",
				Name: "dg_lamp",
			},
		},
	}))

	t.Run("lamp get color with group", testCheckSelection(ctrl, model.FilterCriteriaAndSet{
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light"},
	}, devicemodel.EVENT, true, []model.Selectable{
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "colorlamp1",
					Name:         "colorlamp1",
					DeviceTypeId: "colorlamp",
				},
			},
			Services: []devicemodel.Service{
				{Id: "s8", Name: "s8", Interaction: devicemodel.REQUEST, AspectIds: []string{"light"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getColor"}},
			},
		},
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				},
			},
			Services: []devicemodel.Service{
				{Id: "s8", Name: "s8", Interaction: devicemodel.REQUEST, AspectIds: []string{"light"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getColor"}},
			},
		},
		{
			DeviceGroup: &model.DeviceGroup{
				Id:   "dg_colorlamp",
				Name: "dg_colorlamp",
			},
		},
	}))

	t.Run("plug on/off with group", testCheckSelection(ctrl, model.FilterCriteriaAndSet{
		{FunctionId: "setOn", DeviceClassId: "plug", AspectId: ""},
		{FunctionId: "setOff", DeviceClassId: "plug", AspectId: ""},
	}, devicemodel.EVENT, true, []model.Selectable{
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "plug1",
					Name:         "plug1",
					DeviceTypeId: "plug",
				},
			},
			Services: []devicemodel.Service{
				{Id: "s9", Name: "s9", Interaction: devicemodel.REQUEST, AspectIds: []string{"device"}, FunctionIds: []string{"setOn"}},
				{Id: "s10", Name: "s10", Interaction: devicemodel.REQUEST, AspectIds: []string{"device"}, FunctionIds: []string{"setOff"}},
			},
		},
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				},
			},
			Services: []devicemodel.Service{
				{Id: "s9", Name: "s9", Interaction: devicemodel.REQUEST, AspectIds: []string{"device"}, FunctionIds: []string{"setOn"}},
				{Id: "s10", Name: "s10", Interaction: devicemodel.REQUEST, AspectIds: []string{"device"}, FunctionIds: []string{"setOff"}},
			},
		},
		{
			DeviceGroup: &model.DeviceGroup{
				Id:   "dg_plug",
				Name: "dg_plug",
			},
		},
	}))

}

const token = `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJyb2xlcyI6WyJjcmVhdGUtcmVhbG0iLCJvZmZsaW5lX2FjY2VzcyIsImFkbWluIiwiZGV2ZWxvcGVyIiwidW1hX2F1dGhvcml6YXRpb24iLCJ1c2VyIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJjcmVhdGUtcmVhbG0iLCJvZmZsaW5lX2FjY2VzcyIsImFkbWluIiwiZGV2ZWxvcGVyIiwidW1hX2F1dGhvcml6YXRpb24iLCJ1c2VyIl19fQ.s-bPUbJc8e04WmwD7ei_XGRjAMuRKkpfqmgQKXjjqqI`

func testCheckSelection(ctrl *controller.Controller, criteria model.FilterCriteriaAndSet, interaction devicemodel.Interaction, includeGroups bool, expectedResult []model.Selectable) func(t *testing.T) {
	return func(t *testing.T) {
		result, err, _ := ctrl.GetFilteredDevices(token, criteria, nil, interaction, includeGroups)
		if err != nil {
			t.Error(err)
			return
		}
		normalizeTestSelectables(&result)
		normalizeTestSelectables(&expectedResult)
		if !reflect.DeepEqual(result, expectedResult) {
			resultJson, _ := json.Marshal(result)
			expectedJson, _ := json.Marshal(expectedResult)
			t.Error(string(resultJson), "\n", string(expectedJson))
		}
	}
}

func normalizeTestSelectables(selectables *[]model.Selectable) {
	for i, v := range *selectables {
		normalizeTestSelectable(&v)
		(*selectables)[i] = v
	}
	sort.SliceStable(*selectables, func(i, j int) bool {
		iName := ""
		if (*selectables)[i].Device != nil {
			iName = (*selectables)[i].Device.Name
		}
		if (*selectables)[i].DeviceGroup != nil {
			iName = (*selectables)[i].DeviceGroup.Name
		}

		jName := ""
		if (*selectables)[j].Device != nil {
			jName = (*selectables)[j].Device.Name
		}
		if (*selectables)[j].DeviceGroup != nil {
			jName = (*selectables)[j].DeviceGroup.Name
		}
		return iName < jName
	})
}

func normalizeTestSelectable(selectable *model.Selectable) {
	if selectable.Device != nil {
		selectable.Device.Id = ""
		selectable.Device.LocalId = ""
		selectable.Device.Creator = ""
		selectable.Device.Permissions = model.Permissions{}
		selectable.Device.Shared = false

		sort.SliceStable(selectable.Services, func(i, j int) bool {
			iName := selectable.Services[i].Name
			jName := selectable.Services[j].Name
			return iName < jName
		})
	}
}

func testCreateDeviceGroups(managerUrl string, groups []devicemodel.DeviceGroup) func(t *testing.T) {
	return func(t *testing.T) {
		for _, group := range groups {
			buff := new(bytes.Buffer)
			err := json.NewEncoder(buff).Encode(group)
			if err != nil {
				t.Error(err)
				return
			}
			req, err := http.NewRequest("POST", managerUrl+"/device-groups", buff)
			if err != nil {
				t.Error(err)
				return
			}
			req.Header.Set("Authorization", token)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Error(err)
				return
			}
			if resp.StatusCode != 200 {
				temp, _ := ioutil.ReadAll(resp.Body)
				t.Error(resp.StatusCode, string(temp))
			}
		}
	}
}

func testCreateDevices(managerUrl string, devices []devicemodel.Device) func(t *testing.T) {
	return func(t *testing.T) {
		for _, device := range devices {
			buff := new(bytes.Buffer)
			err := json.NewEncoder(buff).Encode(device)
			if err != nil {
				t.Error(err)
				return
			}
			req, err := http.NewRequest("POST", managerUrl+"/devices", buff)
			if err != nil {
				t.Error(err)
				return
			}
			req.Header.Set("Authorization", token)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Error(err)
				return
			}
			if resp.StatusCode != 200 {
				temp, _ := ioutil.ReadAll(resp.Body)
				t.Error(resp.StatusCode, string(temp))
				return
			}
		}
	}
}

func testCreateDeviceTypes(managerUrl string, deviceTypes []devicemodel.DeviceType) func(t *testing.T) {
	return func(t *testing.T) {
		for _, deviceType := range deviceTypes {
			buff := new(bytes.Buffer)
			err := json.NewEncoder(buff).Encode(deviceType)
			if err != nil {
				t.Error(err)
				return
			}
			req, err := http.NewRequest("POST", managerUrl+"/device-types", buff)
			if err != nil {
				t.Error(err)
				return
			}
			req.Header.Set("Authorization", token)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Error(err)
				return
			}
			if resp.StatusCode != 200 {
				temp, _ := ioutil.ReadAll(resp.Body)
				t.Error(resp.StatusCode, string(temp))
				return
			}
		}
	}
}