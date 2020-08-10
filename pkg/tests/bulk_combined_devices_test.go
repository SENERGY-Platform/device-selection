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

func TestApiBulkCombinedDevices(t *testing.T) {
	mux, calls, semanticmock, searchmock, devicerepomock, selectionApi, err := testenv()
	if err != nil {
		t.Error(err)
		return
	}
	defer semanticmock.Close()
	defer selectionApi.Close()
	defer searchmock.Close()
	defer devicerepomock.Close()

	result := []model.PermSearchDevice{}

	eventInteraction := devicemodel.EVENT

	request := model.BulkRequest{
		{
			Id:              "1",
			FilterProtocols: []string{"mqtt"},
			Criteria: []model.FilterCriteria{{
				FunctionId:    devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION + "_1",
				DeviceClassId: "dc1",
				AspectId:      "a1",
			}},
		},
		{
			Id:                "2",
			FilterInteraction: &eventInteraction,
			Criteria: []model.FilterCriteria{{
				FunctionId:    devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION + "_1",
				DeviceClassId: "dc1",
				AspectId:      "a1",
			}},
		},
	}

	temp, _ := json.Marshal(request)
	t.Log("request:", string(temp))

	t.Run("send request", sendBulkCombinedDevicesRequest(selectionApi.URL, &result, request))

	temp, _ = json.Marshal(result)
	t.Log("response:", string(temp))

	t.Run("check bulk result", func(t *testing.T) {
		if len(result) != 1 {
			t.Error(result)
			return
		}
		if result[0].Name != "1" ||
			result[0].Id != "1" ||
			!result[0].Permissions.R ||
			result[0].Permissions.W ||
			!result[0].Permissions.X ||
			result[0].Permissions.A {
			t.Error(result[0])
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

func sendBulkCombinedDevicesRequest(apiurl string, result interface{}, request model.BulkRequest) func(t *testing.T) {
	return func(t *testing.T) {
		buff := new(bytes.Buffer)
		err := json.NewEncoder(buff).Encode(request)
		if err != nil {
			t.Error(err)
			return
		}
		req, err := http.NewRequest("POST", apiurl+"/bulk/selectables/combined/devices", buff)
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
