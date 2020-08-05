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
	"device-selection/pkg/api"
	"device-selection/pkg/configuration"
	"device-selection/pkg/devicemodel"
	"device-selection/pkg/devices"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sync"
	"testing"
)

func TestApiSimpleGet(t *testing.T) {
	mux, calls, semanticmock, selectionApi, err := testenv()
	if err != nil {
		t.Error(err)
		return
	}
	defer semanticmock.Close()
	defer selectionApi.Close()

	result := []devicemodel.Selectable{}

	t.Run("send simple request", sendSimpleRequest(selectionApi.URL, &result, devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION+"_1", "dc1", "a1", "mqtt"))

	t.Run("check result", func(t *testing.T) {
		if len(result) != 1 || result[0].Device.Name != "1" || result[0].Device.Id != "1" || len(result[0].Services) != 1 || result[0].Services[0].Id != "11" {
			t.Error(result)
			return
		}
	})

	t.Run("check semantic calls", func(t *testing.T) {
		mux.Lock()
		defer mux.Unlock()
		expected := []string{
			"/device-types?filter=" + url.QueryEscape(`[{"function_id":"`+devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION+`_1","device_class_id":"dc1","aspect_id":"a1"}]`),
		}
		if !reflect.DeepEqual(*calls, expected) {
			actualStr, _ := json.Marshal(calls)
			expectedStr, _ := json.Marshal(expected)
			t.Error(string(actualStr), string(expectedStr))
		}
	})
}

func TestApiJsonGet(t *testing.T) {
	mux, calls, semanticmock, selectionApi, err := testenv()
	if err != nil {
		t.Error(err)
		return
	}
	defer semanticmock.Close()
	defer selectionApi.Close()

	result := []devicemodel.Selectable{}

	t.Run("send json request", sendJsonRequest(selectionApi.URL, &result, devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION+"_1", "dc1", "a1", "mqtt"))

	t.Run("check result", func(t *testing.T) {
		if len(result) != 1 || result[0].Device.Name != "1" || result[0].Device.Id != "1" || len(result[0].Services) != 1 || result[0].Services[0].Id != "11" {
			t.Error(result)
			return
		}
	})

	t.Run("check semantic calls", func(t *testing.T) {
		mux.Lock()
		defer mux.Unlock()
		expected := []string{
			"/device-types?filter=" + url.QueryEscape(`[{"function_id":"`+devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION+`_1","device_class_id":"dc1","aspect_id":"a1"}]`),
		}
		if !reflect.DeepEqual(*calls, expected) {
			actualStr, _ := json.Marshal(calls)
			expectedStr, _ := json.Marshal(expected)
			t.Error(string(actualStr), string(expectedStr))
		}
	})
}

func TestApiBase64Get(t *testing.T) {
	mux, calls, semanticmock, selectionApi, err := testenv()
	if err != nil {
		t.Error(err)
		return
	}
	defer semanticmock.Close()
	defer selectionApi.Close()

	result := []devicemodel.Selectable{}

	t.Run("send base64 request", sendBase64Request(selectionApi.URL, &result, devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION+"_1", "dc1", "a1", "mqtt"))

	t.Run("check result", func(t *testing.T) {
		if len(result) != 1 || result[0].Device.Name != "1" || result[0].Device.Id != "1" || len(result[0].Services) != 1 || result[0].Services[0].Id != "11" {
			t.Error(result)
			return
		}
	})

	t.Run("check semantic calls", func(t *testing.T) {
		mux.Lock()
		defer mux.Unlock()
		expected := []string{
			"/device-types?filter=" + url.QueryEscape(`[{"function_id":"`+devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION+`_1","device_class_id":"dc1","aspect_id":"a1"}]`),
		}
		if !reflect.DeepEqual(*calls, expected) {
			actualStr, _ := json.Marshal(calls)
			expectedStr, _ := json.Marshal(expected)
			t.Error(string(actualStr), string(expectedStr))
		}
	})
}

func sendSimpleRequest(apiurl string, result interface{}, functionId string, deviceClassId string, aspectId string, blockList string) func(t *testing.T) {
	return func(t *testing.T) {
		resp, err := http.Get(apiurl + "/selectables?function_id=" + url.QueryEscape(functionId) + "&device_class_id=" + url.QueryEscape(deviceClassId) + "&aspect_id=" + url.QueryEscape(aspectId) + "&filter_protocols=" + url.QueryEscape(blockList))
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != 200 {
			t.Error(resp.StatusCode)
			return
		}
		err = json.NewDecoder(resp.Body).Decode(result)
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func sendJsonRequest(apiurl string, result interface{}, functionId string, deviceClassId string, aspectId string, blockList string) func(t *testing.T) {
	return func(t *testing.T) {
		jsonStr, err := json.Marshal(devicemodel.DeviceTypesFilter{{
			FunctionId:    functionId,
			DeviceClassId: deviceClassId,
			AspectId:      aspectId,
		}})
		resp, err := http.Get(apiurl + "/selectables?json=" + url.QueryEscape(string(jsonStr)) + "&filter_protocols=" + url.QueryEscape(blockList))
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != 200 {
			t.Error(resp.StatusCode)
			return
		}
		err = json.NewDecoder(resp.Body).Decode(result)
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func sendBase64Request(apiurl string, result interface{}, functionId string, deviceClassId string, aspectId string, blockList string) func(t *testing.T) {
	return func(t *testing.T) {
		jsonStr, err := json.Marshal(devicemodel.DeviceTypesFilter{{
			FunctionId:    functionId,
			DeviceClassId: deviceClassId,
			AspectId:      aspectId,
		}})
		b64Str := base64.StdEncoding.EncodeToString(jsonStr)
		resp, err := http.Get(apiurl + "/selectables?base64=" + b64Str + "&filter_protocols=" + url.QueryEscape(blockList))
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != 200 {
			t.Error(resp.StatusCode)
			return
		}
		err = json.NewDecoder(resp.Body).Decode(result)
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func testenv() (mux *sync.Mutex, semanticCalls *[]string, semanticmock *httptest.Server, selectionApi *httptest.Server, err error) {
	mux = &sync.Mutex{}
	calls := []string{}
	semanticCalls = &calls

	semanticmock = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.Lock()
		defer mux.Unlock()
		calls = append(calls, r.URL.Path+"?"+r.URL.RawQuery)
		json.NewEncoder(w).Encode([]devicemodel.DeviceType{
			{Id: "dt1", Name: "dt1name", DeviceClass: devicemodel.DeviceClass{Id: "dc1"}, Services: []devicemodel.Service{
				testService("11", "pid", devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION),
				testService("11_b", "mqtt", devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION),
				testService("12", "pid", devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION),
			}},
			{Id: "dt2", Name: "dt2name", DeviceClass: devicemodel.DeviceClass{Id: "dc1"}, Services: []devicemodel.Service{
				testService("21", "pid", devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION),
				testService("22", "pid", devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION),
			}},
			{Id: "dt3", Name: "dt1name", DeviceClass: devicemodel.DeviceClass{Id: "dc1"}, Services: []devicemodel.Service{
				testService("31", "mqtt", devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION),
				testService("32", "mqtt", devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION),
			}},
			{Id: "dt4", Name: "dt2name", DeviceClass: devicemodel.DeviceClass{Id: "dc1"}, Services: []devicemodel.Service{
				testService("41", "mqtt", devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION),
				testService("42", "mqtt", devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION),
			}},
		})
	}))

	searchmock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/jwt/select/devices/device_type_id/dt1/x" {
			json.NewEncoder(w).Encode([]devices.PermSearchDevice{
				{Id: "1", Name: "1", DeviceType: "dt1"},
			})
		}
		if r.URL.Path == "/jwt/select/devices/device_type_id/dt2/x" {
			json.NewEncoder(w).Encode([]devices.PermSearchDevice{
				{Id: "2", Name: "2", DeviceType: "dt2"},
			})
		}
		if r.URL.Path == "/jwt/select/devices/device_type_id/dt3/x" {
			json.NewEncoder(w).Encode([]devices.PermSearchDevice{
				{Id: "3", Name: "3", DeviceType: "dt3"},
			})
		}
		if r.URL.Path == "/jwt/select/devices/device_type_id/dt4/x" {
			json.NewEncoder(w).Encode([]devices.PermSearchDevice{
				{Id: "4", Name: "4", DeviceType: "dt4"},
			})
		}
	}))

	c := &configuration.ConfigStruct{
		SemanticRepoUrl: semanticmock.URL,
		PermSearchUrl:   searchmock.URL,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	repo, err := devices.New(ctx, c)
	if err != nil {
		searchmock.Close()
		selectionApi.Close()
		return mux, semanticCalls, semanticmock, selectionApi, err
	}

	router := api.Router(c, repo)
	selectionApi = httptest.NewServer(router)

	return
}

func testService(id string, protocolId string, functionType string) devicemodel.Service {
	return devicemodel.Service{
		Id:         id,
		LocalId:    id + "_l",
		Name:       id + "_name",
		Aspects:    []devicemodel.Aspect{{Id: "a1"}},
		ProtocolId: protocolId,
		Functions:  []devicemodel.Function{{Id: functionType + "_1", RdfType: functionType}},
	}
}

type DeviceDescriptions []DeviceDescription
type DeviceDescription struct {
	CharacteristicId string                   `json:"characteristic_id"`
	Function         devicemodel.Function     `json:"function"`
	DeviceClass      *devicemodel.DeviceClass `json:"device_class,omitempty"`
	Aspect           *devicemodel.Aspect      `json:"aspect,omitempty"`
}

func (this DeviceDescriptions) ToFilter() (result devicemodel.DeviceTypesFilter) {
	for _, element := range this {
		newElement := devicemodel.DeviceTypeFilterElement{
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
