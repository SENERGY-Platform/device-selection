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
	"strconv"
)

func (this *Controller) GetDeviceTypeSelectablesCached(token string, descriptions model.FilterCriteriaAndSet) (result []devicemodel.DeviceTypeSelectable, err error) {
	hash := hashCriteriaAndSet(descriptions)
	err = this.cache.Use("device-type-selectables."+hash, func() (interface{}, error) {
		return this.GetDeviceTypeSelectables(token, descriptions)
	}, &result)
	return
}

func (this *Controller) GetDeviceTypeSelectablesCachedV2(token string, descriptions model.FilterCriteriaAndSet, includeIdModified bool) (result []devicemodel.DeviceTypeSelectable, err error) {
	hash := hashCriteriaAndSet(descriptions)
	hash = hash + strconv.FormatBool(includeIdModified)
	err = this.cache.Use("device-type-selectables.v2."+hash, func() (interface{}, error) {
		return this.GetDeviceTypeSelectablesV2(token, descriptions, includeIdModified)
	}, &result)
	return
}

func (this *Controller) GetDeviceTypeSelectables(token string, descriptions model.FilterCriteriaAndSet) (result []devicemodel.DeviceTypeSelectable, err error) {
	criteria := []client.FilterCriteria{}
	for _, c := range descriptions {
		criteria = append(criteria, client.FilterCriteria{
			Interaction:   models.Interaction(c.Interaction),
			FunctionId:    c.FunctionId,
			DeviceClassId: c.DeviceClassId,
			AspectId:      c.AspectId,
		})
	}
	result, err, _ = this.devicerepo.GetDeviceTypeSelectables(criteria, "", nil, false)
	return result, err
}

func (this *Controller) GetDeviceTypeSelectablesV2(token string, descriptions model.FilterCriteriaAndSet, includeIdModified bool) (result []devicemodel.DeviceTypeSelectable, err error) {
	criteria := []client.FilterCriteria{}
	for _, c := range descriptions {
		criteria = append(criteria, client.FilterCriteria{
			Interaction:   models.Interaction(c.Interaction),
			FunctionId:    c.FunctionId,
			DeviceClassId: c.DeviceClassId,
			AspectId:      c.AspectId,
		})
	}
	result, err, _ = this.devicerepo.GetDeviceTypeSelectablesV2(criteria, "", includeIdModified, false)
	return result, err
}
