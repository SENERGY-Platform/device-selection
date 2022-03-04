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
	"device-selection/pkg/model"
	"device-selection/pkg/model/devicemodel"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

func (this *Controller) GetDeviceTypeSelectablesCached(token string, descriptions model.FilterCriteriaAndSet) (result []devicemodel.DeviceTypeSelectable, err error) {
	hash := hashCriteriaAndSet(descriptions)
	err = this.cache.Use("device-type-selectables."+hash, func() (interface{}, error) {
		return this.GetDeviceTypeSelectables(token, descriptions)
	}, &result)
	return
}

func (this *Controller) GetDeviceTypeSelectables(token string, descriptions model.FilterCriteriaAndSet) (result []devicemodel.DeviceTypeSelectable, err error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	payload := new(bytes.Buffer)
	err = json.NewEncoder(payload).Encode(descriptions)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	req, err := http.NewRequest(
		"POST",
		this.config.DeviceRepoUrl+"/query/device-type-selectables",
		payload,
	)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	req.Header.Set("Authorization", token)

	resp, err := client.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		debug.PrintStack()
		temp, _ := io.ReadAll(resp.Body)
		log.Println("ERROR: GetDeviceTypeSelectables():", resp.StatusCode, string(temp))
		return result, errors.New("unexpected statuscode")
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		debug.PrintStack()
		return result, err
	}

	temp, _ := json.Marshal(result)
	log.Println("TEST:", string(temp))

	return result, err
}
