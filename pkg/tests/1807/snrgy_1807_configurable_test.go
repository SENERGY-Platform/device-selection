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

package _807

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/device-selection/pkg/configuration"
	"github.com/SENERGY-Platform/device-selection/pkg/model"
	"github.com/SENERGY-Platform/device-selection/pkg/model/devicemodel"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/helper"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestConfigurables(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	devicemanager, conf, err := createTestEnv(ctx, wg, t)
	if err != nil {
		t.Error(err)
		return
	}

	interaction := devicemodel.EVENT_AND_REQUEST

	t.Run("init metadata", createTestConfigurableMetadata(devicemanager))

	getTemperaturesAverageTimeConfigurables := []devicemodel.Configurable{
		{
			Path:             "duration.sec",
			CharacteristicId: "",
			AspectNode:       devicemodel.AspectNode{},
			FunctionId:       "",
			Value:            30.0,
			Type:             devicemodel.Integer,
		},
		{
			Path:             "duration.ms",
			CharacteristicId: "ms",
			AspectNode:       devicemodel.AspectNode{},
			FunctionId:       "",
			Value:            32.0,
			Type:             devicemodel.Integer,
		},
	}

	t.Run("measuring temperature with config input", testSnrgy1807Configurable(conf, []devicemodel.FilterCriteria{
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature", AspectId: "device"},
	}, nil, []model.Selectable{
		{
			Device: snrgy1807Device("pc_cooling_controller"),
			Services: []devicemodel.Service{
				{
					Id:          "getTemperatures",
					Interaction: interaction,
					Inputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:   "duration",
								Name: "duration",
								Type: devicemodel.Structure,
								SubContentVariables: []devicemodel.ContentVariable{
									{
										Id:               "sec",
										Name:             "sec",
										Type:             devicemodel.Integer,
										CharacteristicId: "",
										Value:            30.0,
									},
									{
										Id:               "ms",
										Name:             "ms",
										Type:             devicemodel.Integer,
										CharacteristicId: "ms",
										Value:            32.0,
									},
								},
							},
						},
					},
					Outputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:   "avg_temperatures",
								Name: "avg_temperatures",
								SubContentVariables: []devicemodel.ContentVariable{
									{
										Id:         "cpu",
										Name:       "cpu",
										FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
										AspectId:   "cpu",
									},
									{
										Id:         "gpu",
										Name:       "gpu",
										FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
										AspectId:   "gpu",
									},
									{
										Id:         "case",
										Name:       "case",
										FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
										AspectId:   "case",
									},
								},
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"getTemperatures": {
					{
						Path:             "avg_temperatures.case",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "case",
							RootId:        "device",
							ParentId:      "device",
							ChildIds:      []string{},
							AncestorIds:   []string{"device"},
							DescendentIds: []string{},
						},
						FunctionId:    devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
						Configurables: getTemperaturesAverageTimeConfigurables,
					},
					{
						Path:             "avg_temperatures.cpu",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "cpu",
							RootId:        "device",
							ParentId:      "device",
							ChildIds:      []string{},
							AncestorIds:   []string{"device"},
							DescendentIds: []string{},
						},
						FunctionId:    devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
						Configurables: getTemperaturesAverageTimeConfigurables,
					},
					{
						Path:             "avg_temperatures.gpu",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "gpu",
							RootId:        "device",
							ParentId:      "device",
							ChildIds:      []string{},
							AncestorIds:   []string{"device"},
							DescendentIds: []string{},
						},
						FunctionId:    devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
						Configurables: getTemperaturesAverageTimeConfigurables,
					},
				},
			},
		},
	}))

}

func testSnrgy1807Configurable(config configuration.Config, criteria []devicemodel.FilterCriteria, interactionsFilter []devicemodel.Interaction, expectedResult []model.Selectable) func(t *testing.T) {
	return func(t *testing.T) {
		result, err := GetSelectables(config, interactionsFilter, criteria)
		if err != nil {
			t.Error(err)
			return
		}
		result = sortServices(result)
		expectedResult = sortServices(expectedResult)
		normalizeTestSelectables(&result, false)
		normalizeTestSelectables(&expectedResult, false)
		resultJson, _ := json.Marshal(result)
		expectedJson, _ := json.Marshal(expectedResult)
		if !reflect.DeepEqual(normalize(result), normalize(expectedResult)) {
			t.Error("\n", string(resultJson), "\n", string(expectedJson))
		} else {
			t.Log("\n", string(resultJson), "\n", string(expectedJson))
		}
	}
}

func createTestConfigurableMetadata(devicemanager string) func(t *testing.T) {
	return func(t *testing.T) {
		interaction := devicemodel.EVENT_AND_REQUEST
		aspects := []devicemodel.Aspect{
			{
				Id: "air",
				SubAspects: []devicemodel.Aspect{
					{Id: "inside_air"},
					{Id: "outside_air",
						SubAspects: []devicemodel.Aspect{
							{Id: "morning_outside_air"},
							{Id: "evening_outside_air"},
						},
					},
				},
			},
			{
				Id: "water",
			},
			{
				Id: "device",
				SubAspects: []devicemodel.Aspect{
					{Id: "cpu"},
					{Id: "gpu"},
					{Id: "case"},
				},
			},
			{
				Id: "fan",
				SubAspects: []devicemodel.Aspect{
					{Id: "cpu_fan"},
					{Id: "gpu_fan"},
					{Id: "case_fan",
						SubAspects: []devicemodel.Aspect{
							{Id: "case_fan_1"},
							{Id: "case_fan_2"},
							{Id: "case_fan_3"},
							{Id: "case_fan_4"},
							{Id: "case_fan_5"},
						},
					},
				},
			},
		}
		functions := []devicemodel.Function{
			{Id: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature"},
			{Id: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature"},
			{Id: devicemodel.MEASURING_FUNCTION_PREFIX + "getVolume"},
			{Id: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setVolume"},
			{Id: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed"},
			{Id: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed"},
			{Id: devicemodel.CONTROLLING_FUNCTION_PREFIX + "toggle"},
			{Id: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setMeasuringTime"},
		}
		devicetypes := []devicemodel.DeviceType{
			{
				Id:            "pc_cooling_controller",
				DeviceClassId: "pc_cooling_controller",
				Services: []devicemodel.Service{
					{
						Id:          "getTemperatures",
						Interaction: interaction,
						Inputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:   "duration",
									Name: "duration",
									Type: devicemodel.Structure,
									SubContentVariables: []devicemodel.ContentVariable{
										{
											Id:               "sec",
											Name:             "sec",
											Type:             devicemodel.Integer,
											CharacteristicId: "",
											Value:            30,
										},
										{
											Id:               "ms",
											Name:             "ms",
											Type:             devicemodel.Integer,
											CharacteristicId: "ms",
											Value:            32,
										},
									},
								},
							},
						},
						Outputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:   "avg_temperatures",
									Name: "avg_temperatures",
									SubContentVariables: []devicemodel.ContentVariable{
										{
											Id:         "cpu",
											Name:       "cpu",
											FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
											AspectId:   "cpu",
										},
										{
											Id:         "gpu",
											Name:       "gpu",
											FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
											AspectId:   "gpu",
										},
										{
											Id:         "case",
											Name:       "case",
											FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
											AspectId:   "case",
										},
									},
								},
							},
						},
					},

					{
						Id:          "getCaseFan1Speed",
						Interaction: interaction,
						Outputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:         "speed",
									Name:       "speed",
									FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
									AspectId:   "case_fan_1",
								},
							},
						},
					},
					{
						Id:          "getCaseFan2Speed",
						Interaction: interaction,
						Inputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:               "sec",
									Name:             "sec",
									Type:             devicemodel.Integer,
									CharacteristicId: "",
									Value:            24,
								},
							},
						},
						Outputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:         "speed",
									Name:       "speed",
									FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
									AspectId:   "case_fan_2",
								},
							},
						},
					},
					{
						Id:          "getCpuSpeed",
						Interaction: interaction,
						Inputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:               "sec",
									Name:             "sec",
									Type:             devicemodel.Integer,
									CharacteristicId: "sec",
									Value:            24,
								},
							},
						},
						Outputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:         "speed",
									Name:       "speed",
									FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
									AspectId:   "cpu_fan",
								},
							},
						},
					},
					{
						Id:          "getGpuSpeed",
						Interaction: interaction,
						Inputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:               "sec",
									Name:             "sec",
									Type:             devicemodel.Integer,
									CharacteristicId: "sec",
									FunctionId:       devicemodel.CONTROLLING_FUNCTION_PREFIX + "setMeasuringTime",
									Value:            24,
								},
							},
						},
						Outputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:         "speed",
									Name:       "speed",
									FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
									AspectId:   "gpu_fan",
								},
							},
						},
					},

					{
						Id:          "setCaseFanSpeed",
						Interaction: interaction,
						Inputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:   "speed",
									Name: "speed",
									SubContentVariables: []devicemodel.ContentVariable{
										{
											Id:         "1",
											Name:       "1",
											FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
											AspectId:   "case_fan_1",
											Value:      13,
										},
										{
											Id:         "2",
											Name:       "2",
											FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
											AspectId:   "case_fan_2",
											Value:      14,
										},
									},
								},
							},
						},
					},
					{
						Id:          "setCaseFan1Speed",
						Interaction: interaction,
						Inputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:         "speed",
									Name:       "speed",
									FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
									AspectId:   "case_fan_1",
								},
							},
						},
					},
					{
						Id:          "setCaseFan2Speed",
						Interaction: interaction,
						Inputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:    "header",
									Name:  "header",
									Type:  devicemodel.String,
									Value: "auth",
								},
							},
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:               "speed",
									Name:             "speed",
									FunctionId:       devicemodel.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
									AspectId:         "case_fan_2",
									CharacteristicId: "foo",
								},
							},
						},
					},
				},
			},
		}

		var err error

		for _, aspect := range aspects {
			err = helper.SetAspect(devicemanager, aspect)
			if err != nil {
				t.Error(err)
				return
			}
		}

		for _, function := range functions {
			err = helper.SetFunction(devicemanager, function)
			if err != nil {
				t.Error(err)
				return
			}
		}

		for _, dt := range devicetypes {
			err = helper.SetDeviceType(devicemanager, dt)
			if err != nil {
				t.Error(err)
				return
			}
			err = helper.SetDevice(devicemanager, devicemodel.Device{
				Id:           dt.Id,
				LocalId:      dt.Id,
				Name:         dt.Id,
				DeviceTypeId: dt.Id,
			})
			if err != nil {
				t.Error(err)
				return
			}
		}

		time.Sleep(5 * time.Second)
	}
}

func normalize(expected interface{}) (result interface{}) {
	temp, _ := json.Marshal(expected)
	json.Unmarshal(temp, &result)
	return
}
