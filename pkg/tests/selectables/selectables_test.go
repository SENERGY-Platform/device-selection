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

func TestApiSimpleGet(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, _, _, selectionurl, err := legacy.Testenv(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	result := []model.Selectable{}

	t.Run("send simple request", sendSimpleRequest(selectionurl, &result, devicemodel.MEASURING_FUNCTION_PREFIX+"_1", "dc1", "a1", "mqtt"))

	t.Run("check result", func(t *testing.T) {
		if len(result) != 1 {
			t.Error(len(result), result)
			return
		}
		if result[0].Device.Name != "1" ||
			result[0].Device.Id != "1" ||
			len(result[0].Services) != 1 ||
			result[0].Services[0].Id != "11" {
			temp, _ := json.Marshal(result[0])
			t.Error(string(temp))
			return
		}
	})
}

func TestApiCompleteSimpledGet(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, _, _, selectionurl, err := legacy.Testenv(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	result := []model.Selectable{}

	t.Run("send simple request", sendCompletedSimpleRequest(selectionurl, &result, devicemodel.MEASURING_FUNCTION_PREFIX+"_1", "dc1", "a1", "mqtt"))

	t.Run("check result", func(t *testing.T) {
		if len(result) != 1 ||
			result[0].Device.Name != "1" ||
			result[0].Device.Id != "1" ||
			len(result[0].Services) != 1 ||
			result[0].Services[0].Id != "11" ||
			len(result[0].Services[0].Outputs) != 1 ||
			result[0].Services[0].Outputs[0].Id != "content1" {
			t.Error(result)
			return
		}
	})
}

func TestApiJsonGet(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, _, _, selectionurl, err := legacy.Testenv(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	result := []model.Selectable{}

	t.Run("send json request", sendJsonRequest(selectionurl, &result, devicemodel.MEASURING_FUNCTION_PREFIX+"_1", "dc1", "a1", "mqtt"))

	t.Run("check result", func(t *testing.T) {
		if len(result) != 1 ||
			result[0].Device.Name != "1" ||
			result[0].Device.Id != "1" ||
			len(result[0].Services) != 1 ||
			result[0].Services[0].Id != "11" {
			t.Error(result)
			return
		}
	})

}

func TestApiBase64Get(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, _, _, selectionurl, err := legacy.Testenv(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	result := []model.Selectable{}

	t.Run("send base64 request", sendBase64Request(selectionurl, &result, devicemodel.MEASURING_FUNCTION_PREFIX+"_1", "dc1", "a1", "mqtt"))

	t.Run("check result", func(t *testing.T) {
		if len(result) != 1 ||
			result[0].Device.Name != "1" ||
			result[0].Device.Id != "1" ||
			len(result[0].Services) != 1 ||
			result[0].Services[0].Id != "11" {
			t.Error(result)
			return
		}
	})
}

func sendSimpleRequest(apiurl string, result interface{}, functionId string, deviceClassId string, aspectId string, blockList string) func(t *testing.T) {
	return func(t *testing.T) {
		endpoint := apiurl + "/selectables?function_id=" + url.QueryEscape(functionId) + "&device_class_id=" + url.QueryEscape(deviceClassId) + "&aspect_id=" + url.QueryEscape(aspectId) + "&filter_protocols=" + url.QueryEscape(blockList)
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

func sendCompletedSimpleRequest(apiurl string, result interface{}, functionId string, deviceClassId string, aspectId string, blockList string) func(t *testing.T) {
	return func(t *testing.T) {
		endpoint := apiurl + "/selectables?complete_services=true&function_id=" + url.QueryEscape(functionId) + "&device_class_id=" + url.QueryEscape(deviceClassId) + "&aspect_id=" + url.QueryEscape(aspectId) + "&filter_protocols=" + url.QueryEscape(blockList)
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
		endpoint := apiurl + "/selectables?json=" + url.QueryEscape(string(jsonStr)) + "&filter_protocols=" + url.QueryEscape(blockList)
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
		endpoint := apiurl + "/selectables?base64=" + b64Str + "&filter_protocols=" + url.QueryEscape(blockList)
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
