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
	"log"
	"net/http"
)

func (this *Controller) DeviceGroupHelper(token string, deviceIds []string, search model.QueryFind, maintainGroupUsability bool) (result model.DeviceGroupHelperResult, err error, code int) {
	deviceCache := &map[string]devicemodel.Device{}
	deviceTypeCache := &map[string]devicemodel.DeviceType{}
	result.Criteria, err, code = this.getDeviceGroupCriteria(token, deviceTypeCache, deviceCache, deviceIds)
	if err != nil {
		return
	}
	result.Options, err, code = this.getDeviceGroupOptions(token, deviceTypeCache, deviceCache, deviceIds, result.Criteria, search, maintainGroupUsability)
	return result, err, code
}

func (this *Controller) getDeviceGroupCriteria(token string, deviceTypeCache *map[string]devicemodel.DeviceType, deviceCache *map[string]devicemodel.Device, deviceIds []string) (result []devicemodel.DeviceGroupFilterCriteria, err error, code int) {
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
	device, err, code := this.getCachedTechnicalDevice(token, deviceId, deviceCache)
	if err != nil {
		return result, err, code
	}
	deviceType, err := this.getCachedTechnicalDeviceType(token, device.DeviceTypeId, deviceTypeCache)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	resultSet := map[string]devicemodel.DeviceGroupFilterCriteria{}
	for _, service := range deviceType.Services {
		interactions := []devicemodel.Interaction{service.Interaction}
		if service.Interaction == devicemodel.EVENT_AND_REQUEST {
			interactions = []devicemodel.Interaction{devicemodel.EVENT, devicemodel.REQUEST}
		}

		for _, functionId := range service.FunctionIds {
			for _, interaction := range interactions {
				if isMeasuringFunctionId(functionId) {
					for _, aspectId := range service.AspectIds {
						criteria := devicemodel.DeviceGroupFilterCriteria{
							FunctionId:  functionId,
							AspectId:    aspectId,
							Interaction: interaction,
						}
						resultSet[criteriaHash(criteria)] = criteria
					}
				} else {
					criteria := devicemodel.DeviceGroupFilterCriteria{
						FunctionId:    functionId,
						DeviceClassId: deviceType.DeviceClassId,
						Interaction:   interaction,
					}
					resultSet[criteriaHash(criteria)] = criteria
				}
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

func (this *Controller) getDeviceGroupOptions(
	token string,
	deviceTypeCache *map[string]devicemodel.DeviceType,
	deviceCache *map[string]devicemodel.Device,
	currentDeviceIds []string,
	criteria []devicemodel.DeviceGroupFilterCriteria,
	search model.QueryFind,
	maintainGroupUsability bool,
) (
	result []model.DeviceGroupOption,
	err error,
	code int,
) {

	devices := []model.PermSearchDevice{}

	filter := []model.Selection{
		{
			Not: &model.Selection{
				Condition: model.ConditionConfig{
					Feature:   "id",
					Operation: model.QueryAnyValueInFeatureOperation,
					Value:     currentDeviceIds,
				},
			},
		},
	}

	if maintainGroupUsability && len(criteria) > 0 {
		validDeviceTypes, err := this.getValidDeviceTypesForDeviceGroup(token, criteria)
		if err != nil {
			log.Println("ERROR: getValidDeviceTypesForDeviceGroup()", err)
			err = nil
		} else {
			filter = append(filter, model.Selection{
				Condition: model.ConditionConfig{
					Feature:   "device_type_id",
					Operation: model.QueryAnyValueInFeatureOperation,
					Value:     validDeviceTypes,
				},
			})
		}
	}

	search.Filter = &model.Selection{
		And: filter,
	}

	err, code = this.Search(token, model.QueryMessage{
		Resource: "devices",
		Find:     &search,
	}, &devices)
	if err != nil {
		return result, err, code
	}

	deviceTypeToRemoveCache := map[string][]devicemodel.DeviceGroupFilterCriteria{}
	deviceTypeToCriteriaCache := map[string][]devicemodel.DeviceGroupFilterCriteria{}
	for _, device := range devices {
		option := model.DeviceGroupOption{
			Device:          device.Device,
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

func (this *Controller) getValidDeviceTypesForDeviceGroup(token string, criteria []devicemodel.DeviceGroupFilterCriteria) (deviceTypeIds []string, err error) {
	deviceTypeIds = []string{}
	deviceIdSet := map[string]bool{}
	for _, c := range criteria {
		temp, err := this.getValidDeviceTypesForDeviceGroupCriteria(token, c)
		if err != nil {
			return deviceTypeIds, err
		}
		for _, id := range temp {
			deviceIdSet[id] = true
		}
	}
	for id, _ := range deviceIdSet {
		deviceTypeIds = append(deviceTypeIds, id)
	}
	return deviceTypeIds, nil
}

func (this *Controller) getValidDeviceTypesForDeviceGroupCriteria(token string, criteria devicemodel.DeviceGroupFilterCriteria) (deviceTypeIds []string, err error) {
	descriptions := model.FilterCriteriaAndSet{{
		FunctionId:    criteria.FunctionId,
		AspectId:      criteria.AspectId,
		DeviceClassId: criteria.DeviceClassId,
	}}
	deviceTypes, err, _ := this.getCachedFilteredDeviceTypes(token, descriptions, nil)
	if err != nil {
		return deviceTypeIds, err
	}
	if this.config.Debug {
		log.Println("DEBUG: GetFilteredDevices()::getCachedFilteredDeviceTypes()", deviceTypes)
	}
	for _, dt := range deviceTypes {
		serviceIndex := map[string]devicemodel.Service{}

		//repeat criteria filtering locally to add interaction filtering
		for _, service := range dt.Services {
			if criteria.Interaction == service.Interaction ||
				(service.Interaction == devicemodel.EVENT_AND_REQUEST && (criteria.Interaction == devicemodel.REQUEST || criteria.Interaction == devicemodel.EVENT)) {
				for _, desc := range descriptions {
					for _, functionId := range service.FunctionIds {
						if functionId == desc.FunctionId {
							if desc.AspectId == "" {
								serviceIndex[service.Id] = service
							} else {
								for _, aspect := range service.AspectIds {
									if aspect == desc.AspectId {
										serviceIndex[service.Id] = service
									}
								}
							}
						}
					}
				}
			}
		}
		if len(serviceIndex) > 0 {
			deviceTypeIds = append(deviceTypeIds, dt.Id)
		}
	}
	return deviceTypeIds, nil
}
