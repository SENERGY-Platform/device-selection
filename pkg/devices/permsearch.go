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
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"runtime/debug"
)

func (this *Devices) getDevicesOfType(token string, deviceTypeId string) (result []model.PermSearchDevice, err error, code int) {
	return this.getCachedDevicesOfType(token, deviceTypeId, nil)
}

func (this *Devices) getCachedDevicesOfType(token string, deviceTypeId string, cache *map[string][]model.PermSearchDevice) (result []model.PermSearchDevice, err error, code int) {
	if cache != nil {
		if cacheResult, ok := (*cache)[deviceTypeId]; ok {
			return cacheResult, nil, http.StatusOK
		}
	}
	req, err := http.NewRequest("GET", this.config.PermSearchUrl+"/jwt/select/devices/device_type_id/"+url.PathEscape(deviceTypeId)+"/x", nil)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		debug.PrintStack()
		return result, errors.New(buf.String()), resp.StatusCode
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}

	if cache != nil {
		(*cache)[deviceTypeId] = result
	}

	return result, nil, http.StatusOK
}
