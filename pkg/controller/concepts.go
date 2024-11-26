/*
 * Copyright 2021 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/device-selection/pkg/model/devicemodel"
)

func (this *Controller) GetConcept(id string, token string) (c devicemodel.Concept, err error) {
	err = this.cache.Use(id, func() (interface{}, error) {
		result, err, _ := this.devicerepo.GetConceptWithoutCharacteristics(id)
		return result, err
	}, &c)
	return
}
