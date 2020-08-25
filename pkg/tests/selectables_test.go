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
	"device-selection/pkg/devices"
	"device-selection/pkg/model"
	"device-selection/pkg/model/devicemodel"
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
	mux, calls, semanticmock, searchmock, devicerepomock, selectionApi, err := testenv()
	if err != nil {
		t.Error(err)
		return
	}
	defer semanticmock.Close()
	defer selectionApi.Close()
	defer searchmock.Close()
	defer devicerepomock.Close()

	result := []model.Selectable{}

	t.Run("send simple request", sendSimpleRequest(selectionApi.URL, &result, devicemodel.MEASURING_FUNCTION_PREFIX+"_1", "dc1", "a1", "mqtt"))

	t.Run("check result", func(t *testing.T) {
		if len(result) != 1 {
			t.Error(len(result), result)
			return
		}
		if result[0].Device.Name != "1" ||
			result[0].Device.Id != "1" ||
			len(result[0].Services) != 1 ||
			result[0].Services[0].Id != "11" ||
			!result[0].Device.Permissions.R ||
			result[0].Device.Permissions.W ||
			!result[0].Device.Permissions.X ||
			result[0].Device.Permissions.A {
			t.Error(result)
			return
		}
	})

	t.Run("check semantic calls", func(t *testing.T) {
		mux.Lock()
		defer mux.Unlock()
		expected := []string{
			"/device-types?filter=" + url.QueryEscape(`[{"function_id":"`+devicemodel.MEASURING_FUNCTION_PREFIX+`_1","device_class_id":"dc1","aspect_id":"a1"}]`),
		}
		if !reflect.DeepEqual(*calls, expected) {
			actualStr, _ := json.Marshal(calls)
			expectedStr, _ := json.Marshal(expected)
			t.Error(string(actualStr), string(expectedStr))
		}
	})
}

func TestApiCompleteSimpledGet(t *testing.T) {
	mux, calls, semanticmock, searchmock, devicerepomock, selectionApi, err := testenv()
	if err != nil {
		t.Error(err)
		return
	}
	defer semanticmock.Close()
	defer selectionApi.Close()
	defer searchmock.Close()
	defer devicerepomock.Close()

	result := []model.Selectable{}

	t.Run("send simple request", sendCompletedSimpleRequest(selectionApi.URL, &result, devicemodel.MEASURING_FUNCTION_PREFIX+"_1", "dc1", "a1", "mqtt"))

	t.Run("check result", func(t *testing.T) {
		if len(result) != 1 ||
			result[0].Device.Name != "1" ||
			result[0].Device.Id != "1" ||
			len(result[0].Services) != 1 ||
			result[0].Services[0].Id != "11" ||
			len(result[0].Services[0].Outputs) != 1 ||
			result[0].Services[0].Outputs[0].Id != "content1" ||
			!result[0].Device.Permissions.R ||
			result[0].Device.Permissions.W ||
			!result[0].Device.Permissions.X ||
			result[0].Device.Permissions.A {
			t.Error(result)
			return
		}
	})

	t.Run("check semantic calls", func(t *testing.T) {
		mux.Lock()
		defer mux.Unlock()
		expected := []string{
			"/device-types?filter=" + url.QueryEscape(`[{"function_id":"`+devicemodel.MEASURING_FUNCTION_PREFIX+`_1","device_class_id":"dc1","aspect_id":"a1"}]`),
		}
		if !reflect.DeepEqual(*calls, expected) {
			actualStr, _ := json.Marshal(calls)
			expectedStr, _ := json.Marshal(expected)
			t.Error(string(actualStr), string(expectedStr))
		}
	})
}

func TestApiJsonGet(t *testing.T) {
	mux, calls, semanticmock, searchmock, devicerepomock, selectionApi, err := testenv()
	if err != nil {
		t.Error(err)
		return
	}
	defer semanticmock.Close()
	defer selectionApi.Close()
	defer searchmock.Close()
	defer devicerepomock.Close()

	result := []model.Selectable{}

	t.Run("send json request", sendJsonRequest(selectionApi.URL, &result, devicemodel.MEASURING_FUNCTION_PREFIX+"_1", "dc1", "a1", "mqtt"))

	t.Run("check result", func(t *testing.T) {
		if len(result) != 1 ||
			result[0].Device.Name != "1" ||
			result[0].Device.Id != "1" ||
			len(result[0].Services) != 1 ||
			result[0].Services[0].Id != "11" ||
			!result[0].Device.Permissions.R ||
			result[0].Device.Permissions.W ||
			!result[0].Device.Permissions.X ||
			result[0].Device.Permissions.A {
			t.Error(result)
			return
		}
	})

	t.Run("check semantic calls", func(t *testing.T) {
		mux.Lock()
		defer mux.Unlock()
		expected := []string{
			"/device-types?filter=" + url.QueryEscape(`[{"function_id":"`+devicemodel.MEASURING_FUNCTION_PREFIX+`_1","device_class_id":"dc1","aspect_id":"a1"}]`),
		}
		if !reflect.DeepEqual(*calls, expected) {
			actualStr, _ := json.Marshal(calls)
			expectedStr, _ := json.Marshal(expected)
			t.Error(string(actualStr), string(expectedStr))
		}
	})
}

func TestApiBase64Get(t *testing.T) {
	mux, calls, semanticmock, searchmock, devicerepomock, selectionApi, err := testenv()
	if err != nil {
		t.Error(err)
		return
	}
	defer semanticmock.Close()
	defer selectionApi.Close()
	defer searchmock.Close()
	defer devicerepomock.Close()

	result := []model.Selectable{}

	t.Run("send base64 request", sendBase64Request(selectionApi.URL, &result, devicemodel.MEASURING_FUNCTION_PREFIX+"_1", "dc1", "a1", "mqtt"))

	t.Run("check result", func(t *testing.T) {
		if len(result) != 1 ||
			result[0].Device.Name != "1" ||
			result[0].Device.Id != "1" ||
			len(result[0].Services) != 1 ||
			result[0].Services[0].Id != "11" ||
			!result[0].Device.Permissions.R ||
			result[0].Device.Permissions.W ||
			!result[0].Device.Permissions.X ||
			result[0].Device.Permissions.A {
			t.Error(result)
			return
		}
	})

	t.Run("check semantic calls", func(t *testing.T) {
		mux.Lock()
		defer mux.Unlock()
		expected := []string{
			"/device-types?filter=" + url.QueryEscape(`[{"function_id":"`+devicemodel.MEASURING_FUNCTION_PREFIX+`_1","device_class_id":"dc1","aspect_id":"a1"}]`),
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

func sendCompletedSimpleRequest(apiurl string, result interface{}, functionId string, deviceClassId string, aspectId string, blockList string) func(t *testing.T) {
	return func(t *testing.T) {
		resp, err := http.Get(apiurl + "/selectables?complete_services=true&function_id=" + url.QueryEscape(functionId) + "&device_class_id=" + url.QueryEscape(deviceClassId) + "&aspect_id=" + url.QueryEscape(aspectId) + "&filter_protocols=" + url.QueryEscape(blockList))
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
		jsonStr, err := json.Marshal(model.FilterCriteriaAndSet{{
			FunctionId:    functionId,
			DeviceClassId: deviceClassId,
			AspectId:      aspectId,
		}})
		if err != nil {
			t.Error(err)
			return
		}
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
		jsonStr, err := json.Marshal(model.FilterCriteriaAndSet{{
			FunctionId:    functionId,
			DeviceClassId: deviceClassId,
			AspectId:      aspectId,
		}})
		if err != nil {
			t.Error(err)
			return
		}
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

func testenv() (mux *sync.Mutex, semanticCalls *[]string, semanticmock *httptest.Server, searchmock *httptest.Server, devicerepomock *httptest.Server, selectionApi *httptest.Server, err error) {
	mux = &sync.Mutex{}
	calls := []string{}
	semanticCalls = &calls

	semanticmock = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.Lock()
		defer mux.Unlock()
		calls = append(calls, r.URL.Path+"?"+r.URL.RawQuery)
		json.NewEncoder(w).Encode([]devicemodel.DeviceType{
			{Id: "dt1", Name: "dt1name", DeviceClassId: "dc1", Services: []devicemodel.Service{
				testService("11", "pid", devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION, devicemodel.REQUEST),
				testService("11_b", "mqtt", devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION, devicemodel.EVENT),
				testService("12", "pid", devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION, devicemodel.REQUEST),
			}},
			{Id: "dt2", Name: "dt2name", DeviceClassId: "dc1", Services: []devicemodel.Service{
				testService("21", "pid", devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION, devicemodel.REQUEST),
				testService("22", "pid", devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION, devicemodel.REQUEST),
			}},
			{Id: "dt3", Name: "dt1name", DeviceClassId: "dc1", Services: []devicemodel.Service{
				testService("31", "mqtt", devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION, devicemodel.EVENT),
				testService("32", "mqtt", devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION, devicemodel.EVENT),
			}},
			{Id: "dt4", Name: "dt2name", DeviceClassId: "dc1", Services: []devicemodel.Service{
				testService("41", "mqtt", devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION, devicemodel.EVENT),
				testService("42", "mqtt", devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION, devicemodel.EVENT),
			}},
		})
	}))

	devicerepomock = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/device-types/dt1" {
			json.NewEncoder(w).Encode(devicemodel.DeviceType{Id: "dt1", Name: "dt1name", DeviceClassId: "dc1", Services: []devicemodel.Service{
				testTechnicalService("11", "pid", []devicemodel.Content{{
					Id: "content1",
					ContentVariable: devicemodel.ContentVariable{
						Id:   "variable1",
						Name: "variable1",
					},
				}}, devicemodel.REQUEST),
				testTechnicalService("11_b", "mqtt", []devicemodel.Content{{
					Id: "content2",
					ContentVariable: devicemodel.ContentVariable{
						Id:   "variable2",
						Name: "variable2",
					},
				}}, devicemodel.EVENT),
				testTechnicalService("12", "pid", []devicemodel.Content{{
					Id: "content3",
					ContentVariable: devicemodel.ContentVariable{
						Id:   "variable3",
						Name: "variable3",
					},
				}}, devicemodel.REQUEST),
			}})
			return
		}

		if r.URL.Path == "/device-types/dt2" {
			json.NewEncoder(w).Encode(devicemodel.DeviceType{Id: "dt2", Name: "dt2name", DeviceClassId: "dc1", Services: []devicemodel.Service{
				testTechnicalService("21", "pid", []devicemodel.Content{{
					Id: "content4",
					ContentVariable: devicemodel.ContentVariable{
						Id:   "variable4",
						Name: "variable4",
					},
				}}, devicemodel.REQUEST),
				testTechnicalService("22", "pid", []devicemodel.Content{{
					Id: "content5",
					ContentVariable: devicemodel.ContentVariable{
						Id:   "variable5",
						Name: "variable5",
					},
				}}, devicemodel.REQUEST),
			}})
			return
		}
	}))

	searchmock = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/jwt/select/devices/device_type_id/dt1/x" {
			json.NewEncoder(w).Encode([]TestPermSearchDevice{
				{Id: "1", Name: "1", DeviceType: "dt1", Permissions: model.Permissions{
					R: true,
					W: false,
					X: true,
					A: false,
				}},
			})
		}
		if r.URL.Path == "/jwt/select/devices/device_type_id/dt2/x" {
			json.NewEncoder(w).Encode([]TestPermSearchDevice{
				{Id: "2", Name: "2", DeviceType: "dt2", Permissions: model.Permissions{
					R: true,
					W: false,
					X: true,
					A: false,
				}},
			})
		}
		if r.URL.Path == "/jwt/select/devices/device_type_id/dt3/x" {
			json.NewEncoder(w).Encode([]TestPermSearchDevice{
				{Id: "3", Name: "3", DeviceType: "dt3", Permissions: model.Permissions{
					R: true,
					W: false,
					X: true,
					A: false,
				}},
			})
		}
		if r.URL.Path == "/jwt/select/devices/device_type_id/dt4/x" {
			json.NewEncoder(w).Encode([]TestPermSearchDevice{
				{Id: "4", Name: "4", DeviceType: "dt4", Permissions: model.Permissions{
					R: true,
					W: false,
					X: true,
					A: false,
				}},
			})
		}
	}))

	c := &configuration.ConfigStruct{
		SemanticRepoUrl: semanticmock.URL,
		PermSearchUrl:   searchmock.URL,
		DeviceRepoUrl:   devicerepomock.URL,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	repo, err := devices.New(ctx, c)
	if err != nil {
		searchmock.Close()
		selectionApi.Close()
		semanticmock.Close()
		devicerepomock.Close()
		return mux, semanticCalls, semanticmock, searchmock, devicerepomock, selectionApi, err
	}

	router := api.Router(c, repo)
	selectionApi = httptest.NewServer(router)

	return
}

func testService(id string, protocolId string, functionType string, interaction devicemodel.Interaction) devicemodel.Service {
	result := devicemodel.Service{
		Id:          id,
		LocalId:     id + "_l",
		Name:        id + "_name",
		AspectIds:   []string{"a1"},
		ProtocolId:  protocolId,
		Interaction: interaction,
	}
	if functionType == devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION {
		result.FunctionIds = []string{devicemodel.MEASURING_FUNCTION_PREFIX + "_1"}
	} else {
		result.FunctionIds = []string{devicemodel.CONTROLLING_FUNCTION_PREFIX + "_1"}
	}
	return result
}

func testTechnicalService(id string, protocolId string, outputs []devicemodel.Content, interaction devicemodel.Interaction) devicemodel.Service {
	return devicemodel.Service{
		Id:          id,
		LocalId:     id + "_l",
		Name:        id + "_name",
		ProtocolId:  protocolId,
		Outputs:     outputs,
		Interaction: interaction,
	}
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
		newElement := model.FilterCriteria{
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
