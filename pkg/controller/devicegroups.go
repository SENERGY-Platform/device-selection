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
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-selection/pkg/model"
	"github.com/SENERGY-Platform/device-selection/pkg/model/devicemodel"
	"github.com/SENERGY-Platform/models/go/models"
)

func (this *Controller) getFilteredDeviceGroups(token string, descriptions model.FilterCriteriaAndSet, expectedInteraction devicemodel.Interaction) (result []model.Selectable, err error, code int) {
	criteriaList := []client.FilterCriteria{}
	for _, c := range descriptions {
		criteria := client.FilterCriteria{
			FunctionId:    c.FunctionId,
			DeviceClassId: c.DeviceClassId,
			AspectId:      c.AspectId,
		}
		if expectedInteraction != "" {
			criteria.Interaction = expectedInteraction
		}
		criteriaList = append(criteriaList, criteria)
	}

	groups, _, err, code := this.devicerepo.ListDeviceGroups(token, client.DeviceGroupListOptions{
		Ids:             nil,
		Limit:           1000,
		SortBy:          "name.asc",
		Criteria:        criteriaList,
		Permission:      client.EXECUTE,
		IgnoreGenerated: false,
	})
	if err != nil {
		return result, err, code
	}
	for _, group := range groups {
		temp := group //prevent that every result element becomes the last element of groups
		result = append(result, model.Selectable{DeviceGroup: &model.DeviceGroup{
			Id:   temp.Id,
			Name: temp.Name,
		}})
	}
	return result, nil, 200
}

func (this *Controller) getFilteredDeviceGroupsV2(token string, descriptions model.FilterCriteriaAndSet) (result []model.Selectable, err error, code int) {
	criteriaList := []client.FilterCriteria{}
	for _, c := range descriptions {
		interaction := models.Interaction(c.Interaction)
		if interaction == models.EVENT_AND_REQUEST {
			interaction = ""
		}
		criteria := client.FilterCriteria{
			Interaction:   interaction,
			FunctionId:    c.FunctionId,
			DeviceClassId: c.DeviceClassId,
			AspectId:      c.AspectId,
		}
		criteriaList = append(criteriaList, criteria)
	}

	groups, _, err, code := this.devicerepo.ListDeviceGroups(token, client.DeviceGroupListOptions{
		Ids:             nil,
		Limit:           1000,
		SortBy:          "name.asc",
		Criteria:        criteriaList,
		Permission:      client.EXECUTE,
		IgnoreGenerated: false,
	})
	if err != nil {
		return result, err, code
	}
	for _, group := range groups {
		temp := group //prevent that every result element becomes the last element of groups
		result = append(result, model.Selectable{DeviceGroup: &model.DeviceGroup{
			Id:   temp.Id,
			Name: temp.Name,
		}})
	}
	return result, nil, 200

}
