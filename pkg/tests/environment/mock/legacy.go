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

package mock

import (
	"device-selection/pkg/model/devicemodel"
	"log"
	"math/rand"
	"runtime/debug"
	"strconv"
	"strings"
)

type Service struct {
	Id          string                  `json:"id"`
	LocalId     string                  `json:"local_id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Interaction devicemodel.Interaction `json:"interaction"`
	AspectIds   []string                `json:"aspect_ids"`
	ProtocolId  string                  `json:"protocol_id"`
	FunctionIds []string                `json:"function_ids"`
	RdfType     string                  `json:"rdf_type"`
}

func FromLegacyServices(legacy []Service) (result []devicemodel.Service) {
	for _, s := range legacy {
		result = append(result, FromLegacyService(s))
	}
	return
}

func FromLegacyService(service Service) (result devicemodel.Service) {
	inputVariables := []devicemodel.ContentVariable{}
	outputVariables := []devicemodel.ContentVariable{}
	if len(service.FunctionIds) == 0 {
		log.Println("WARNING: empty function id list")
		debug.PrintStack()
	}
	if len(service.AspectIds) == 0 {
		log.Println("WARNING: empty aspect id list")
		debug.PrintStack()
	}
	for _, f := range service.FunctionIds {
		for _, a := range service.AspectIds {
			variable := devicemodel.ContentVariable{
				Id:               strconv.Itoa(rand.Int()),
				Name:             "foo",
				CharacteristicId: "",
				FunctionId:       f,
				AspectId:         a,
			}
			if strings.HasPrefix(f, devicemodel.MEASURING_FUNCTION_PREFIX) {
				outputVariables = append(outputVariables, variable)
			} else {
				inputVariables = append(inputVariables, variable)
			}
		}
	}
	return devicemodel.Service{
		Id:          service.Id,
		LocalId:     service.LocalId,
		Name:        service.Name,
		Description: service.Description,
		Interaction: service.Interaction,
		ProtocolId:  service.ProtocolId,
		Inputs: []devicemodel.Content{{ContentVariable: devicemodel.ContentVariable{
			SubContentVariables: inputVariables,
		}}},
		Outputs: []devicemodel.Content{{ContentVariable: devicemodel.ContentVariable{
			SubContentVariables: outputVariables,
		}}},
	}
}
