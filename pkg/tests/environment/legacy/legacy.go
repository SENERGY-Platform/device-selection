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

import (
	"context"
	"github.com/SENERGY-Platform/device-selection/pkg/api"
	"github.com/SENERGY-Platform/device-selection/pkg/configuration"
	"github.com/SENERGY-Platform/device-selection/pkg/controller"
	"github.com/SENERGY-Platform/device-selection/pkg/model/devicemodel"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/helper"
	"net/http/httptest"
	"sync"
	"time"
)

func Testenv(ctx context.Context, wg *sync.WaitGroup) (managerurl string, repourl string, searchurl string, selectionurl string, err error) {
	managerurl, repourl, searchurl, err = TestenvWithoutApi(ctx, wg)
	if err != nil {
		return managerurl, repourl, searchurl, selectionurl, err
	}

	c := &configuration.ConfigStruct{
		PermSearchUrl: searchurl,
		DeviceRepoUrl: repourl,
		Debug:         true,
	}

	ctrl, err := controller.New(ctx, c)
	if err != nil {
		return managerurl, repourl, searchurl, selectionurl, err
	}

	router := api.Router(c, ctrl)
	selectionApi := httptest.NewServer(router)
	wg.Add(1)
	go func() {
		<-ctx.Done()
		selectionApi.Close()
		wg.Done()
	}()
	selectionurl = selectionApi.URL

	time.Sleep(2 * time.Second)

	return
}

func TestenvWithoutApi(ctx context.Context, wg *sync.WaitGroup) (managerurl string, repourl string, searchurl string, err error) {
	deviceTypes := []devicemodel.DeviceType{
		{Id: "dt1", Name: "dt1name", DeviceClassId: "dc1", Services: []devicemodel.Service{
			testTechnicalService("11", "pid", nil, []devicemodel.Content{{
				Id:            "content1",
				Serialization: "json",
				ContentVariable: devicemodel.ContentVariable{
					Id:         "variable1",
					Name:       "variable1",
					AspectId:   "a1",
					FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "_1",
				},
			}}, devicemodel.REQUEST),
			testTechnicalService("11_b", "mqtt", nil, []devicemodel.Content{{
				Id:            "content2",
				Serialization: "json",
				ContentVariable: devicemodel.ContentVariable{
					Id:         "variable2",
					Name:       "variable2",
					AspectId:   "a1",
					FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "_1",
				},
			}}, devicemodel.EVENT),
			testTechnicalService("12", "pid", []devicemodel.Content{{
				Id:            "content3",
				Serialization: "json",
				ContentVariable: devicemodel.ContentVariable{
					Id:         "variable3",
					Name:       "variable3",
					AspectId:   "a1",
					FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "_1",
				},
			}}, nil, devicemodel.REQUEST),
		}},
		{Id: "dt2", Name: "dt2name", DeviceClassId: "dc1", Services: []devicemodel.Service{
			testTechnicalService("21", "pid", []devicemodel.Content{{
				Id:            "content4",
				Serialization: "json",
				ContentVariable: devicemodel.ContentVariable{
					Id:         "variable4",
					Name:       "variable4",
					AspectId:   "a1",
					FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "_1",
				},
			}}, nil, devicemodel.REQUEST),
			testTechnicalService("22", "pid", []devicemodel.Content{{
				Id:            "content5",
				Serialization: "json",
				ContentVariable: devicemodel.ContentVariable{
					Id:         "variable5",
					Name:       "variable5",
					AspectId:   "a1",
					FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "_1",
				},
			}}, nil, devicemodel.REQUEST),
		}},
		{Id: "dt3", Name: "dt1name", DeviceClassId: "dc1", Services: []devicemodel.Service{
			testService("31", "mqtt", devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION, devicemodel.EVENT),
			testService("32", "mqtt", devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION, devicemodel.EVENT),
		}},
		{Id: "dt4", Name: "dt2name", DeviceClassId: "dc1", Services: []devicemodel.Service{
			testService("41", "mqtt", devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION, devicemodel.EVENT),
			testService("42", "mqtt", devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION, devicemodel.EVENT),
		}},
	}
	deviceInstances := []devicemodel.Device{
		{
			Id:           "1",
			Name:         "1",
			DeviceTypeId: "dt1",
		},
		{
			Id:           "2",
			Name:         "2",
			DeviceTypeId: "dt2",
		},
		{
			Id:           "3",
			Name:         "3",
			DeviceTypeId: "dt3",
		},
		{
			Id:           "4",
			Name:         "4",
			DeviceTypeId: "dt4",
		},
	}

	concepts := []devicemodel.Concept{
		{
			Id: "concept",
		},
	}

	functions := []devicemodel.Function{
		{
			Id:        devicemodel.MEASURING_FUNCTION_PREFIX + "_1",
			ConceptId: "concept",
		},
		{
			Id:        devicemodel.CONTROLLING_FUNCTION_PREFIX + "_1",
			ConceptId: "concept",
		},
	}

	aspects := []devicemodel.Aspect{
		{
			Id:   "a1",
			Name: "a1",
		},
	}

	_, managerurl, repourl, searchurl, err = helper.EnvWithDevices(ctx, wg, deviceTypes, deviceInstances)
	if err != nil {
		return managerurl, repourl, searchurl, err
	}

	for _, concept := range concepts {
		err = helper.SetConcept(managerurl, concept)
		if err != nil {
			return managerurl, repourl, searchurl, err
		}
	}

	for _, f := range functions {
		err = helper.SetFunction(managerurl, f)
		if err != nil {
			return managerurl, repourl, searchurl, err
		}
	}

	for _, a := range aspects {
		err = helper.SetAspect(managerurl, a)
		if err != nil {
			return managerurl, repourl, searchurl, err
		}
	}

	time.Sleep(2 * time.Second)

	return
}

func testService(id string, protocolId string, functionType string, interaction devicemodel.Interaction) devicemodel.Service {
	result := devicemodel.Service{
		Id:          id,
		LocalId:     id + "_l",
		Name:        id + "_name",
		ProtocolId:  protocolId,
		Interaction: interaction,
	}
	if functionType == devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION {
		result.Outputs = append(result.Outputs, devicemodel.Content{
			ContentVariable: devicemodel.ContentVariable{
				FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "_1",
				AspectId:   "a1",
			},
			Serialization:     "json",
			ProtocolSegmentId: "ProtocolSegmentId",
		})
	} else {
		result.Inputs = append(result.Outputs, devicemodel.Content{
			ContentVariable: devicemodel.ContentVariable{
				FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "_1",
				AspectId:   "a1",
			},
			Serialization:     "json",
			ProtocolSegmentId: "ProtocolSegmentId",
		})
	}
	return result
}

func testTechnicalService(id string, protocolId string, inputs, outputs []devicemodel.Content, interaction devicemodel.Interaction) devicemodel.Service {
	return devicemodel.Service{
		Id:          id,
		LocalId:     id + "_l",
		Name:        id + "_name",
		ProtocolId:  protocolId,
		Outputs:     outputs,
		Inputs:      inputs,
		Interaction: interaction,
	}
}
