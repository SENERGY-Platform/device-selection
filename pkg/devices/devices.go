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
	"context"
	"device-selection/pkg/configuration"
	"device-selection/pkg/model"
	"device-selection/pkg/model/devicemodel"
	"log"
)

type Devices struct {
	config configuration.Config
}

func New(ctx context.Context, config configuration.Config) (*Devices, error) {
	return &Devices{
		config: config,
	}, nil
}

func (this *Devices) GetBlockedProtocols(token string, interaction devicemodel.Interaction) (result []string, err error, code int) {
	protocols, err, _ := this.GetProtocols(token)
	if err != nil {
		return result, err, code
	}
	result = this.FilterProtocols(protocols, interaction)
	return result, nil, 200
}

func (this *Devices) FilterProtocols(protocols []devicemodel.Protocol, filterBy devicemodel.Interaction) (result []string) {
	for _, protocol := range protocols {
		if protocol.Interaction == filterBy {
			result = append(result, protocol.Id)
		}
	}
	return result
}

func (this *Devices) GetFilteredDevices(token string, descriptions model.DeviceTypesFilter, protocolBlockList []string) (result []model.Selectable, err error, code int) {
	if len(descriptions) == 0 {
		return []model.Selectable{}, nil, 200
	}
	filteredProtocols := map[string]bool{}
	for _, protocolId := range protocolBlockList {
		filteredProtocols[protocolId] = true
	}
	deviceTypes, err, code := this.GetFilteredDeviceTypes(token, descriptions)
	if err != nil {
		return result, err, code
	}
	if this.config.Debug {
		log.Println("DEBUG: GetFilteredDevices()::GetFilteredDeviceTypes()", deviceTypes)
	}
	for _, dt := range deviceTypes {
		services := []devicemodel.Service{}
		serviceIndex := map[string]devicemodel.Service{}
		for _, service := range dt.Services {
			for _, desc := range descriptions {
				for _, function := range service.Functions {
					if !(function.RdfType == devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION && filteredProtocols[service.ProtocolId]) {
						if function.Id == desc.FunctionId {
							if desc.AspectId == "" {
								serviceIndex[service.Id] = service
							} else {
								for _, aspect := range service.Aspects {
									if aspect.Id == desc.AspectId {
										serviceIndex[service.Id] = service
									}
								}
							}
						}
					}
				}
			}
		}
		for _, service := range serviceIndex {
			services = append(services, service)
		}
		if len(services) > 0 {
			devices, err, code := this.GetDevicesOfType(token, dt.Id)
			if err != nil {
				return result, err, code
			}
			if this.config.Debug {
				log.Println("DEBUG: GetFilteredDevices()::GetDevicesOfType()", dt.Id, devices)
			}
			for _, device := range devices {
				result = append(result, model.Selectable{
					Device:   device,
					Services: services,
				})
			}
		}
	}
	if this.config.Debug {
		log.Println("DEBUG: GetFilteredDevices()", result)
	}
	return result, nil, 200
}
