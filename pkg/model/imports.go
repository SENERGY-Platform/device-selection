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

import (
	"github.com/SENERGY-Platform/device-selection/v2/pkg/model/basecontentvariable"
	"github.com/SENERGY-Platform/device-selection/v2/pkg/model/devicemodel"
	"github.com/SENERGY-Platform/models/go/models"
)

// The import shapes live in the shared model, so the declarations here were duplicates of
// what the wire already carries. The import-repository and the import-deploy alias the same
// types since import-repository v0.1.1, which makes the three of them one type instead of
// three structurally equal ones - the cast layer this replaced dropped every field the copy
// here did not name, Cost among them.

type ImportType = models.ImportType

type ImportContentVariable = models.ImportContentVariable

type ImportTypeConfig = models.ImportTypeConfig

type Type = models.Type

type Import = models.Import

type ImportConfig = models.ImportConfig

type ImportTypeFilterCriteria = models.ImportTypeFilterCriteria

// ImportVariable views a content variable of an import type as a
// basecontentvariable.Descriptor. The shape is the shared model's, which carries the fields
// but no methods, and a method cannot be declared on an alias to another package's type -
// hence a wrapper rather than methods on ImportContentVariable.
func ImportVariable(variable *ImportContentVariable) basecontentvariable.Descriptor {
	return importVariable{variable}
}

type importVariable struct {
	*ImportContentVariable
}

func (this importVariable) GetName() string {
	return this.Name
}

func (this importVariable) GetCharacteristicId() string {
	return this.CharacteristicId
}

func (this importVariable) GetSubContentVariables() []basecontentvariable.Descriptor {
	ls := make([]basecontentvariable.Descriptor, len(this.SubContentVariables))
	for idx := range this.SubContentVariables {
		ls[idx] = importVariable{&this.SubContentVariables[idx]}
	}
	return ls
}

func (this importVariable) GetFunctionId() string {
	return this.FunctionId
}

// GetAspectIds returns the aspects this variable carries, with the deprecated AspectId folded
// into the list, the same way the import-repository folds it on write.
func (this importVariable) GetAspectIds() []string {
	return devicemodel.AspectIds(this.AspectId, this.AspectIds)
}

func (this importVariable) GetIsVoid() bool {
	return false
}
