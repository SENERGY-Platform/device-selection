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

package controller

import (
	"context"
	"device-selection/pkg/configuration"
	"device-selection/pkg/model/devicemodel"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sort"
	"testing"
)

func TestGroupHelperCriteria(t *testing.T) {
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
				{Id: "se1", Name: "se1", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "se2", Name: "se2", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
				{Id: "se3", Name: "se3", Interaction: devicemodel.EVENT, AspectIds: []string{"device", "light"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getState"}},
			},
		},
		{
			Id:            "both_lamp",
			Name:          "both_lamp",
			DeviceClassId: "lamp",
			Services: []devicemodel.Service{
				{Id: "sb1", Name: "sb1", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "sb2", Name: "sb2", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
				{Id: "sb3", Name: "sb3", Interaction: devicemodel.EVENT_AND_REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getState"}},
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
			Id:           "blamp",
			Name:         "blamp",
			DeviceTypeId: "both_lamp",
		},
		{
			Id:           "elamp",
			Name:         "elamp",
			DeviceTypeId: "event_lamp",
		},
		{
			Id:           "lamp1",
			Name:         "lamp1",
			DeviceTypeId: "lamp",
		},
		{
			Id:           "lamp2",
			Name:         "lamp2",
			DeviceTypeId: "lamp",
		},
		{
			Id:           "colorlamp1",
			Name:         "colorlamp1",
			DeviceTypeId: "colorlamp",
		},
		{
			Id:           "colorlamp2",
			Name:         "colorlamp2",
			DeviceTypeId: "colorlamp",
		},
		{
			Id:           "plug1",
			Name:         "plug1",
			DeviceTypeId: "plug",
		},
		{
			Id:           "plug2",
			Name:         "plug2",
			DeviceTypeId: "plug",
		},
	}

	semanticmock, searchmock, devicerepomock, repo, err := grouphelpertestenv(deviceTypes, devicesInstances)
	if err != nil {
		t.Error(err)
		return
	}
	defer semanticmock.Close()
	defer searchmock.Close()
	defer devicerepomock.Close()

	t.Run("empty list", testGroupHelper(repo, []string{}, []devicemodel.DeviceGroupFilterCriteria{}))

	t.Run("lamp1", testGroupHelper(repo, []string{"lamp1"}, []devicemodel.DeviceGroupFilterCriteria{
		{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
	}))

	t.Run("elamp", testGroupHelper(repo, []string{"elamp"}, []devicemodel.DeviceGroupFilterCriteria{
		{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
	}))

	t.Run("blamp", testGroupHelper(repo, []string{"blamp"}, []devicemodel.DeviceGroupFilterCriteria{
		{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
	}))

	t.Run("lamp1 blamp", testGroupHelper(repo, []string{"lamp1", "blamp"}, []devicemodel.DeviceGroupFilterCriteria{
		{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
	}))

	t.Run("lamp1 elamp", testGroupHelper(repo, []string{"lamp1", "elamp"}, []devicemodel.DeviceGroupFilterCriteria{
		{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
	}))

	t.Run("colorlamp1", testGroupHelper(repo, []string{"colorlamp1"}, []devicemodel.DeviceGroupFilterCriteria{
		{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
	}))

	t.Run("plug2", testGroupHelper(repo, []string{"plug2"}, []devicemodel.DeviceGroupFilterCriteria{
		{FunctionId: "setOn", DeviceClassId: "plug", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: "setOff", DeviceClassId: "plug", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
	}))

	t.Run("lamp1 colorlamp1", testGroupHelper(repo, []string{"lamp1", "colorlamp1"}, []devicemodel.DeviceGroupFilterCriteria{
		{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
	}))

	t.Run("lamp1 colorlamp1 plug1", testGroupHelper(repo, []string{"lamp1", "colorlamp1", "plug1"}, []devicemodel.DeviceGroupFilterCriteria{
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
	}))
}

func testGroupHelper(repo *Controller, deviceIds []string, expectedResult []devicemodel.DeviceGroupFilterCriteria) func(t *testing.T) {
	return func(t *testing.T) {
		dtCache := &map[string]devicemodel.DeviceType{}
		dCache := &map[string]devicemodel.Device{}
		result, err, code := repo.getDeviceGroupCriteria("test-token", dtCache, dCache, deviceIds)
		if err != nil {
			t.Error(err, code)
			return
		}
		result = normalizeCriteria(result)
		expectedResult = normalizeCriteria(expectedResult)
		if !reflect.DeepEqual(result, expectedResult) {
			resultJson, _ := json.Marshal(result)
			expectedJson, _ := json.Marshal(expectedResult)
			t.Error(string(resultJson), string(expectedJson))
		}
	}
}

func normalizeCriteria(criteria []devicemodel.DeviceGroupFilterCriteria) []devicemodel.DeviceGroupFilterCriteria {
	sort.SliceStable(criteria, func(i, j int) bool {
		return criteria[i].AspectId < criteria[j].AspectId
	})
	sort.SliceStable(criteria, func(i, j int) bool {
		return criteria[i].FunctionId < criteria[j].FunctionId
	})
	sort.SliceStable(criteria, func(i, j int) bool {
		return criteria[i].Interaction < criteria[j].Interaction
	})
	return criteria
}

func grouphelpertestenv(deviceTypes []devicemodel.DeviceType, deviceInstances []devicemodel.Device) (semanticmock *httptest.Server, searchmock *httptest.Server, devicerepomock *httptest.Server, repo *Controller, err error) {

	semanticmock = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("DEBUG: semantic call: " + r.URL.Path + "?" + r.URL.RawQuery)
		http.Error(w, "not implemented", http.StatusNotImplemented)
	}))

	devicerepomock = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, dt := range deviceTypes {
			if r.URL.Path == "/device-types/"+url.PathEscape(dt.Id) {
				json.NewEncoder(w).Encode(dt)
				return
			}
		}
		for _, d := range deviceInstances {
			if r.URL.Path == "/devices/"+url.PathEscape(d.Id) {
				json.NewEncoder(w).Encode(d)
				return
			}
		}
		log.Println("DEBUG: devicerepo call: " + r.URL.Path + "?" + r.URL.RawQuery)
		http.Error(w, "not implemented", http.StatusNotImplemented)
	}))

	searchmock = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("DEBUG: search call: " + r.URL.Path + "?" + r.URL.RawQuery)
		http.Error(w, "not implemented", http.StatusNotImplemented)
	}))

	c := &configuration.ConfigStruct{
		SemanticRepoUrl: semanticmock.URL,
		PermSearchUrl:   searchmock.URL,
		DeviceRepoUrl:   devicerepomock.URL,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	repo, err = New(ctx, c)
	if err != nil {
		searchmock.Close()
		semanticmock.Close()
		devicerepomock.Close()
		return semanticmock, searchmock, devicerepomock, repo, err
	}
	return
}
