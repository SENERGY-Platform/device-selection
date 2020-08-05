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
	"device-selection/pkg/devicemodel"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"runtime/debug"
)

func (this *Devices) GetDevicesOfType(token string, deviceTypeId string) (result []devicemodel.Device, err error, code int) {
	devices, err, code := this.getDevicesOfTypeFromPermsearch(token, deviceTypeId)
	if err != nil {
		return result, err, code
	}
	for _, device := range devices {
		result = append(result, devicemodel.Device{
			Id:           device.Id,
			LocalId:      device.LocalId,
			Name:         device.Name,
			DeviceTypeId: device.DeviceType,
		})
	}
	return result, nil, 200
}

type PermSearchDevice struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	DeviceType string `json:"device_type_id"`
	LocalId    string `json:"local_id"`
}

func (this *Devices) getDevicesOfTypeFromPermsearch(token string, deviceTypeId string) (result []PermSearchDevice, err error, code int) {
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
	return result, nil, http.StatusOK
}
