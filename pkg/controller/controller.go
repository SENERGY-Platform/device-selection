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
	"context"
	"device-selection/pkg/configuration"
	"device-selection/pkg/controller/cache"
	"device-selection/pkg/controller/cacheinvalidator"
	"device-selection/pkg/model"
	"device-selection/pkg/model/devicemodel"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"
)

type Controller struct {
	config configuration.Config
	cache  cache.Cache
}

func New(ctx context.Context, config configuration.Config) (*Controller, error) {
	c := cache.New(config.MemcachedUrls)
	if config.KafkaUrl != "" && config.KafkaConsumerGroup != "" && len(config.KafkaTopicsForCacheInvalidation) > 0 {
		log.Println("start listeners to invalidate cache on kafka message to:", config.KafkaTopicsForCacheInvalidation)
		err := cacheinvalidator.StartCacheInvalidator(ctx, config, c)
		if err != nil {
			return nil, err
		}
	}
	return &Controller{
		config: config,
		cache:  c,
	}, nil
}

func (this *Controller) GetFilteredDevices(token string, descriptions model.FilterCriteriaAndSet, protocolBlockList []string, blockedInteraction devicemodel.Interaction, includeGroups bool, includeImports bool, withLocalDeviceIds []string) (result []model.Selectable, err error, code int) {
	return this.getFilteredDevices(token, descriptions, protocolBlockList, blockedInteraction, nil, includeGroups, includeImports, withLocalDeviceIds)
}

func (this *Controller) GetFilteredDevicesV2(token string, descriptions model.FilterCriteriaAndSet, includeDevices bool, includeGroups bool, includeImports bool, withLocalDeviceIds []string, includeIdModified bool) (result []model.Selectable, err error, code int) {
	return this.getFilteredDevicesV2(token, descriptions, nil, includeDevices, includeGroups, includeImports, withLocalDeviceIds, includeIdModified)
}

func (this *Controller) BulkGetFilteredDevices(token string, requests model.BulkRequest) (result model.BulkResult, err error, code int) {
	devicesByDeviceTypeCache := map[string][]model.PermSearchDevice{}
	for _, request := range requests {
		resultElement, err, code := this.handleBulkRequestElement(token, request, &devicesByDeviceTypeCache)
		if err != nil {
			return result, err, code
		}
		result = append(result, resultElement)
	}
	return result, nil, http.StatusOK
}

func (this *Controller) BulkGetFilteredDevicesV2(token string, requests model.BulkRequestV2) (result model.BulkResult, err error, code int) {
	devicesByDeviceTypeCache := map[string][]model.PermSearchDevice{}
	for _, request := range requests {
		resultElement, err, code := this.handleBulkRequestElementV2(token, request, &devicesByDeviceTypeCache)
		if err != nil {
			return result, err, code
		}
		result = append(result, resultElement)
	}
	return result, nil, http.StatusOK
}

func (this *Controller) handleBulkRequestElement(
	token string,
	request model.BulkRequestElement,
	devicesByDeviceTypeCache *map[string][]model.PermSearchDevice,
) (
	result model.BulkResultElement,
	err error,
	code int,
) {

	var blockedInteraction devicemodel.Interaction = ""
	if request.FilterInteraction != nil {
		blockedInteraction = *request.FilterInteraction
	}

	protocolBlockList := request.FilterProtocols
	selectables, err, code := this.getFilteredDevices(token, request.Criteria, protocolBlockList, blockedInteraction, devicesByDeviceTypeCache, request.IncludeGroups, request.IncludeImports, request.LocalDevices)
	if err != nil {
		return result, err, code
	}
	return model.BulkResultElement{
		Id:          request.Id,
		Selectables: selectables,
	}, nil, http.StatusOK
}

func (this *Controller) handleBulkRequestElementV2(
	token string,
	request model.BulkRequestElementV2,
	devicesByDeviceTypeCache *map[string][]model.PermSearchDevice,
) (
	result model.BulkResultElement,
	err error,
	code int,
) {
	selectables, err, code := this.getFilteredDevicesV2(token,
		request.Criteria,
		devicesByDeviceTypeCache,
		request.IncludeDevices,
		request.IncludeGroups,
		request.IncludeImports,
		request.LocalDevices,
		request.IncludeIdModifiedDevices)
	if err != nil {
		return result, err, code
	}
	return model.BulkResultElement{
		Id:          request.Id,
		Selectables: selectables,
	}, nil, http.StatusOK
}

func (this *Controller) getFilteredDevices(
	token string,
	descriptions model.FilterCriteriaAndSet,
	protocolBlockList []string,
	blockedInteraction devicemodel.Interaction,
	devicesByDeviceTypeCache *map[string][]model.PermSearchDevice,
	includeGroups bool,
	includeImports bool,
	withLocalDeviceIds []string,
) (
	result []model.Selectable,
	err error,
	code int,
) {

	if len(descriptions) == 0 {
		return []model.Selectable{}, nil, 200
	}
	filteredProtocols := map[string]bool{}
	for _, protocolId := range protocolBlockList {
		filteredProtocols[protocolId] = true
	}

	deviceTypeSelectables, err := this.GetDeviceTypeSelectablesCached(token, descriptions)
	if err != nil {
		return result, err, code
	}
	for _, dtSelectable := range deviceTypeSelectables {
		servicesProtocolBlock := map[string]bool{}
		servicesBlockedByInteraction := map[string]bool{}
		for _, service := range dtSelectable.Services {
			if service.Interaction == blockedInteraction {
				servicesBlockedByInteraction[service.Id] = true
			}
			if filteredProtocols[service.ProtocolId] {
				servicesProtocolBlock[service.Id] = true
			}
		}
		pathOptions := getServicePathOptionsFromDeviceRepoResult(dtSelectable.ServicePathOptions, servicesProtocolBlock, servicesBlockedByInteraction)
		usedServices := []devicemodel.Service{}
		for serviceId, _ := range pathOptions {
			for _, service := range dtSelectable.Services {
				if serviceId == service.Id {
					usedServices = append(usedServices, service)
					break
				}
			}
		}
		if len(usedServices) > 0 {
			var devices []model.PermSearchDevice
			if len(withLocalDeviceIds) == 0 {
				devices, err, code = this.getCachedDevicesOfType(token, dtSelectable.DeviceTypeId, devicesByDeviceTypeCache)
			} else {
				devices, err, code = this.getCachedDevicesOfTypeFilteredByLocalIdList(token, dtSelectable.DeviceTypeId, devicesByDeviceTypeCache, withLocalDeviceIds)
			}
			if err != nil {
				return result, err, code
			}
			if this.config.Debug {
				log.Println("DEBUG: GetFilteredDevices()::getDevicesOfType()", dtSelectable.DeviceTypeId, devices)
			}
			for _, device := range devices {
				temp := device //make copy to prevent that Selectable.Device is the last element of devices every time
				result = append(result, model.Selectable{
					Device:             &temp,
					Services:           usedServices,
					ServicePathOptions: pathOptions,
				})
			}
		}
	}
	var expectedInteraction = devicemodel.REQUEST
	switch blockedInteraction {
	case devicemodel.REQUEST:
		expectedInteraction = devicemodel.EVENT
	case devicemodel.EVENT:
		expectedInteraction = devicemodel.REQUEST
	case devicemodel.EVENT_AND_REQUEST:
		expectedInteraction = ""
	case "":
		expectedInteraction = devicemodel.EVENT_AND_REQUEST
	}
	if includeGroups {
		groupResult, err, code := this.getFilteredDeviceGroups(token, descriptions, expectedInteraction)
		if err != nil {
			return result, err, code
		}
		result = append(result, groupResult...)
	}
	if includeImports && (expectedInteraction == devicemodel.EVENT || expectedInteraction == devicemodel.EVENT_AND_REQUEST) {
		if this.config.Debug {
			log.Println("DEBUG: GetFilteredDevices() Loading matching imports")
		}
		importResult, err, code := this.getFilteredImports(token, descriptions)
		if err != nil {
			return result, err, code
		}
		result = append(result, importResult...)
	} else if this.config.Debug {
		log.Println("DEBUG: GetFilteredDevices() Not loading imports")
	}
	if this.config.Debug {
		log.Println("DEBUG: GetFilteredDevices()", result)
	}
	return result, nil, 200
}

func (this *Controller) getFilteredDevicesV2(
	token string,
	descriptions model.FilterCriteriaAndSet,
	devicesByDeviceTypeCache *map[string][]model.PermSearchDevice,
	includeDevices bool,
	includeGroups bool,
	includeImports bool,
	withLocalDeviceIds []string,
	includeIdModified bool,
) (
	result []model.Selectable,
	err error,
	code int,
) {
	if len(descriptions) == 0 {
		return []model.Selectable{}, nil, 200
	}
	if this.config.Debug {
		temp, _ := json.Marshal(descriptions)
		log.Println("DEBUG: getFilteredDevicesV2() inputs:", includeDevices, includeGroups, includeImports, string(temp), withLocalDeviceIds)
	}
	if includeDevices {
		deviceTypeSelectables, err := this.GetDeviceTypeSelectablesCachedV2(token, descriptions, includeIdModified)
		if err != nil {
			return result, err, code
		}
		if this.config.Debug {
			log.Println("DEBUG: getFilteredDevicesV2()::GetDeviceTypeSelectablesCachedV2()", len(deviceTypeSelectables))
		}
		for _, dtSelectable := range deviceTypeSelectables {
			pathOptions := getServicePathOptionsFromDeviceRepoResultV2(dtSelectable.ServicePathOptions)
			usedServices := []devicemodel.Service{}
			for serviceId, _ := range pathOptions {
				for _, service := range dtSelectable.Services {
					if serviceId == service.Id {
						usedServices = append(usedServices, service)
						break
					}
				}
			}
			if len(usedServices) > 0 {
				var devices []model.PermSearchDevice
				if len(withLocalDeviceIds) == 0 {
					devices, err, code = this.getCachedDevicesOfType(token, dtSelectable.DeviceTypeId, devicesByDeviceTypeCache)
					if this.config.Debug {
						log.Println("DEBUG: getFilteredDevicesV2()::getCachedDevicesOfType()", dtSelectable.DeviceTypeId, len(devices))
					}
				} else {
					devices, err, code = this.getCachedDevicesOfTypeFilteredByLocalIdList(token, dtSelectable.DeviceTypeId, devicesByDeviceTypeCache, withLocalDeviceIds)
				}
				if err != nil {
					return result, err, code
				}
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
				for _, device := range devices {
					temp := device //make copy to prevent that Selectable.Device is the last element of devices every time
					result = append(result, model.Selectable{
						Device:             &temp,
						Services:           usedServices,
						ServicePathOptions: pathOptions,
					})
				}
			}
		}
	}
	if includeGroups {
		groupResult, err, code := this.getFilteredDeviceGroupsV2(token, descriptions)
		if err != nil {
			return result, err, code
		}
		result = append(result, groupResult...)
	}
	if includeImports && !criteriaContainRequestInteraction(descriptions) {
		importResult, err, code := this.getFilteredImportsV2(token, descriptions)
		if err != nil {
			return result, err, code
		}
		result = append(result, importResult...)
	}
	if this.config.Debug {
		log.Println("DEBUG: GetFilteredDevices()", result)
	}

	for i, e := range result {
		sort.Slice(e.Services, func(i, j int) bool {
			return e.Services[i].Id < e.Services[j].Id
		})
		result[i] = e
	}

	return result, nil, http.StatusOK
}

func criteriaContainRequestInteraction(criteria model.FilterCriteriaAndSet) bool {
	for _, c := range criteria {
		if devicemodel.Interaction(c.Interaction) == devicemodel.REQUEST {
			return true
		}
	}
	return false
}

func getServicePathOptionsFromDeviceRepoResult(in map[string][]devicemodel.ServicePathOption, serviceBlocketByProtocolIndex map[string]bool, serviceBlocketByInteractionIndex map[string]bool) (out map[string][]model.PathOption) {
	out = map[string][]model.PathOption{}
	for serviceId, list := range in {
		if !serviceBlocketByInteractionIndex[serviceId] {
			temp := []model.PathOption{}
			for _, element := range list {
				if !(isMeasuringFunctionId(element.FunctionId) && serviceBlocketByProtocolIndex[serviceId]) { //legacy check; should be covered by interaction check
					temp = append(temp, model.PathOption{
						Path:             element.Path,
						CharacteristicId: element.CharacteristicId,
						AspectNode:       element.AspectNode,
						FunctionId:       element.FunctionId,
						IsVoid:           element.IsVoid,
						Value:            element.Value,
						Type:             element.Type,
						Configurables:    element.Configurables,
					})
				}
			}
			if len(temp) > 0 {
				out[serviceId] = temp
			}
		}
	}
	return out
}

func getServicePathOptionsFromDeviceRepoResultV2(in map[string][]devicemodel.ServicePathOption) (out map[string][]model.PathOption) {
	out = map[string][]model.PathOption{}
	for serviceId, list := range in {
		temp := []model.PathOption{}
		for _, element := range list {
			temp = append(temp, model.PathOption{
				Path:             element.Path,
				CharacteristicId: element.CharacteristicId,
				AspectNode:       element.AspectNode,
				FunctionId:       element.FunctionId,
				IsVoid:           element.IsVoid,
				Value:            element.Value,
				Type:             element.Type,
				Configurables:    element.Configurables,
				Interaction:      element.Interaction,
			})
		}
		if len(temp) > 0 {
			out[serviceId] = temp
		}
	}
	return out
}

func (this *Controller) CombinedDevices(bulk model.BulkResult) (result []model.PermSearchDevice) {
	seen := map[string]bool{}
	for _, bulkElement := range bulk {
		for _, selectable := range bulkElement.Selectables {
			if selectable.Device != nil && !seen[selectable.Device.Id] {
				seen[selectable.Device.Id] = true
				result = append(result, *selectable.Device)
			}
		}
	}
	return
}

func isMeasuringFunctionId(id string) bool {
	if strings.HasPrefix(id, devicemodel.MEASURING_FUNCTION_PREFIX) {
		return true
	}
	return false
}
