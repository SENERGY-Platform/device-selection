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

package legacy

import "device-selection/pkg/model"

type ImportType struct {
	Id             string                      `json:"id"`
	Name           string                      `json:"name"`
	Description    string                      `json:"description"`
	Image          string                      `json:"image"`
	DefaultRestart bool                        `json:"default_restart"`
	Configs        []model.ImportTypeConfig    `json:"configs"`
	AspectIds      interface{}                 `json:"aspect_ids"` // permSearch gives string or []string
	Output         model.ImportContentVariable `json:"output,omitempty"`
	FunctionIds    interface{}                 `json:"function_ids"` // permSearch gives string or []string
	Owner          string                      `json:"owner,omitempty"`
}

func FromLegacyImportTypePointer(importType ImportType) *model.ImportType {
	e := FromLegacyImportType(importType)
	return &e
}

func FromLegacyImportType(importType ImportType) model.ImportType {
	//TODO
	return model.ImportType{}
}
