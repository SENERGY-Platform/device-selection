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
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-selection/pkg/controller/idmodifier"
	"github.com/SENERGY-Platform/device-selection/pkg/model"
	"github.com/SENERGY-Platform/device-selection/pkg/model/devicemodel"
	"github.com/SENERGY-Platform/models/go/models"
	"log"
	"net/http"
	"slices"
	"sort"
	"strings"
)

func (this *Controller) DeviceGroupHelper(token string, deviceIds []string, search model.DeviceGroupHelperPagination, maintainGroupUsability bool, functionBlockList []string) (result model.DeviceGroupHelperResult, err error, code int) {
	deviceCache := &map[string]devicemodel.Device{}
	deviceTypeCache := &map[string]devicemodel.DeviceType{}
	result.Criteria, err, code = this.GetDeviceGroupCriteria(token, deviceTypeCache, deviceCache, deviceIds)
	if err != nil {
		return
	}
	result.Options, err, code = this.getDeviceGroupOptions(token, deviceTypeCache, deviceCache, deviceIds, result.Criteria, search, maintainGroupUsability, functionBlockList)
	return result, err, code
}

func (this *Controller) GetDeviceGroupCriteria(token string, deviceTypeCache *map[string]devicemodel.DeviceType, deviceCache *map[string]devicemodel.Device, deviceIds []string) (result []devicemodel.DeviceGroupFilterCriteria, err error, code int) {
	currentSet := map[string]devicemodel.DeviceGroupFilterCriteria{}
	for i, deviceId := range deviceIds {
		deviceCriterias, err, code := this.getDeviceCriteria(token, deviceTypeCache, deviceCache, deviceId)
		if err != nil {
			return result, err, code
		}
		nextSet := map[string]devicemodel.DeviceGroupFilterCriteria{}
		for _, criteria := range deviceCriterias {
			criteriaHash := criteriaHash(criteria)
			_, usedInCurrent := currentSet[criteriaHash]
			if i == 0 || usedInCurrent {
				nextSet[criteriaHash] = criteria
			}
		}
		currentSet = nextSet
	}
	result = []devicemodel.DeviceGroupFilterCriteria{}
	for _, element := range currentSet {
		result = append(result, element)
	}
	return result, nil, http.StatusOK
}

func (this *Controller) getDeviceCriteria(token string, deviceTypeCache *map[string]devicemodel.DeviceType, deviceCache *map[string]devicemodel.Device, deviceId string) (result []devicemodel.DeviceGroupFilterCriteria, err error, code int) {
	device, err, code := this.getCachedDevice(token, deviceId, deviceCache)
	if err != nil {
		return result, err, code
	}
	deviceType, err := this.getCachedDeviceType(token, device.DeviceTypeId, deviceTypeCache)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	resultSet := map[string]devicemodel.DeviceGroupFilterCriteria{}
	for _, service := range deviceType.Services {
		interactions := []devicemodel.Interaction{service.Interaction}
		if service.Interaction == devicemodel.EVENT_AND_REQUEST {
			interactions = []devicemodel.Interaction{devicemodel.EVENT, devicemodel.REQUEST}
		}
		work := []devicemodel.ContentVariable{}
		for _, content := range service.Inputs {
			work = append(work, content.ContentVariable)
		}
		for _, content := range service.Outputs {
			work = append(work, content.ContentVariable)
		}
		for i := 0; i < len(work); i++ {
			current := work[i]
			if current.FunctionId != "" {
				for _, interaction := range interactions {
					if isMeasuringFunctionId(current.FunctionId) {
						criteria := devicemodel.DeviceGroupFilterCriteria{
							FunctionId:  current.FunctionId,
							AspectId:    current.AspectId,
							Interaction: interaction,
						}
						resultSet[criteriaHash(criteria)] = criteria
						if current.AspectId != "" {
							aspectNode, err := this.GetAspectNode(current.AspectId, token)
							if err != nil {
								return result, err, http.StatusInternalServerError
							}
							for _, aspect := range aspectNode.AncestorIds {
								criteria := devicemodel.DeviceGroupFilterCriteria{
									FunctionId:  current.FunctionId,
									AspectId:    aspect,
									Interaction: interaction,
								}
								resultSet[criteriaHash(criteria)] = criteria
							}
						}
					} else {
						criteria := devicemodel.DeviceGroupFilterCriteria{
							FunctionId:    current.FunctionId,
							DeviceClassId: deviceType.DeviceClassId,
							Interaction:   interaction,
						}
						resultSet[criteriaHash(criteria)] = criteria
					}
				}
			}
			if len(current.SubContentVariables) > 0 {
				work = append(work, current.SubContentVariables...)
			}
		}
	}
	for _, element := range resultSet {
		result = append(result, element)
	}
	return result, nil, http.StatusOK
}

func criteriaHash(criteria devicemodel.DeviceGroupFilterCriteria) string {
	return criteria.FunctionId + "_" + criteria.AspectId + "_" + criteria.DeviceClassId + "_" + string(criteria.Interaction)
}

func (this *Controller) getDeviceGroupOptionsGetDevice(
	token string,
	currentDeviceIds []string,
	criteria []devicemodel.DeviceGroupFilterCriteria,
	search model.DeviceGroupHelperPagination,
	maintainGroupUsability bool,
	functionBlockList []string) (devices []model.PermSearchDevice, err error, code int) {

	validDeviceTypes := []string{}
	if maintainGroupUsability && len(criteria) > 0 {
		validDeviceTypes, err = this.getValidDeviceTypesForDeviceGroup(token, criteria, functionBlockList)
		if err != nil {
			log.Println("ERROR: getValidDeviceTypesForDeviceGroup()", err)
			err = nil
		}
	}

	unmodifiedDevices, err, code := this.getDeviceGroupOptionsGetDevicesUnmodified(token, currentDeviceIds, search, validDeviceTypes)
	if err != nil {
		return devices, err, code
	}
	devices = append(devices, unmodifiedDevices...)

	modifiedDevices, err, code := this.getDeviceGroupOptionsGetDevicesModified(token, currentDeviceIds, search, validDeviceTypes)
	if err != nil {
		return devices, err, code
	}
	devices = append(devices, modifiedDevices...)
	devices = RemoveDuplicatesF(devices, func(d model.PermSearchDevice) string { return d.Id })
	sort.Slice(devices, func(i, j int) bool {
		nameI := devices[i].DisplayName
		if nameI == "" {
			nameI = devices[i].Name
		}
		nameJ := devices[j].DisplayName
		if nameJ == "" {
			nameJ = devices[j].Name
		}
		return nameI < nameJ
	})

	blockedDevices := map[string]bool{}
	for _, id := range currentDeviceIds {
		blockedDevices[id] = true

		//if a modified version of a device is used, the unmodified version may not be selected
		if pureId, modifier := idmodifier.SplitModifier(id); pureId != id && len(modifier) > 0 {
			blockedDevices[pureId] = true
		}
	}

	filteredDevices := []model.PermSearchDevice{}
	for _, device := range devices {
		if !blockedDevices[device.Id] {
			filteredDevices = append(filteredDevices, device)
		}
	}
	return filteredDevices, err, code
}

func (this *Controller) getDeviceGroupOptionsGetDevicesModified(token string, currentDeviceIds []string, search model.DeviceGroupHelperPagination, validDeviceTypes []string) (devices []model.PermSearchDevice, err error, code int) {
	if currentDeviceIds == nil {
		currentDeviceIds = []string{}
	}

	if len(validDeviceTypes) == 0 {
		deviceTypes, err, code := this.getOnlyDeviceTypesIncludingIdModifier(token)
		if err != nil {
			return devices, err, code
		}
		for _, dt := range deviceTypes {
			validDeviceTypes = append(validDeviceTypes, dt.Id)
		}
	}

	pureIdToModifier := map[string]map[string][]string{}
	searchedDeviceTypeIds := []string{}
	for _, deviceTypeId := range validDeviceTypes {
		pureId, modifier := idmodifier.SplitModifier(deviceTypeId)
		if pureId != deviceTypeId {
			pureIdToModifier[pureId] = modifier
			searchedDeviceTypeIds = append(searchedDeviceTypeIds, pureId)
		}
	}

	unModDevices, err, code := this.devicerepo.ListDevices(token, client.DeviceListOptions{
		DeviceTypeIds: searchedDeviceTypeIds,
		Search:        search.Search,
		Limit:         search.Limit + int64(len(currentDeviceIds)),
		Offset:        search.Offset,
		SortBy:        "name.asc",
	})
	if err != nil {
		return devices, err, code
	}
	modefiedDeviceIds := []string{}
	for _, device := range unModDevices {
		mod, ok := pureIdToModifier[device.DeviceTypeId]
		if ok {
			modefiedDeviceIds = append(modefiedDeviceIds, idmodifier.JoinModifier(device.Id, mod))
		}
	}
	modDevices, _, err, _ := this.devicerepo.ListExtendedDevices(token, client.ExtendedDeviceListOptions{
		Ids:    modefiedDeviceIds,
		SortBy: "name.asc",
	})
	if err != nil {
		return devices, err, code
	}
	for _, device := range modDevices {
		if !slices.Contains(currentDeviceIds, device.Id) {
			devices = append(devices, model.PermSearchDevice{
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
	}
	return devices, nil, 200
}

func (this *Controller) getDeviceGroupOptionsGetDevicesUnmodified(
	token string,
	currentDeviceIds []string,
	search model.DeviceGroupHelperPagination,
	validDeviceTypes []string) (devices []model.PermSearchDevice, err error, code int) {

	if len(validDeviceTypes) == 0 && validDeviceTypes != nil {
		validDeviceTypes = nil
	}
	//trim modified ids
	trimmedCurrentDeviceIds := make([]string, len(currentDeviceIds))
	for i, id := range currentDeviceIds {
		trimmedCurrentDeviceIds[i], _ = idmodifier.SplitModifier(id)
	}

	temp, _, err, _ := this.devicerepo.ListExtendedDevices(token, client.ExtendedDeviceListOptions{
		DeviceTypeIds: validDeviceTypes,
		Search:        search.Search,
		Limit:         search.Limit + int64(len(currentDeviceIds)),
		Offset:        search.Offset,
		SortBy:        "name.asc",
	})
	if err != nil {
		return devices, err, code
	}
	for _, device := range temp {
		if !slices.Contains(trimmedCurrentDeviceIds, device.Id) {
			devices = append(devices, model.PermSearchDevice{
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
	}
	return devices, nil, 200
}

func (this *Controller) getDeviceGroupOptions(
	token string,
	deviceTypeCache *map[string]devicemodel.DeviceType,
	deviceCache *map[string]devicemodel.Device,
	currentDeviceIds []string,
	criteria []devicemodel.DeviceGroupFilterCriteria,
	search model.DeviceGroupHelperPagination,
	maintainGroupUsability bool,
	functionBlockList []string,
) (
	result []model.DeviceGroupOption,
	err error,
	code int,
) {

	devices, err, code := this.getDeviceGroupOptionsGetDevice(token, currentDeviceIds, criteria, search, maintainGroupUsability, functionBlockList)
	if err != nil {
		return result, err, code
	}

	deviceTypeToRemoveCache := map[string][]devicemodel.DeviceGroupFilterCriteria{}
	deviceTypeToCriteriaCache := map[string][]devicemodel.DeviceGroupFilterCriteria{}
	for _, device := range devices {
		option := model.DeviceGroupOption{
			Device:          device,
			RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{},
		}
		deviceCriteria := []devicemodel.DeviceGroupFilterCriteria{}
		if cached, ok := deviceTypeToRemoveCache[device.DeviceTypeId]; ok {
			option.RemovesCriteria = cached
			deviceCriteria = deviceTypeToCriteriaCache[device.DeviceTypeId]
		} else {
			option.RemovesCriteria, deviceCriteria, err, code = this.getDeviceGroupOptionCriteria(token, deviceTypeCache, deviceCache, criteria, option.Device.Id)
			if err != nil {
				return result, err, code
			}
			deviceTypeToRemoveCache[device.DeviceTypeId] = option.RemovesCriteria
			deviceTypeToCriteriaCache[device.DeviceTypeId] = deviceCriteria
		}
		option.MaintainsGroupUsability = len(criteria) > len(option.RemovesCriteria) || (len(currentDeviceIds) == 0 && len(deviceCriteria) > 0)
		result = append(result, option)
	}
	return result, nil, http.StatusOK
}

func (this *Controller) getDeviceGroupOptionCriteria(
	token string,
	deviceTypeCache *map[string]devicemodel.DeviceType,
	deviceCache *map[string]devicemodel.Device,
	currentCriteria []devicemodel.DeviceGroupFilterCriteria,
	deviceId string,
) (
	result []devicemodel.DeviceGroupFilterCriteria,
	deviceCriteria []devicemodel.DeviceGroupFilterCriteria,
	err error,
	code int,
) {
	result = []devicemodel.DeviceGroupFilterCriteria{}
	deviceCriteria, err, code = this.getDeviceCriteria(token, deviceTypeCache, deviceCache, deviceId)
	if err != nil {
		return result, deviceCriteria, err, code
	}
	deviceCriteriaSet := map[string]devicemodel.DeviceGroupFilterCriteria{}
	for _, criteria := range deviceCriteria {
		deviceCriteriaSet[criteriaHash(criteria)] = criteria
	}
	for _, criteria := range currentCriteria {
		if _, ok := deviceCriteriaSet[criteriaHash(criteria)]; !ok {
			result = append(result, criteria)
		}
	}
	return
}

func (this *Controller) getValidDeviceTypesForDeviceGroup(token string, criteria []devicemodel.DeviceGroupFilterCriteria, functionBlockList []string) (deviceTypeIds []string, err error) {
	functionBlockSet := map[string]bool{}
	for _, fId := range functionBlockList {
		functionBlockSet[strings.TrimSpace(fId)] = true
	}
	deviceTypeIds = []string{}
	deviceIdSet := map[string]bool{}
	for _, c := range criteria {
		if !functionBlockSet[c.FunctionId] {
			temp, err := this.cachedGetValidDeviceTypesForDeviceGroupCriteria(token, c)
			if err != nil {
				return deviceTypeIds, err
			}
			for _, id := range temp {
				deviceIdSet[id] = true
			}
		}
	}
	for id, _ := range deviceIdSet {
		deviceTypeIds = append(deviceTypeIds, id)
	}
	return deviceTypeIds, nil
}

func (this *Controller) cachedGetValidDeviceTypesForDeviceGroupCriteria(token string, criteria devicemodel.DeviceGroupFilterCriteria) (deviceTypeIds []string, err error) {
	err = this.cache.Use("dt_by_criteria."+criteria.Short(), func() (interface{}, error) {
		return this.getValidDeviceTypesForDeviceGroupCriteria(token, criteria)
	}, &deviceTypeIds)
	return
}

func (this *Controller) getValidDeviceTypesForDeviceGroupCriteria(token string, criteria devicemodel.DeviceGroupFilterCriteria) (deviceTypeIds []string, err error) {
	descriptions := []client.FilterCriteria{
		{
			Interaction:   models.Interaction(criteria.Interaction),
			FunctionId:    criteria.FunctionId,
			AspectId:      criteria.AspectId,
			DeviceClassId: criteria.DeviceClassId,
		},
	}
	deviceTypes, err, _ := this.getCachedFilteredDeviceTypes(token, descriptions, nil)
	if err != nil {
		return deviceTypeIds, err
	}
	if this.config.Debug {
		log.Println("DEBUG: GetFilteredDevices()::getCachedFilteredDeviceTypes()", deviceTypes)
	}

	for _, dt := range deviceTypes {
		deviceTypeIds = append(deviceTypeIds, dt.Id)
	}
	return deviceTypeIds, nil
}
