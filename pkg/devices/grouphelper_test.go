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
	"context"
	"device-selection/pkg/configuration"
	"device-selection/pkg/model"
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

	t.Run("empty list", testGroupHelper(repo, []string{}, []model.FilterCriteria{}, devicemodel.EVENT))

	t.Run("empty filter request", testGroupHelper(repo, []string{}, []model.FilterCriteria{}, devicemodel.REQUEST))

	t.Run("lamp1 unfiltered", testGroupHelper(repo, []string{"lamp1"}, []model.FilterCriteria{
		{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: ""},
		{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: ""},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device"},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light"},
	}, ""))

	t.Run("lamp1", testGroupHelper(repo, []string{"lamp1"}, []model.FilterCriteria{
		{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: ""},
		{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: ""},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device"},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light"},
	}, devicemodel.EVENT))

	t.Run("colorlamp1", testGroupHelper(repo, []string{"colorlamp1"}, []model.FilterCriteria{
		{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: ""},
		{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: ""},
		{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: ""},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device"},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light"},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light"},
	}, devicemodel.EVENT))

	t.Run("plug2", testGroupHelper(repo, []string{"plug2"}, []model.FilterCriteria{
		{FunctionId: "setOn", DeviceClassId: "plug", AspectId: ""},
		{FunctionId: "setOff", DeviceClassId: "plug", AspectId: ""},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device"},
	}, devicemodel.EVENT))

	t.Run("lamp1 colorlamp1", testGroupHelper(repo, []string{"lamp1", "colorlamp1"}, []model.FilterCriteria{
		{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: ""},
		{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: ""},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device"},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light"},
	}, devicemodel.EVENT))

	t.Run("lamp1 colorlamp1 plug1", testGroupHelper(repo, []string{"lamp1", "colorlamp1", "plug1"}, []model.FilterCriteria{
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device"},
	}, devicemodel.EVENT))
}

func testGroupHelper(repo *Devices, deviceIds []string, expectedResult []model.FilterCriteria, filteredInteraction devicemodel.Interaction) func(t *testing.T) {
	return func(t *testing.T) {
		dtCache := &map[string]devicemodel.DeviceType{}
		dCache := &map[string]devicemodel.Device{}
		result, err, code := repo.getDeviceGroupCriteria("test-token", dtCache, dCache, filteredInteraction, deviceIds)
		if err != nil {
			t.Error(err, code)
			return
		}
		result = normalizeCriteria(result)
		expectedResult = normalizeCriteria(expectedResult)
		if !reflect.DeepEqual(result, expectedResult) {
			t.Error(result, expectedResult)
		}
	}
}

func normalizeCriteria(criteria []model.FilterCriteria) []model.FilterCriteria {
	sort.SliceStable(criteria, func(i, j int) bool {
		return criteria[i].AspectId < criteria[j].AspectId
	})
	sort.SliceStable(criteria, func(i, j int) bool {
		return criteria[i].FunctionId < criteria[j].FunctionId
	})
	return criteria
}

func grouphelpertestenv(deviceTypes []devicemodel.DeviceType, deviceInstances []devicemodel.Device) (semanticmock *httptest.Server, searchmock *httptest.Server, devicerepomock *httptest.Server, repo *Devices, err error) {

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
