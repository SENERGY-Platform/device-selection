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

package devicemodel

import (
	"slices"

	"github.com/SENERGY-Platform/models/go/models"
)

type FilterCriteria struct {
	Interaction   string   `json:"interaction,omitempty"`
	FunctionId    string   `json:"function_id"`
	AspectId      string   `json:"aspect_id"` //deprecated: alias for a single element AspectIds
	AspectIds     []string `json:"aspect_ids,omitempty"`
	DeviceClassId string   `json:"device_class_id"`
}

func (this FilterCriteria) Short() string {
	return this.FunctionId + "_" + models.AspectIdsShort(this.AspectId, this.AspectIds) + "_" + this.DeviceClassId
}

// GetAspectIds returns the aspects a criteria asks for, with the deprecated AspectId folded
// in. Everything that evaluates aspects reads the list, so the alias is resolved here instead
// of at every call site. Several aspects in one criteria are an AND on the content variable:
// the same variable has to carry all of them, each of them covering its own subtree.
func (this FilterCriteria) GetAspectIds() []string {
	return AspectIds(this.AspectId, this.AspectIds)
}

// AspectIds folds a deprecated single aspect id into an aspect id list. It is the free
// function behind FilterCriteria.GetAspectIds, for the criteria and content variables that
// carry the same pair of fields without being a FilterCriteria.
func AspectIds(aspectId string, aspectIds []string) []string {
	if aspectId == "" || slices.Contains(aspectIds, aspectId) {
		return aspectIds
	}
	return append(slices.Clone(aspectIds), aspectId)
}
