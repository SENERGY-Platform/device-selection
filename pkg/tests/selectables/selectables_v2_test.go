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
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/SENERGY-Platform/device-selection/pkg/model"
	"github.com/SENERGY-Platform/device-selection/pkg/model/devicemodel"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/environment/legacy"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/helper"
	"net/http"
	"net/url"
	"sync"
	"testing"
)

func TestApiSelectablesV2(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, _, _, selectionurl, err := legacy.Testenv(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	resultSimple := []model.Selectable{}
	t.Run("send simple request", sendSimpleRequestV2(selectionurl, &resultSimple, devicemodel.MEASURING_FUNCTION_PREFIX+"_1", "dc1", "a1", devicemodel.REQUEST))
	t.Run("check result", func(t *testing.T) {
		if len(resultSimple) != 1 {
			t.Error(len(resultSimple), resultSimple)
			return
		}
		if resultSimple[0].Device.Name != "1" ||
			resultSimple[0].Device.Id != "1" ||
			len(resultSimple[0].Services) != 1 ||
			resultSimple[0].Services[0].Id != "11" {
			temp, _ := json.Marshal(resultSimple[0])
			t.Error(string(temp))
			return
		}
	})

	resultJson := []model.Selectable{}
	t.Run("send json request", sendJsonRequestV2(selectionurl, &resultJson, devicemodel.MEASURING_FUNCTION_PREFIX+"_1", "dc1", "a1", devicemodel.REQUEST))
	t.Run("check json result", func(t *testing.T) {
		if len(resultJson) != 1 ||
			resultJson[0].Device.Name != "1" ||
			resultJson[0].Device.Id != "1" ||
			len(resultJson[0].Services) != 1 ||
			resultJson[0].Services[0].Id != "11" {
			t.Error(resultJson)
			return
		}
	})

	resultBase64 := []model.Selectable{}
	t.Run("send base64 request", sendBase64RequestV2(selectionurl, &resultBase64, devicemodel.MEASURING_FUNCTION_PREFIX+"_1", "dc1", "a1", devicemodel.REQUEST))
	t.Run("check base64 result", func(t *testing.T) {
		if len(resultBase64) != 1 ||
			resultBase64[0].Device.Name != "1" ||
			resultBase64[0].Device.Id != "1" ||
			len(resultBase64[0].Services) != 1 ||
			resultBase64[0].Services[0].Id != "11" {
			t.Error(resultBase64)
			return
		}
	})

	resultPost := []model.Selectable{}
	t.Run("send post request", sendPostRequestV2(selectionurl, &resultPost, devicemodel.MEASURING_FUNCTION_PREFIX+"_1", "dc1", "a1", devicemodel.REQUEST))
	t.Run("check post result", func(t *testing.T) {
		if len(resultPost) != 1 ||
			resultPost[0].Device.Name != "1" ||
			resultPost[0].Device.Id != "1" ||
			len(resultPost[0].Services) != 1 ||
			resultPost[0].Services[0].Id != "11" {
			t.Error(resultPost)
			return
		}
	})

	emptyInteractionResult := []model.Selectable{}
	t.Run("request empty interaction", sendPostRequestV2(selectionurl, &emptyInteractionResult, devicemodel.MEASURING_FUNCTION_PREFIX+"_1", "dc1", "a1", ""))
	t.Run("check post result", func(t *testing.T) {
		if len(resultPost) != 1 ||
			resultPost[0].Device.Name != "1" ||
			resultPost[0].Device.Id != "1" ||
			len(resultPost[0].Services) != 1 ||
			resultPost[0].Services[0].Id != "11" {
			t.Error(resultPost)
			return
		}
	})
}

func sendSimpleRequestV2(apiurl string, result interface{}, functionId string, deviceClassId string, aspectId string, interaction devicemodel.Interaction) func(t *testing.T) {
	return func(t *testing.T) {
		endpoint := apiurl + "/v2/selectables?include_devices=true&function_id=" + url.QueryEscape(functionId) + "&device_class_id=" + url.QueryEscape(deviceClassId) + "&aspect_id=" + url.QueryEscape(aspectId) + "&interaction=" + url.QueryEscape(string(interaction))
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", helper.AdminJwt)
		req.Header.Set("Content-Type", "application/json")
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

func sendJsonRequestV2(apiurl string, result interface{}, functionId string, deviceClassId string, aspectId string, interaction devicemodel.Interaction) func(t *testing.T) {
	return func(t *testing.T) {
		jsonStr, err := json.Marshal(model.FilterCriteriaAndSet{{
			FunctionId:    functionId,
			DeviceClassId: deviceClassId,
			AspectId:      aspectId,
			Interaction:   string(interaction),
		}})
		if err != nil {
			t.Error(err)
			return
		}
		endpoint := apiurl + "/v2/selectables?include_devices=true&json=" + url.QueryEscape(string(jsonStr))
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", helper.AdminJwt)
		req.Header.Set("Content-Type", "application/json")
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

func sendPostRequestV2(apiurl string, result interface{}, functionId string, deviceClassId string, aspectId string, interaction devicemodel.Interaction) func(t *testing.T) {
	return func(t *testing.T) {
		requestBody := new(bytes.Buffer)
		err := json.NewEncoder(requestBody).Encode(model.FilterCriteriaAndSet{{
			FunctionId:    functionId,
			DeviceClassId: deviceClassId,
			AspectId:      aspectId,
			Interaction:   string(interaction),
		}})
		if err != nil {
			t.Error(err)
			return
		}
		endpoint := apiurl + "/v2/query/selectables?include_devices=true"
		req, err := http.NewRequest("POST", endpoint, requestBody)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", helper.AdminJwt)
		req.Header.Set("Content-Type", "application/json")
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

func sendBase64RequestV2(apiurl string, result interface{}, functionId string, deviceClassId string, aspectId string, interaction devicemodel.Interaction) func(t *testing.T) {
	return func(t *testing.T) {
		jsonStr, err := json.Marshal(model.FilterCriteriaAndSet{{
			FunctionId:    functionId,
			DeviceClassId: deviceClassId,
			AspectId:      aspectId,
			Interaction:   string(interaction),
		}})
		if err != nil {
			t.Error(err)
			return
		}
		b64Str := base64.StdEncoding.EncodeToString(jsonStr)
		endpoint := apiurl + "/v2/selectables?include_devices=true&base64=" + b64Str
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", helper.AdminJwt)
		req.Header.Set("Content-Type", "application/json")
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
