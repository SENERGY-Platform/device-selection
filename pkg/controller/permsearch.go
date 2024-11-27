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
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-selection/pkg/controller/idmodifier"
	"github.com/SENERGY-Platform/device-selection/pkg/model"
	"log"
	"net/http"
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

	devices, _, err, code := this.devicerepo.ListExtendedDevices(token, client.ExtendedDeviceListOptions{
		DeviceTypeIds: []string{pureId},
		Limit:         9999,
		Offset:        0,
		SortBy:        "name.asc",
		Permission:    client.READ,
	})
	if err != nil {
		return result, err, code
	}

	if pureId != deviceTypeId {
		ids := []string{}
		for _, device := range devices {
			ids = append(ids, idmodifier.JoinModifier(device.Id, modifier))
		}
		devices, _, err, code = this.devicerepo.ListExtendedDevices(token, client.ExtendedDeviceListOptions{
			Ids:        ids,
			SortBy:     "name.asc",
			Permission: client.READ,
		})
		if err != nil {
			return result, err, code
		}
	}

	for _, device := range devices {
		result = append(result, model.PermSearchDevice{
			Device:      device.Device,
			DisplayName: device.DisplayName,
			Permissions: model.Permissions{
				R: device.Permissions.Read,
				W: device.Permissions.Write,
				X: device.Permissions.Execute,
				A: device.Permissions.Administrate,
			},
			Shared:  device.Shared,
			Creator: device.OwnerId,
		})
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

func (this *Controller) getCachedDevicesOfTypeFilteredByLocalIdList(token string, deviceTypeId string, cache *map[string][]model.PermSearchDevice, localDeviceIds []string) (result []model.PermSearchDevice, err error, code int) {
	if cache != nil {
		if cacheResult, ok := (*cache)[deviceTypeId]; ok {
			return cacheResult, nil, http.StatusOK
		}
	}
	pureId, modifier := idmodifier.SplitModifier(deviceTypeId)

	devices, _, err, code := this.devicerepo.ListExtendedDevices(token, client.ExtendedDeviceListOptions{
		DeviceTypeIds: []string{pureId},
		LocalIds:      localDeviceIds,
		Limit:         9999,
		Offset:        0,
		SortBy:        "name.asc",
		Permission:    client.READ,
	})
	if err != nil {
		return result, err, code
	}

	if pureId != deviceTypeId {
		ids := []string{}
		for _, device := range devices {
			ids = append(ids, idmodifier.JoinModifier(device.Id, modifier))
		}
		devices, _, err, code = this.devicerepo.ListExtendedDevices(token, client.ExtendedDeviceListOptions{
			Ids:        ids,
			SortBy:     "name.asc",
			Permission: client.READ,
		})
		if err != nil {
			return result, err, code
		}
	}

	for _, device := range devices {
		result = append(result, model.PermSearchDevice{
			Device:      device.Device,
			DisplayName: device.DisplayName,
			Permissions: model.Permissions{
				R: device.Permissions.Read,
				W: device.Permissions.Write,
				X: device.Permissions.Execute,
				A: device.Permissions.Administrate,
			},
			Shared:  device.Shared,
			Creator: device.OwnerId,
		})
	}
	if cache != nil {
		(*cache)[deviceTypeId] = result
	}
	return result, nil, http.StatusOK
}
