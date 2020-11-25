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
	"device-selection/pkg/model/devicemodel"
)

func (this *Controller) getFilteredDeviceGroups(token string, descriptions model.FilterCriteriaAndSet, interaction devicemodel.Interaction) (result []model.Selectable, err error, code int) {
	groups := []model.DeviceGroup{}
	criteriaFilter := []model.Selection{{
		Condition: model.ConditionConfig{
			Feature:   "features.blocked_interaction",
			Operation: model.QueryEqualOperation,
			Value:     interaction,
		},
	}}
	for _, criteria := range descriptions {
		criteriaFilter = append(criteriaFilter, model.Selection{
			Condition: model.ConditionConfig{
				Feature:   "features.criteria_short",
				Operation: model.QueryEqualOperation,
				Value:     criteria.Short(),
			},
		})
	}
	err, code = this.Search(token, model.QueryMessage{
		Resource: "device-groups",
		Find: &model.QueryFind{
			QueryListCommons: model.QueryListCommons{
				Limit:    1000,
				Offset:   0,
				Rights:   "rx",
				SortBy:   "name",
				SortDesc: false,
			},
			Search: "",
			Filter: &model.Selection{
				And: criteriaFilter,
			},
		},
	}, &groups)
	if err != nil {
		return
	}
	for _, group := range groups {
		temp := group //prevent that every result element becomes the last element of groups
		result = append(result, model.Selectable{DeviceGroup: &temp})
	}
	return
}
