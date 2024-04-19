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
	"github.com/SENERGY-Platform/device-selection/pkg/model"
	"github.com/SENERGY-Platform/device-selection/pkg/model/devicemodel"
	"net/http"
)

func (this *Controller) getFilteredDeviceGroups(token string, descriptions model.FilterCriteriaAndSet, expectedInteraction devicemodel.Interaction) (result []model.Selectable, err error, code int) {
	groups := []model.DeviceGroup{}
	filter := []model.Selection{}
	for _, criteria := range descriptions {
		aspectIds := []string{}
		aspectIds = append(aspectIds, criteria.AspectId)
		if criteria.AspectId != "" {
			aspect, err := this.GetAspectNode(criteria.AspectId, token)
			if err != nil {
				return result, err, http.StatusInternalServerError
			}
			aspectIds = append(aspectIds, aspect.DescendentIds...)
		}
		or := []model.Selection{}
		for _, aspectId := range aspectIds {
			if expectedInteraction == devicemodel.EVENT_AND_REQUEST {
				groupCriteriaEvent := devicemodel.DeviceGroupFilterCriteria{
					Interaction:   devicemodel.EVENT,
					FunctionId:    criteria.FunctionId,
					AspectId:      aspectId,
					DeviceClassId: criteria.DeviceClassId,
				}
				groupCriteriaRequest := devicemodel.DeviceGroupFilterCriteria{
					Interaction:   devicemodel.REQUEST,
					FunctionId:    criteria.FunctionId,
					AspectId:      aspectId,
					DeviceClassId: criteria.DeviceClassId,
				}
				or = append(or, model.Selection{
					Condition: model.ConditionConfig{
						Feature:   "features.criteria_short",
						Operation: model.QueryEqualOperation,
						Value:     groupCriteriaEvent.Short(),
					},
				})
				or = append(or, model.Selection{
					Condition: model.ConditionConfig{
						Feature:   "features.criteria_short",
						Operation: model.QueryEqualOperation,
						Value:     groupCriteriaRequest.Short(),
					},
				})
			} else {
				groupCriteria := devicemodel.DeviceGroupFilterCriteria{
					Interaction:   expectedInteraction,
					FunctionId:    criteria.FunctionId,
					AspectId:      aspectId,
					DeviceClassId: criteria.DeviceClassId,
				}
				or = append(or, model.Selection{
					Condition: model.ConditionConfig{
						Feature:   "features.criteria_short",
						Operation: model.QueryEqualOperation,
						Value:     groupCriteria.Short(),
					},
				})
			}
		}
		filter = append(filter, model.Selection{Or: or})
	}

	var queryFilter *model.Selection
	if len(filter) > 0 {
		queryFilter = &model.Selection{
			And: filter,
		}
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
			Filter: queryFilter,
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

func (this *Controller) getFilteredDeviceGroupsV2(token string, descriptions model.FilterCriteriaAndSet) (result []model.Selectable, err error, code int) {
	groups := []model.DeviceGroup{}
	filter := []model.Selection{}
	for _, criteria := range descriptions {
		aspectIds := []string{}
		aspectIds = append(aspectIds, criteria.AspectId)
		if criteria.AspectId != "" {
			aspect, err := this.GetAspectNode(criteria.AspectId, token)
			if err != nil {
				return result, err, http.StatusInternalServerError
			}
			aspectIds = append(aspectIds, aspect.DescendentIds...)
		}
		or := []model.Selection{}
		for _, aspectId := range aspectIds {
			if devicemodel.Interaction(criteria.Interaction) == devicemodel.EVENT_AND_REQUEST || criteria.Interaction == "" {
				groupCriteriaEvent := devicemodel.DeviceGroupFilterCriteria{
					Interaction:   devicemodel.EVENT,
					FunctionId:    criteria.FunctionId,
					AspectId:      aspectId,
					DeviceClassId: criteria.DeviceClassId,
				}
				groupCriteriaRequest := devicemodel.DeviceGroupFilterCriteria{
					Interaction:   devicemodel.REQUEST,
					FunctionId:    criteria.FunctionId,
					AspectId:      aspectId,
					DeviceClassId: criteria.DeviceClassId,
				}
				or = append(or, model.Selection{
					Condition: model.ConditionConfig{
						Feature:   "features.criteria_short",
						Operation: model.QueryEqualOperation,
						Value:     groupCriteriaEvent.Short(),
					},
				})
				or = append(or, model.Selection{
					Condition: model.ConditionConfig{
						Feature:   "features.criteria_short",
						Operation: model.QueryEqualOperation,
						Value:     groupCriteriaRequest.Short(),
					},
				})
			} else {
				groupCriteria := devicemodel.DeviceGroupFilterCriteria{
					Interaction:   devicemodel.Interaction(criteria.Interaction),
					FunctionId:    criteria.FunctionId,
					AspectId:      aspectId,
					DeviceClassId: criteria.DeviceClassId,
				}
				or = append(or, model.Selection{
					Condition: model.ConditionConfig{
						Feature:   "features.criteria_short",
						Operation: model.QueryEqualOperation,
						Value:     groupCriteria.Short(),
					},
				})
			}
		}
		filter = append(filter, model.Selection{Or: or})
	}

	var queryFilter *model.Selection
	if len(filter) > 0 {
		queryFilter = &model.Selection{
			And: filter,
		}
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
			Filter: queryFilter,
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
