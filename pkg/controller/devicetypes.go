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
	"fmt"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-selection/pkg/model"
	"github.com/SENERGY-Platform/device-selection/pkg/model/devicemodel"
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
	"runtime/debug"
	"sort"
)

func (this *Controller) getCachedDeviceType(token string, id string, cache *map[string]devicemodel.DeviceType) (result devicemodel.DeviceType, err error) {
	if cache != nil {
		if cacheResult, ok := (*cache)[id]; ok {
			return cacheResult, nil
		}
	}
	result, err, _ = this.devicerepo.ReadDeviceType(id, token)
	if err != nil {
		debug.PrintStack()
		return result, err
	}

	if cache != nil {
		(*cache)[id] = result
	}

	return result, err
}

func (this *Controller) GetFilteredDeviceTypes(token string, criteria []client.FilterCriteria) (result []models.DeviceType, err error, code int) {
	return this.getCachedFilteredDeviceTypes(token, criteria, nil)
}

func (this *Controller) getCachedFilteredDeviceTypes(token string, criteria []client.FilterCriteria, cache *map[string][]models.DeviceType) (result []models.DeviceType, err error, code int) {
	hash := hashClientCriteriaList(criteria)
	if cache != nil {
		if cacheResult, ok := (*cache)[hash]; ok {
			return cacheResult, nil, http.StatusOK
		}
	}

	query := client.DeviceTypeListOptions{
		Ids:             nil,
		Search:          "",
		Limit:           10000,
		Offset:          0,
		SortBy:          "name.asc",
		Criteria:        criteria,
		IncludeModified: true,
	}

	result, err, code = this.devicerepo.ListDeviceTypesV3(token, query)
	if err != nil {
		debug.PrintStack()
		return result, err, code
	}

	if cache != nil {
		(*cache)[hash] = result
	}

	return result, err, code
}

func (this *Controller) getOnlyDeviceTypesIncludingIdModifier(token string) (result []devicemodel.DeviceType, err error, code int) {
	return this.devicerepo.ListDeviceTypesV3(token, client.DeviceTypeListOptions{
		Limit:            9999,
		Offset:           0,
		SortBy:           "name.asc",
		IncludeModified:  true,
		IgnoreUnmodified: true,
	})
}

func hashCriteriaAndSet(criteria model.FilterCriteriaAndSet) string {
	arr := append(model.FilterCriteriaAndSet{}, criteria...) //make copy to prevent sorting to effect original
	sort.SliceStable(arr, func(i, j int) bool {
		return fmt.Sprint(arr[i]) < fmt.Sprint(arr[j])
	})
	return fmt.Sprint(arr)
}

func hashClientCriteriaList(criteria []client.FilterCriteria) string {
	arr := []client.FilterCriteria{}
	arr = append(arr, criteria...) //make copy to prevent sorting to effect original
	sort.SliceStable(arr, func(i, j int) bool {
		return fmt.Sprint(arr[i]) < fmt.Sprint(arr[j])
	})
	return fmt.Sprint(arr)
}

func (this *Controller) getCachedDevice(token string, id string, cache *map[string]devicemodel.Device) (result devicemodel.Device, err error, code int) {
	if cache != nil {
		if cacheResult, ok := (*cache)[id]; ok {
			return cacheResult, nil, http.StatusOK
		}
	}

	result, err, code = this.devicerepo.ReadDevice(id, token, client.READ)
	if err != nil {
		debug.PrintStack()
		return result, err, code
	}

	if cache != nil {
		(*cache)[id] = result
	}

	return result, nil, http.StatusOK
}
