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
	"errors"
	"github.com/SENERGY-Platform/device-selection/pkg/model"
	"github.com/SENERGY-Platform/device-selection/pkg/model/basecontentvariable"
	"github.com/SENERGY-Platform/device-selection/pkg/model/devicemodel"
	"log"
)

func (this *Controller) CompleteServices(token string, selectables []model.Selectable, filter []devicemodel.FilterCriteria) ([]model.Selectable, error) {
	return this.completeServices(token, selectables, filter)
}

func (this *Controller) CompleteBulkServices(token string, bulk model.BulkResult, request model.BulkRequest) (_ model.BulkResult, err error) {
	for index, element := range bulk {
		bulk[index].Selectables, err = this.completeServices(token, element.Selectables, request[index].Criteria)
		if err != nil {
			return bulk, err
		}
	}
	return bulk, nil
}

func (this *Controller) CompleteBulkServicesV2(token string, bulk model.BulkResult, request model.BulkRequestV2) (_ model.BulkResult, err error) {
	for index, element := range bulk {
		bulk[index].Selectables, err = this.completeServices(token, element.Selectables, request[index].Criteria)
		if err != nil {
			return bulk, err
		}
	}
	return bulk, nil
}

func (this *Controller) completeServices(token string, selectables []model.Selectable, filter []devicemodel.FilterCriteria) (_ []model.Selectable, err error) {
	aspectCache := &map[string]devicemodel.AspectNode{}
	for selectableIndex, selectable := range selectables {
		selectable.ServicePathOptions = map[string][]model.PathOption{}
		if selectable.Device != nil {
			//already fully handled
		} else if selectable.Import != nil {
			fullType, err := this.getFullImportType(token, selectable.ImportType.Id)
			if err != nil {
				return nil, err
			}
			selectable.ImportType = &fullType
			selectables[selectableIndex] = selectable
			_, ok := selectable.ServicePathOptions[fullType.Id]
			if !ok {
				var pathCharacteristicPairs []model.PathOption
				for _, subOutput := range fullType.Output.SubContentVariables { // root element has to be ignored to find correct path
					var subPathCharacteristicPairs []model.PathOption
					err = this.findPathCharacteristicPairs(&subOutput, filter, "", &subPathCharacteristicPairs, token, aspectCache)
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

func (this *Controller) findPathCharacteristicPairs(contentVariable basecontentvariable.Descriptor, filterCriteria []devicemodel.FilterCriteria, prefix string, res *[]model.PathOption, token string, aspectCache *map[string]devicemodel.AspectNode) (err error) {
	if res == nil || contentVariable == nil {
		return errors.New("encountered nil pointer")
	}
	var path string
	if len(prefix) == 0 {
		path = ""
	} else {
		path = prefix + "."
	}
	path += contentVariable.GetName()

	ok, err := this.contentVariableContainsAnyCriteria(contentVariable, filterCriteria, token, aspectCache)
	if err != nil {
		return err
	}
	if ok {
		aspectNode, err := this.getAspectNodeWithCache(token, aspectCache, contentVariable.GetAspectId())
		if err != nil {
			return err
		}
		*res = append(*res, model.PathOption{
			Path:             path,
			CharacteristicId: contentVariable.GetCharacteristicId(),
			AspectNode:       aspectNode,
			FunctionId:       contentVariable.GetFunctionId(),
			IsVoid:           contentVariable.GetIsVoid(),
		})
	}
	for _, subContentVariable := range contentVariable.GetSubContentVariables() {
		err = this.findPathCharacteristicPairs(subContentVariable, filterCriteria, path, res, token, aspectCache)
		if err != nil {
			return
		}
	}
	return
}

func (this *Controller) contentVariableContainsAnyCriteria(variable basecontentvariable.Descriptor, criteria []devicemodel.FilterCriteria, token string, aspectCache *map[string]devicemodel.AspectNode) (result bool, err error) {
	for _, c := range criteria {
		temp, err := this.contentVariableContainsCriteria(variable, c, token, aspectCache)
		if err != nil {
			return result, err
		}
		if temp {
			return true, nil
		}
	}
	return false, nil
}

func (this *Controller) contentVariableContainsCriteria(variable basecontentvariable.Descriptor, criteria devicemodel.FilterCriteria, token string, aspectCache *map[string]devicemodel.AspectNode) (result bool, err error) {
	aspectNode := devicemodel.AspectNode{}
	if criteria.AspectId != "" {
		aspectNode, err = this.getAspectNodeWithCache(token, aspectCache, criteria.AspectId)
		if err != nil {
			return false, err
		}
	}
	if variable.GetFunctionId() == criteria.FunctionId &&
		(criteria.AspectId == "" ||
			variable.GetAspectId() == criteria.AspectId ||
			listContains(aspectNode.DescendentIds, variable.GetAspectId())) {
		return true, nil
	}
	return false, nil
}

func (this *Controller) getAspectNodeWithCache(token string, aspectCache *map[string]devicemodel.AspectNode, aspectId string) (aspectNode devicemodel.AspectNode, err error) {
	var ok bool
	aspectNode, ok = (*aspectCache)[aspectId]
	if !ok {
		aspectNode, err = this.GetAspectNode(aspectId, token)
		if err != nil {
			log.Println("WARNING: unable to load aspect node", aspectId, err)
			return aspectNode, err
		}
		(*aspectCache)[aspectId] = aspectNode
	}
	return aspectNode, nil
}

func listContains(list []string, search string) bool {
	for _, element := range list {
		if element == search {
			return true
		}
	}
	return false
}
