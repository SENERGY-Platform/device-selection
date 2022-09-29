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
	"bytes"
	"device-selection/pkg/controller/idmodifier"
	"device-selection/pkg/model"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"runtime/debug"
)

func (this *Controller) getDevicesOfType(token string, deviceTypeId string) (result []model.PermSearchDevice, err error, code int) {
	return this.getCachedDevicesOfType(token, deviceTypeId, nil)
}

// limited to 1000 devices
func (this *Controller) getCachedDevicesOfType(token string, deviceTypeId string, cache *map[string][]model.PermSearchDevice) (result []model.PermSearchDevice, err error, code int) {
	if cache != nil {
		if cacheResult, ok := (*cache)[deviceTypeId]; ok {
			return cacheResult, nil, http.StatusOK
		}
	}

	pureId, modifier := idmodifier.SplitModifier(deviceTypeId)
	query := model.QueryMessage{
		Resource: "devices",
		Find: &model.QueryFind{
			Filter: &model.Selection{
				Condition: model.ConditionConfig{
					Feature:   "features.device_type_id",
					Operation: model.QueryEqualOperation,
					Value:     pureId,
				},
			},
		},
	}

	if pureId != deviceTypeId {
		query.Find.AddIdModifier = modifier
	}

	err, code = this.Search(token, query, &result)
	if err != nil {
		debug.PrintStack()
		return result, err, code
	}
	if cache != nil {
		(*cache)[deviceTypeId] = result
	}

	if this.config.Debug {
		jsonResult, _ := json.Marshal(result)
		log.Println("DEBUG: getCachedDevicesOfType(", deviceTypeId, ") = \n\t", string(jsonResult))
	}

	return result, nil, http.StatusOK
}

func (this *Controller) Search(token string, query model.QueryMessage, result interface{}) (err error, code int) {
	requestBody := new(bytes.Buffer)
	err = json.NewEncoder(requestBody).Encode(query)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req, err := http.NewRequest("POST", this.config.PermSearchUrl+"/v3/query", requestBody)
	if err != nil {
		debug.PrintStack()
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.PrintStack()
		return err, http.StatusInternalServerError
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		err = errors.New(buf.String())
		log.Println("ERROR: ", resp.StatusCode, err)
		debug.PrintStack()
		return err, resp.StatusCode
	}
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		debug.PrintStack()
		return err, http.StatusInternalServerError
	}

	return nil, http.StatusOK
}

func (this *Controller) getCachedDevicesOfTypeFilteredByLocalIdList(token string, deviceTypeId string, cache *map[string][]model.PermSearchDevice, localDeviceIds []string) (result []model.PermSearchDevice, err error, code int) {
	if cache != nil {
		if cacheResult, ok := (*cache)[deviceTypeId]; ok {
			return cacheResult, nil, http.StatusOK
		}
	}
	pureId, modifier := idmodifier.SplitModifier(deviceTypeId)

	query := model.QueryMessage{
		Resource: "devices",
		Find: &model.QueryFind{
			QueryListCommons: model.QueryListCommons{
				Limit:    1000,
				Offset:   0,
				Rights:   "rx",
				SortBy:   "name",
				SortDesc: false,
			},
			Search: "",
			Filter: &model.Selection{
				And: []model.Selection{
					{
						Condition: model.ConditionConfig{
							Feature:   "features.device_type_id",
							Operation: model.QueryEqualOperation,
							Value:     pureId,
						},
					},
					{
						Condition: model.ConditionConfig{
							Feature:   "features.local_id",
							Operation: model.QueryAnyValueInFeatureOperation,
							Value:     localDeviceIds,
						},
					},
				},
			},
		},
	}
	if pureId != deviceTypeId {
		query.Find.AddIdModifier = modifier
	}
	err, code = this.Search(token, query, &result)
	if cache != nil {
		(*cache)[deviceTypeId] = result
	}
	return
}
