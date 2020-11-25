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
	"device-selection/pkg/model"
	"device-selection/pkg/model/devicemodel"
	"encoding/json"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestApiBulkSelectables(t *testing.T) {
	mux, calls, semanticmock, searchmock, devicerepomock, selectionApi, err := testenv()
	if err != nil {
		t.Error(err)
		return
	}
	defer semanticmock.Close()
	defer selectionApi.Close()
	defer searchmock.Close()
	defer devicerepomock.Close()

	result := model.BulkResult{}

	eventInteraction := devicemodel.EVENT

	request := model.BulkRequest{
		{
			Id:              "1",
			FilterProtocols: []string{"mqtt"},
			Criteria: []model.FilterCriteria{{
				FunctionId:    devicemodel.MEASURING_FUNCTION_PREFIX + "_1",
				DeviceClassId: "dc1",
				AspectId:      "a1",
			}},
		},
		{
			Id:                "2",
			FilterInteraction: &eventInteraction,
			Criteria: []model.FilterCriteria{{
				FunctionId:    devicemodel.MEASURING_FUNCTION_PREFIX + "_1",
				DeviceClassId: "dc1",
				AspectId:      "a1",
			}},
		},
		{
			Id:              "3",
			FilterProtocols: []string{"mqtt", "pid"},
			Criteria: []model.FilterCriteria{{
				FunctionId:    devicemodel.MEASURING_FUNCTION_PREFIX + "_1",
				DeviceClassId: "unknown",
				AspectId:      "a1",
			}},
		},
	}

	temp, _ := json.Marshal(request)
	t.Log("request:", string(temp))

	t.Run("send bulk request", sendBulkRequest(selectionApi.URL, &result, request))

	temp, _ = json.Marshal(result)
	t.Log("response:", string(temp))

	t.Run("check bulk result", func(t *testing.T) {
		if len(result) != 3 {
			t.Error(result)
			return
		}
		if result[0].Id != "1" {
			t.Error(result[0])
			return
		}
		if result[1].Id != "2" {
			t.Error(result[1])
			return
		}
		if result[2].Id != "3" {
			t.Error(result[2])
			return
		}
		if len(result[0].Selectables) != 1 {
			t.Error(result[0].Selectables)
			return
		}
		if result[0].Selectables[0].Device.Name != "1" ||
			result[0].Selectables[0].Device.Id != "1" ||
			len(result[0].Selectables[0].Services) != 1 ||
			result[0].Selectables[0].Services[0].Id != "11" ||
			!result[0].Selectables[0].Device.Permissions.R ||
			result[0].Selectables[0].Device.Permissions.W ||
			!result[0].Selectables[0].Device.Permissions.X ||
			result[0].Selectables[0].Device.Permissions.A {
			t.Error(result[0].Selectables[0])
			return
		}
		if len(result[1].Selectables) != 1 {
			t.Error(result[1].Selectables)
			return
		}
		if !reflect.DeepEqual(result[0].Selectables[0], result[1].Selectables[0]) {
			t.Error(result[1].Selectables[0])
		}
		if len(result[2].Selectables) != 0 {
			t.Error(result[2].Selectables)
			return
		}
	})

	t.Run("check semantic calls", func(t *testing.T) {
		mux.Lock()
		defer mux.Unlock()
		expected := []string{
			"/device-types?filter=" + url.QueryEscape(`[{"function_id":"`+devicemodel.MEASURING_FUNCTION_PREFIX+`_1","aspect_id":"a1","device_class_id":"dc1"}]`),
			"/device-types?filter=" + url.QueryEscape(`[{"function_id":"`+devicemodel.MEASURING_FUNCTION_PREFIX+`_1","aspect_id":"a1","device_class_id":"unknown"}]`),
		}
		if !reflect.DeepEqual(*calls, expected) {
			actualStr, _ := json.Marshal(calls)
			expectedStr, _ := json.Marshal(expected)
			t.Error(string(actualStr), string(expectedStr))
		}
	})
}

func TestApiCompletedBulkSelectables(t *testing.T) {
	mux, calls, semanticmock, searchmock, devicerepomock, selectionApi, err := testenv()
	if err != nil {
		t.Error(err)
		return
	}
	defer semanticmock.Close()
	defer selectionApi.Close()
	defer searchmock.Close()
	defer devicerepomock.Close()

	result := model.BulkResult{}

	eventInteraction := devicemodel.EVENT

	request := model.BulkRequest{
		{
			Id:              "1",
			FilterProtocols: []string{"mqtt"},
			Criteria: []model.FilterCriteria{{
				FunctionId:    devicemodel.MEASURING_FUNCTION_PREFIX + "_1",
				DeviceClassId: "dc1",
				AspectId:      "a1",
			}},
		},
		{
			Id:                "2",
			FilterInteraction: &eventInteraction,
			Criteria: []model.FilterCriteria{{
				FunctionId:    devicemodel.MEASURING_FUNCTION_PREFIX + "_1",
				DeviceClassId: "dc1",
				AspectId:      "a1",
			}},
		},
		{
			Id:              "3",
			FilterProtocols: []string{"mqtt", "pid"},
			Criteria: []model.FilterCriteria{{
				FunctionId:    devicemodel.MEASURING_FUNCTION_PREFIX + "_1",
				DeviceClassId: "unknown",
				AspectId:      "a1",
			}},
		},
	}

	temp, _ := json.Marshal(request)
	t.Log("request:", string(temp))

	t.Run("send bulk request", sendCompletedBulkRequest(selectionApi.URL, &result, request))

	temp, _ = json.Marshal(result)
	t.Log("response:", string(temp))

	t.Run("check bulk result", func(t *testing.T) {
		if len(result) != 3 {
			t.Error(result)
			return
		}
		if result[0].Id != "1" {
			t.Error(result[0])
			return
		}
		if result[1].Id != "2" {
			t.Error(result[1])
			return
		}
		if result[2].Id != "3" {
			t.Error(result[2])
			return
		}
		if len(result[0].Selectables) != 1 {
			t.Error(result[0].Selectables)
			return
		}
		if result[0].Selectables[0].Device.Name != "1" ||
			result[0].Selectables[0].Device.Id != "1" ||
			len(result[0].Selectables[0].Services) != 1 ||
			result[0].Selectables[0].Services[0].Id != "11" ||
			len(result[0].Selectables[0].Services[0].Outputs) != 1 ||
			result[0].Selectables[0].Services[0].Outputs[0].Id != "content1" ||
			!result[0].Selectables[0].Device.Permissions.R ||
			result[0].Selectables[0].Device.Permissions.W ||
			!result[0].Selectables[0].Device.Permissions.X ||
			result[0].Selectables[0].Device.Permissions.A {
			t.Error(result[0].Selectables[0])
			return
		}
		if len(result[1].Selectables) != 1 {
			t.Error(result[1].Selectables)
			return
		}
		if !reflect.DeepEqual(result[0].Selectables[0], result[1].Selectables[0]) {
			t.Error(result[1].Selectables[0])
		}
		if len(result[2].Selectables) != 0 {
			t.Error(result[2].Selectables)
			return
		}
	})

	t.Run("check semantic calls", func(t *testing.T) {
		mux.Lock()
		defer mux.Unlock()
		expected := []string{
			"/device-types?filter=" + url.QueryEscape(`[{"function_id":"`+devicemodel.MEASURING_FUNCTION_PREFIX+`_1","aspect_id":"a1","device_class_id":"dc1"}]`),
			"/device-types?filter=" + url.QueryEscape(`[{"function_id":"`+devicemodel.MEASURING_FUNCTION_PREFIX+`_1","aspect_id":"a1","device_class_id":"unknown"}]`),
		}
		if !reflect.DeepEqual(*calls, expected) {
			actualStr, _ := json.Marshal(calls)
			expectedStr, _ := json.Marshal(expected)
			t.Error(string(actualStr), string(expectedStr))
		}
	})
}

func sendBulkRequest(apiurl string, result interface{}, request model.BulkRequest) func(t *testing.T) {
	return func(t *testing.T) {
		buff := new(bytes.Buffer)
		err := json.NewEncoder(buff).Encode(request)
		if err != nil {
			t.Error(err)
			return
		}
		req, err := http.NewRequest("POST", apiurl+"/bulk/selectables", buff)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", "test-token")
		resp, err := http.DefaultClient.Do(req)
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

func sendCompletedBulkRequest(apiurl string, result interface{}, request model.BulkRequest) func(t *testing.T) {
	return func(t *testing.T) {
		buff := new(bytes.Buffer)
		err := json.NewEncoder(buff).Encode(request)
		if err != nil {
			t.Error(err)
			return
		}
		req, err := http.NewRequest("POST", apiurl+"/bulk/selectables?complete_services=true", buff)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", "test-token")
		resp, err := http.DefaultClient.Do(req)
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
