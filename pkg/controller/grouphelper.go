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
	"github.com/SENERGY-Platform/device-repository/v2/lib/client"
	"github.com/SENERGY-Platform/device-selection/pkg/model"
)

// DeviceGroupHelper is answered by the device-repository, which owns the device-group criteria:
// it derives them from its device-types and recomputes the stored criteria of a group whenever
// one of those is written. Deriving them a second time here would drift from that on every
// change to the rules - the aspect list of a content variable was the last such change.
func (this *Controller) DeviceGroupHelper(token string, deviceIds []string, search model.DeviceGroupHelperPagination, maintainGroupUsability bool, functionBlockList []string) (result model.DeviceGroupHelperResult, err error, code int) {
	return this.devicerepo.DeviceGroupHelper(token, deviceIds, client.DeviceGroupHelperOptions{
		Search:                  search.Search,
		Limit:                   search.Limit,
		Offset:                  search.Offset,
		MaintainsGroupUsability: maintainGroupUsability,
		FunctionBlockList:       functionBlockList,
	})
}
