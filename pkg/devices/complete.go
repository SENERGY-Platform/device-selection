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
	"device-selection/pkg/model"
	"device-selection/pkg/model/devicemodel"
)

func (this *Devices) CompleteServices(token string, selectables []model.Selectable) ([]model.Selectable, error) {
	return this.completeServices(token, selectables, &map[string]devicemodel.DeviceType{})
}

func (this *Devices) CompleteBulkServices(token string, bulk model.BulkResult) (_ model.BulkResult, err error) {
	cache := &map[string]devicemodel.DeviceType{}
	for index, element := range bulk {
		bulk[index].Selectables, err = this.completeServices(token, element.Selectables, cache)
		if err != nil {
			return bulk, err
		}
	}
	return bulk, nil
}

func (this *Devices) completeServices(token string, selectables []model.Selectable, cache *map[string]devicemodel.DeviceType) (_ []model.Selectable, err error) {
	for selectableIndex, selectable := range selectables {
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
		}
		selectables[selectableIndex] = selectable
	}
	return selectables, nil
}
