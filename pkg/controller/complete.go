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
	"device-selection/pkg/model/basecontentvariable"
	"device-selection/pkg/model/devicemodel"
	"errors"
)

func (this *Controller) CompleteServices(token string, selectables []model.Selectable, filter []devicemodel.FilterCriteria) ([]model.Selectable, error) {
	return this.completeServices(token, selectables, &map[string]devicemodel.DeviceType{}, filter)
}

func (this *Controller) CompleteBulkServices(token string, bulk model.BulkResult, request model.BulkRequest) (_ model.BulkResult, err error) {
	cache := &map[string]devicemodel.DeviceType{}
	for index, element := range bulk {
		bulk[index].Selectables, err = this.completeServices(token, element.Selectables, cache, request[index].Criteria)
		if err != nil {
			return bulk, err
		}
	}
	return bulk, nil
}

func (this *Controller) completeServices(token string, selectables []model.Selectable, cache *map[string]devicemodel.DeviceType, filter []devicemodel.FilterCriteria) (_ []model.Selectable, err error) {
	characteristicIds, err := this.prepareCharacteristicIds(filter, token)
	if err != nil {
		return nil, err
	}
	for selectableIndex, selectable := range selectables {
		selectable.ServicePathOptions = map[string][]model.PathCharacteristicIdPair{}
		if selectable.Device != nil {
			dt, err := this.getCachedTechnicalDeviceType(token, selectable.Device.DeviceTypeId, cache)
			if err != nil {
				return selectables, err
			}
			dtServices := map[string]devicemodel.Service{}
			for _, service := range dt.Services {
				dtServices[service.Id] = service
			}
			for serviceIndex, service := range selectable.Services {
				//merge technical and semantic device-type information
				tdt := dtServices[service.Id]
				tdt.FunctionIds = service.FunctionIds
				tdt.AspectIds = service.AspectIds
				selectable.Services[serviceIndex] = dtServices[service.Id]
				_, ok := selectable.ServicePathOptions[service.Id]
				if !ok {
					var pathCharacteristicPairs []model.PathCharacteristicIdPair
					err = findPathCharacteristicPairs(&dtServices[service.Id].Outputs[0].ContentVariable, &characteristicIds, "value", &pathCharacteristicPairs)
					if err != nil {
						return nil, err
					}
					selectable.ServicePathOptions[service.Id] = pathCharacteristicPairs
				}
			}
			selectables[selectableIndex] = selectable
		} else if selectable.Import != nil {
			fullType, err := this.getFullImportType(token, selectable.ImportType.Id)
			if err != nil {
				return nil, err
			}
			selectable.ImportType = &fullType
			selectables[selectableIndex] = selectable
			_, ok := selectable.ServicePathOptions[fullType.Id]
			if !ok {
				var pathCharacteristicPairs []model.PathCharacteristicIdPair
				for _, subOutput := range fullType.Output.SubContentVariables { // root element has to be ignored to find correct path
					var subPathCharacteristicPairs []model.PathCharacteristicIdPair
					err = findPathCharacteristicPairs(&subOutput, &characteristicIds, "", &subPathCharacteristicPairs)
					if err != nil {
						return nil, err
					}
					pathCharacteristicPairs = append(pathCharacteristicPairs, subPathCharacteristicPairs...)
				}

				selectable.ServicePathOptions[fullType.Id] = pathCharacteristicPairs
			}
		}
	}
	return selectables, nil
}

func findPathCharacteristicPairs(contentVariable basecontentvariable.Descriptor, allowedCharacteristicIds *[]string, prefix string, res *[]model.PathCharacteristicIdPair) (err error) {
	if res == nil || allowedCharacteristicIds == nil || contentVariable == nil {
		return errors.New("encountered nil pointer")
	}
	var path string
	if len(prefix) == 0 {
		path = ""
	} else {
		path = prefix + "."
	}
	path += contentVariable.GetName()
	if contentVariable.GetCharacteristicId() != "" {
		for _, allowedCharacteristicId := range *allowedCharacteristicIds {
			if contentVariable.GetCharacteristicId() == allowedCharacteristicId {
				*res = append(*res, model.PathCharacteristicIdPair{
					Path:             path,
					CharacteristicId: allowedCharacteristicId,
				})
			}
		}
	}
	for _, subContentVariable := range contentVariable.GetSubContentVariables() {
		err = findPathCharacteristicPairs(subContentVariable, allowedCharacteristicIds, path, res)
		if err != nil {
			return
		}
	}
	return
}

func (this *Controller) prepareCharacteristicIds(filters []devicemodel.FilterCriteria, token string) (characteristicIds []string, err error) {
	for _, filter := range filters {
		f, err := this.GetFunction(filter.FunctionId, token)
		if err != nil {
			return nil, err
		}
		c, err := this.GetConcept(f.ConceptId, token)
		if err != nil {
			return nil, err
		}
		characteristicIds = append(characteristicIds, c.CharacteristicIds...)
	}
	return
}
