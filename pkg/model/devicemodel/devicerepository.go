/*
 * Copyright 2022 InfAI (CC SES)
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

package devicemodel

type AspectNode struct {
	Id            string   `json:"id"`
	Name          string   `json:"name"`
	RootId        string   `json:"root_id"`
	ParentId      string   `json:"parent_id"`
	ChildIds      []string `json:"child_ids"`
	AncestorIds   []string `json:"ancestor_ids"`
	DescendentIds []string `json:"descendent_ids"`
}

type DeviceTypeSelectable struct {
	DeviceTypeId       string                         `json:"device_type_id,omitempty"`
	Services           []Service                      `json:"services,omitempty"`
	ServicePathOptions map[string][]ServicePathOption `json:"service_path_options,omitempty"`
}

type ServicePathOption struct {
	ServiceId             string         `json:"service_id"`
	Path                  string         `json:"path"`
	CharacteristicId      string         `json:"characteristic_id"`
	AspectNode            AspectNode     `json:"aspect_node"`
	FunctionId            string         `json:"function_id"`
	IsVoid                bool           `json:"is_void"`
	Value                 interface{}    `json:"value,omitempty"`
	IsControllingFunction bool           `json:"is_controlling_function"`
	Configurables         []Configurable `json:"configurables,omitempty"`
	Type                  Type           `json:"type,omitempty"`
	Interaction           Interaction    `json:"interaction"`
}

type Configurable struct {
	Path             string      `json:"path"`
	CharacteristicId string      `json:"characteristic_id"`
	AspectNode       AspectNode  `json:"aspect_node"`
	FunctionId       string      `json:"function_id"`
	Value            interface{} `json:"value,omitempty"`
	Type             Type        `json:"type,omitempty"`
}
