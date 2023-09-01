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

package model

import "github.com/SENERGY-Platform/device-selection/pkg/model/devicemodel"

type PermSearchDevice struct {
	devicemodel.Device
	DisplayName string      `json:"display_name"`
	Permissions Permissions `json:"permissions"`
	Shared      bool        `json:"shared"`
	Creator     string      `json:"creator"`
}

type Permissions struct {
	R bool `json:"r"`
	W bool `json:"w"`
	X bool `json:"x"`
	A bool `json:"a"`
}

type Selectable struct {
	Device             *PermSearchDevice       `json:"device,omitempty"`
	Services           []devicemodel.Service   `json:"services,omitempty"`
	DeviceGroup        *DeviceGroup            `json:"device_group,omitempty"`
	Import             *Import                 `json:"import,omitempty"`
	ImportType         *ImportType             `json:"importType,omitempty"`
	ServicePathOptions map[string][]PathOption `json:"servicePathOptions,omitempty"`
}

type DeviceGroup struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type FilterCriteriaAndSet []devicemodel.FilterCriteria

type FilterCriteriaOrSet []devicemodel.FilterCriteria

type BulkRequestElement struct {
	Id                string                   `json:"id"`
	FilterInteraction *devicemodel.Interaction `json:"filter_interaction"`
	FilterProtocols   []string                 `json:"filter_protocols"`
	Criteria          FilterCriteriaAndSet     `json:"criteria"`
	IncludeGroups     bool                     `json:"include_groups"`
	IncludeImports    bool                     `json:"include_imports"`
	LocalDevices      []string                 `json:"local_devices"`
}

type BulkRequest []BulkRequestElement

type BulkRequestElementV2 struct {
	Id                       string               `json:"id"`
	Criteria                 FilterCriteriaAndSet `json:"criteria"`
	IncludeGroups            bool                 `json:"include_groups"`
	IncludeImports           bool                 `json:"include_imports"`
	IncludeDevices           bool                 `json:"include_devices"`
	IncludeIdModifiedDevices bool                 `json:"include_id_modified_devices"`
	LocalDevices             []string             `json:"local_devices"`
}

type BulkRequestV2 []BulkRequestElementV2

type BulkResult []BulkResultElement

type BulkResultElement struct {
	Id          string       `json:"id"`
	Selectables []Selectable `json:"selectables"`
}

type DeviceGroupHelperResult struct {
	Criteria []devicemodel.DeviceGroupFilterCriteria `json:"criteria"`
	Options  []DeviceGroupOption                     `json:"options"`
}

type DeviceGroupOption struct {
	Device                  PermSearchDevice                        `json:"device"`
	RemovesCriteria         []devicemodel.DeviceGroupFilterCriteria `json:"removes_criteria"`
	MaintainsGroupUsability bool                                    `json:"maintains_group_usability"`
}

type PathOption struct {
	Path             string                     `json:"path"`
	CharacteristicId string                     `json:"characteristicId"`
	AspectNode       devicemodel.AspectNode     `json:"aspectNode"`
	FunctionId       string                     `json:"functionId"`
	IsVoid           bool                       `json:"isVoid"`
	Value            interface{}                `json:"value,omitempty"`
	Type             Type                       `json:"type,omitempty"`
	Configurables    []devicemodel.Configurable `json:"configurables,omitempty"`
	Interaction      devicemodel.Interaction    `json:"interaction,omitempty"`
}
