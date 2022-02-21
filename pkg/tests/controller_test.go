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

package tests

import (
	"context"
	"device-selection/pkg/configuration"
	"device-selection/pkg/controller"
	"device-selection/pkg/model"
	"device-selection/pkg/model/devicemodel"
	"device-selection/pkg/tests/environment/legacy"
	"device-selection/pkg/tests/helper"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestGetFilteredDeviceTypes(t *testing.T) {

	mux := sync.Mutex{}
	calls := []string{}

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.Lock()
		defer mux.Unlock()
		calls = append(calls, r.URL.Path+"?"+r.URL.RawQuery)
		json.NewEncoder(w).Encode([]devicemodel.DeviceType{{Id: "dt1", Name: "dt1name"}})
	}))

	defer mock.Close()

	c := &configuration.ConfigStruct{
		DeviceRepoUrl: mock.URL,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	repo, err := controller.New(ctx, c)
	if err != nil {
		t.Error(err)
		return
	}

	_, err, _ = repo.GetFilteredDeviceTypes(helper.AdminJwt, DeviceDescriptions{{
		CharacteristicId: "chid1",
		Function:         devicemodel.Function{Id: "fid"},
		DeviceClass:      nil,
		Aspect:           nil,
	}}.ToFilter(), nil)

	if err != nil {
		t.Error(err)
		return
	}

	dt, err, _ := repo.GetFilteredDeviceTypes(helper.AdminJwt, DeviceDescriptions{{
		CharacteristicId: "chid1",
		Function:         devicemodel.Function{Id: "fid"},
		DeviceClass:      &devicemodel.DeviceClass{Id: "dc1"},
		Aspect:           &devicemodel.Aspect{Id: "a1"},
	}}.ToFilter(), nil)

	if err != nil {
		t.Error(err)
		return
	}

	if len(dt) != 1 || dt[0].Name != "dt1name" || dt[0].Id != "dt1" {
		t.Error(dt)
		return
	}

	mux.Lock()
	defer mux.Unlock()
	if !reflect.DeepEqual(calls, []string{
		"/device-types?filter=" + url.QueryEscape(`[{"function_id":"fid","aspect_id":"","device_class_id":""}]`),
		"/device-types?filter=" + url.QueryEscape(`[{"function_id":"fid","aspect_id":"a1","device_class_id":"dc1"}]`),
	}) {
		temp, _ := json.Marshal(calls)
		t.Error(string(temp))
	}
}

func TestGetFilteredDevices(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deviceTypes := []devicemodel.DeviceType{
		{Id: "dt1", Name: "dt1name", DeviceClassId: "dc1", Services: []devicemodel.Service{
			testService("11", "pid", devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION),
			testService("11_b", "mqtt", devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION),
			testService("12", "pid", devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION),
		}},
		{Id: "dt2", Name: "dt2name", DeviceClassId: "dc1", Services: []devicemodel.Service{
			testService("21", "pid", devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION),
			testService("22", "pid", devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION),
		}},
		{Id: "dt3", Name: "dt1name", DeviceClassId: "dc1", Services: []devicemodel.Service{
			testService("31", "mqtt", devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION),
			testService("32", "mqtt", devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION),
		}},
		{Id: "dt4", Name: "dt2name", DeviceClassId: "dc1", Services: []devicemodel.Service{
			testService("41", "mqtt", devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION),
			testService("42", "mqtt", devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION),
		}},
	}

	devices := []devicemodel.Device{
		{Id: "1", Name: "1", DeviceTypeId: "dt1"},
		{Id: "2", Name: "2", DeviceTypeId: "dt2"},
		{Id: "3", Name: "3", DeviceTypeId: "dt3"},
		{Id: "4", Name: "4", DeviceTypeId: "dt4"},
	}

	concepts := []devicemodel.Concept{
		{
			Id: "concept",
		},
	}

	functions := []devicemodel.Function{
		{
			Id:        devicemodel.MEASURING_FUNCTION_PREFIX + "_1",
			ConceptId: "concept",
		},
		{
			Id:        devicemodel.CONTROLLING_FUNCTION_PREFIX + "_1",
			ConceptId: "concept",
		},
	}

	aspects := []devicemodel.Aspect{
		{
			Id:   "a1",
			Name: "a1",
		},
	}

	managerurl, repourl, searchurl, err := helper.EnvWithDevices(ctx, wg, deviceTypes, devices)

	for _, concept := range concepts {
		err = helper.SetConcept(managerurl, concept)
		if err != nil {
			t.Error(err)
			return
		}
	}

	for _, f := range functions {
		err = helper.SetFunction(managerurl, f)
		if err != nil {
			t.Error(err)
			return
		}
	}

	for _, a := range aspects {
		err = helper.SetAspect(managerurl, a)
		if err != nil {
			t.Error(err)
			return
		}
	}

	c := &configuration.ConfigStruct{
		PermSearchUrl: searchurl,
		DeviceRepoUrl: repourl,
	}

	repo, err := controller.New(ctx, c)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(2 * time.Second)

	d, err, _ := repo.GetFilteredDevices(helper.AdminJwt, DeviceDescriptions{{
		CharacteristicId: "chid1",
		Function:         devicemodel.Function{Id: devicemodel.MEASURING_FUNCTION_PREFIX + "_1"},
		DeviceClass:      &devicemodel.DeviceClass{Id: "dc1"},
		Aspect:           &devicemodel.Aspect{Id: "a1"},
	}}.ToFilter(), []string{"mqtt"}, "", false, false, nil)

	if err != nil {
		t.Error(err)
		return
	}

	if len(d) != 1 || d[0].Device.Name != "1" || d[0].Device.Id != "1" || len(d[0].Services) != 1 || d[0].Services[0].Id != "11" {
		t.Error(len(d), d)
		return
	}
}

func testService(id string, protocolId string, functionType string) devicemodel.Service {
	result := legacy.Service{
		Id:         id,
		LocalId:    id + "_l",
		Name:       id + "_name",
		AspectIds:  []string{"a1"},
		ProtocolId: protocolId,
	}
	if functionType == devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION {
		result.FunctionIds = []string{devicemodel.MEASURING_FUNCTION_PREFIX + "_1"}
	} else {
		result.FunctionIds = []string{devicemodel.CONTROLLING_FUNCTION_PREFIX + "_1"}
	}
	return legacy.FromLegacyService(result)
}

type DeviceDescriptions []DeviceDescription
type DeviceDescription struct {
	CharacteristicId string                   `json:"characteristic_id"`
	Function         devicemodel.Function     `json:"function"`
	DeviceClass      *devicemodel.DeviceClass `json:"device_class,omitempty"`
	Aspect           *devicemodel.Aspect      `json:"aspect,omitempty"`
}

func (this DeviceDescriptions) ToFilter() (result model.FilterCriteriaAndSet) {
	for _, element := range this {
		newElement := devicemodel.FilterCriteria{
			FunctionId: element.Function.Id,
		}
		if element.DeviceClass != nil {
			newElement.DeviceClassId = element.DeviceClass.Id
		}
		if element.Aspect != nil {
			newElement.AspectId = element.Aspect.Id
		}
		if !IsZero(element) {
			result = append(result, newElement)
		}
	}
	return
}

func IsZero(x interface{}) bool {
	return x == reflect.Zero(reflect.TypeOf(x)).Interface()
}

type TestPermSearchDevice struct {
	Id          string            `json:"id"`
	LocalId     string            `json:"local_id,omitempty"`
	Name        string            `json:"name,omitempty"`
	DeviceType  string            `json:"device_type_id,omitempty"`
	Permissions model.Permissions `json:"permissions"`
	Shared      bool              `json:"shared"`
	Creator     string            `json:"creator"`
}
