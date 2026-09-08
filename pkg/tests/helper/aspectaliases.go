/*
 * Copyright 2026 InfAI (CC SES)
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

package helper

import (
	"github.com/SENERGY-Platform/device-selection/pkg/model"
	"github.com/SENERGY-Platform/device-selection/pkg/model/devicemodel"
)

//A content variable, a filter criteria, a path option and a configurable each carry a single
//aspect field that is the deprecated alias for a list. The fixtures of the payload comparing
//tests are written with those aliases, while a result carries the lists they stand for. The
//expansion below is applied to the expectation only, so that a result which leaves a list
//unset still fails.

// ExpandExpectedSelectables expands the aspect aliases of expected selectables.
func ExpandExpectedSelectables(selectables []model.Selectable) []model.Selectable {
	for i := range selectables {
		for j := range selectables[i].Services {
			expandExpectedService(&selectables[i].Services[j])
		}
		for serviceId, options := range selectables[i].ServicePathOptions {
			for k := range options {
				expandExpectedPathOption(&options[k])
			}
			selectables[i].ServicePathOptions[serviceId] = options
		}
	}
	return selectables
}

// ExpandExpectedDeviceTypeSelectables expands the aspect aliases of expected device-type
// selectables, which carry the same services and path options without a device.
func ExpandExpectedDeviceTypeSelectables(selectables []devicemodel.DeviceTypeSelectable) []devicemodel.DeviceTypeSelectable {
	for i := range selectables {
		for j := range selectables[i].Services {
			expandExpectedService(&selectables[i].Services[j])
		}
		for serviceId, options := range selectables[i].ServicePathOptions {
			for k := range options {
				expandExpectedServicePathOption(&options[k])
			}
			selectables[i].ServicePathOptions[serviceId] = options
		}
	}
	return selectables
}

func expandExpectedService(service *devicemodel.Service) {
	for i := range service.Inputs {
		expandExpectedContentVariable(&service.Inputs[i].ContentVariable)
	}
	for i := range service.Outputs {
		expandExpectedContentVariable(&service.Outputs[i].ContentVariable)
	}
}

func expandExpectedContentVariable(variable *devicemodel.ContentVariable) {
	variable.AspectIds = devicemodel.AspectIds(variable.AspectId, variable.AspectIds)
	for i := range variable.SubContentVariables {
		expandExpectedContentVariable(&variable.SubContentVariables[i])
	}
}

func expandExpectedPathOption(option *model.PathOption) {
	option.AspectNodes = expandExpectedAspectNodes(option.AspectNode, option.AspectNodes)
	for i := range option.Configurables {
		expandExpectedConfigurable(&option.Configurables[i])
	}
}

func expandExpectedServicePathOption(option *devicemodel.ServicePathOption) {
	option.AspectNodes = expandExpectedAspectNodes(option.AspectNode, option.AspectNodes)
	for i := range option.Configurables {
		expandExpectedConfigurable(&option.Configurables[i])
	}
}

func expandExpectedConfigurable(configurable *devicemodel.Configurable) {
	configurable.AspectNodes = expandExpectedAspectNodes(configurable.AspectNode, configurable.AspectNodes)
}

func expandExpectedAspectNodes(node devicemodel.AspectNode, nodes []devicemodel.AspectNode) []devicemodel.AspectNode {
	if node.Id == "" || len(nodes) > 0 {
		return nodes
	}
	return []devicemodel.AspectNode{node}
}
