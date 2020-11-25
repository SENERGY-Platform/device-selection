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
	"device-selection/pkg/model"
	"device-selection/pkg/model/devicemodel"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"sort"
	"time"
)

func (this *Controller) getCachedTechnicalDeviceType(token string, id string, cache *map[string]devicemodel.DeviceType) (result devicemodel.DeviceType, err error) {
	if cache != nil {
		if cacheResult, ok := (*cache)[id]; ok {
			return cacheResult, nil
		}
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest(
		"GET",
		this.config.DeviceRepoUrl+"/device-types/"+url.QueryEscape(id),
		nil,
	)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	req.Header.Set("Authorization", string(token))

	resp, err := client.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		debug.PrintStack()
		return result, errors.New("unexpected statuscode")
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		debug.PrintStack()
		return result, err
	}

	if cache != nil {
		(*cache)[id] = result
	}

	return result, err
}

func (this *Controller) getFilteredDeviceTypes(token string, descriptions model.FilterCriteriaAndSet) (result []devicemodel.DeviceType, err error, code int) {
	return this.getCachedFilteredDeviceTypes(token, descriptions, nil)
}

func (this *Controller) getCachedFilteredDeviceTypes(token string, descriptions model.FilterCriteriaAndSet, cache *map[string][]devicemodel.DeviceType) (result []devicemodel.DeviceType, err error, code int) {
	hash := hashCriteriaAndSet(descriptions)
	if cache != nil {
		if cacheResult, ok := (*cache)[hash]; ok {
			return cacheResult, nil, http.StatusOK
		}
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}
	payload, err := json.Marshal(descriptions)

	req, err := http.NewRequest(
		"GET",
		this.config.SemanticRepoUrl+"/device-types?filter="+url.QueryEscape(string(payload)),
		nil,
	)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", string(token))

	resp, err := client.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		debug.PrintStack()
		return result, errors.New("unexpected statuscode"), resp.StatusCode
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}

	if cache != nil {
		(*cache)[hash] = result
	}

	return result, err, resp.StatusCode
}

func hashCriteriaAndSet(criteria model.FilterCriteriaAndSet) string {
	arr := append([]model.FilterCriteria{}, criteria...) //make copy to prevent sorting to effect original
	sort.SliceStable(arr, func(i, j int) bool {
		return fmt.Sprint(arr[i]) < fmt.Sprint(arr[j])
	})
	return fmt.Sprint(arr)
}

func (this *Controller) getCachedTechnicalDevice(token string, id string, cache *map[string]devicemodel.Device) (result devicemodel.Device, err error, code int) {
	if cache != nil {
		if cacheResult, ok := (*cache)[id]; ok {
			return cacheResult, nil, http.StatusOK
		}
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest(
		"GET",
		this.config.DeviceRepoUrl+"/devices/"+url.QueryEscape(id),
		nil,
	)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)

	resp, err := client.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		debug.PrintStack()
		return result, errors.New("unable to load device: " + resp.Status), resp.StatusCode
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}

	if cache != nil {
		(*cache)[id] = result
	}

	return result, nil, http.StatusOK
}