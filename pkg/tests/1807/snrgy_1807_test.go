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
	"errors"
	"github.com/SENERGY-Platform/device-selection/pkg/api"
	"github.com/SENERGY-Platform/device-selection/pkg/configuration"
	"github.com/SENERGY-Platform/device-selection/pkg/controller"
	"github.com/SENERGY-Platform/device-selection/pkg/model"
	"github.com/SENERGY-Platform/device-selection/pkg/model/devicemodel"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/environment"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/helper"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

func createTestEnv(ctx context.Context, wg *sync.WaitGroup, t *testing.T) (devicemanager string, config configuration.Config, err error) {
	config, err = configuration.Load("../../../config.json")
	if err != nil {
		return
	}
	_, devicemanager, config.DeviceRepoUrl, config.PermSearchUrl, _, config.ImportRepoUrl, config.ImportDeployUrl, err = environment.NewWithImport(ctx, wg)
	if err != nil {
		return
	}

	ctrl, err := controller.New(ctx, config)
	if err != nil {
		return
	}

	router := api.Router(config, ctrl)
	selectionApi := httptest.NewServer(router)
	wg.Add(1)
	go func() {
		<-ctx.Done()
		selectionApi.Close()
		wg.Done()
	}()
	var selectionUrl *url.URL
	selectionUrl, err = url.Parse(selectionApi.URL)
	if err != nil {
		return
	}
	config.ApiPort = selectionUrl.Port()
	return
}

func TestDeviceTypeMeasuringSelectables(t *testing.T) {
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

	t.Run("init metadata", createTestMetadata(devicemanager, interaction))

	t.Run("inside and outside temp", testSnrgy1807Selectables(conf, []devicemodel.FilterCriteria{
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature", AspectId: "inside_air"},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature", AspectId: "outside_air"},
	}, nil, []model.Selectable{
		{
			Device: snrgy1807Device("thermometer"),
			Services: []devicemodel.Service{
				{
					Id:          "getInsideTemperature",
					Interaction: interaction,
					Outputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:         "temperature",
								Name:       "temperature",
								FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
								AspectId:   "inside_air",
							},
						},
					},
				},
				{
					Id:          "getOutsideTemperature",
					Interaction: interaction,
					Outputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:         "temperature",
								Name:       "temperature",
								FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
								AspectId:   "outside_air",
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"getInsideTemperature": {
					{
						Path:             "temperature",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "inside_air",
							RootId:        "air",
							ParentId:      "air",
							ChildIds:      []string{},
							AncestorIds:   []string{"air"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
					},
				},
				"getOutsideTemperature": {
					{
						Path:             "temperature",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "outside_air",
							RootId:        "air",
							ParentId:      "air",
							ChildIds:      []string{"evening_outside_air", "morning_outside_air"},
							AncestorIds:   []string{"air"},
							DescendentIds: []string{"evening_outside_air", "morning_outside_air"},
						},
						FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
					},
				},
			},
		},
	}))

	t.Run("inside temp", testSnrgy1807Selectables(conf, []devicemodel.FilterCriteria{
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature", AspectId: "inside_air"},
	}, nil, []model.Selectable{
		{
			Device: snrgy1807Device("thermometer"),
			Services: []devicemodel.Service{
				{
					Id:          "getInsideTemperature",
					Interaction: interaction,
					Outputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:         "temperature",
								Name:       "temperature",
								FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
								AspectId:   "inside_air",
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"getInsideTemperature": {
					{
						Path:             "temperature",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "inside_air",
							RootId:        "air",
							ParentId:      "air",
							ChildIds:      []string{},
							AncestorIds:   []string{"air"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
					},
				},
			},
		},
		{
			Device: snrgy1807Device("thermostat"),
			Services: []devicemodel.Service{
				{
					Id:          "getTargetTemperature",
					Interaction: interaction,
					Outputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:         "temperature",
								Name:       "temperature",
								FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
								AspectId:   "inside_air",
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"getTargetTemperature": {
					{
						Path:             "temperature",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "inside_air",
							RootId:        "air",
							ParentId:      "air",
							ChildIds:      []string{},
							AncestorIds:   []string{"air"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
					},
				},
			},
		},
	}))

	t.Run("air temperature", testSnrgy1807Selectables(conf, []devicemodel.FilterCriteria{
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature", AspectId: "air"},
	}, nil, []model.Selectable{
		{
			Device: snrgy1807Device("simple_thermometer"),
			Services: []devicemodel.Service{
				{
					Id:          "getTemperature",
					Interaction: interaction,
					Outputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:         "temperature",
								Name:       "temperature",
								FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
								AspectId:   "air",
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"getTemperature": {
					{
						Path:             "temperature",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "air",
							RootId:        "air",
							ParentId:      "",
							ChildIds:      []string{"inside_air", "outside_air"},
							AncestorIds:   []string{},
							DescendentIds: []string{"evening_outside_air", "inside_air", "morning_outside_air", "outside_air"},
						},
						FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
					},
				},
			},
		},
		{
			Device: snrgy1807Device("thermometer"),
			Services: []devicemodel.Service{
				{
					Id:          "getInsideTemperature",
					Interaction: interaction,
					Outputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:         "temperature",
								Name:       "temperature",
								FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
								AspectId:   "inside_air",
							},
						},
					},
				},
				{
					Id:          "getOutsideTemperature",
					Interaction: interaction,
					Outputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:         "temperature",
								Name:       "temperature",
								FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
								AspectId:   "outside_air",
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"getInsideTemperature": {
					{
						Path:             "temperature",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "inside_air",
							RootId:        "air",
							ParentId:      "air",
							ChildIds:      []string{},
							AncestorIds:   []string{"air"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
					},
				},
				"getOutsideTemperature": {
					{
						Path:             "temperature",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "outside_air",
							RootId:        "air",
							ParentId:      "air",
							ChildIds:      []string{"evening_outside_air", "morning_outside_air"},
							AncestorIds:   []string{"air"},
							DescendentIds: []string{"evening_outside_air", "morning_outside_air"},
						},
						FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
					},
				},
			},
		},
		{
			Device: snrgy1807Device("thermostat"),
			Services: []devicemodel.Service{
				{
					Id:          "getTargetTemperature",
					Interaction: interaction,
					Outputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:         "temperature",
								Name:       "temperature",
								FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
								AspectId:   "inside_air",
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"getTargetTemperature": {
					{
						Path:             "temperature",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "inside_air",
							RootId:        "air",
							ParentId:      "air",
							ChildIds:      []string{},
							AncestorIds:   []string{"air"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
					},
				},
			},
		},
	}))

	t.Run("device temperature", testSnrgy1807Selectables(conf, []devicemodel.FilterCriteria{
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature", AspectId: "device"},
	}, nil, []model.Selectable{
		{
			Device: snrgy1807Device("pc_cooling_controller"),
			Services: []devicemodel.Service{
				{
					Id:          "getTemperatures",
					Interaction: interaction,
					Outputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:   "temperatures",
								Name: "temperatures",
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
						Path:             "temperatures.case",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "case",
							RootId:        "device",
							ParentId:      "device",
							ChildIds:      []string{},
							AncestorIds:   []string{"device"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
					},
					{
						Path:             "temperatures.cpu",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "cpu",
							RootId:        "device",
							ParentId:      "device",
							ChildIds:      []string{},
							AncestorIds:   []string{"device"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
					},
					{
						Path:             "temperatures.gpu",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "gpu",
							RootId:        "device",
							ParentId:      "device",
							ChildIds:      []string{},
							AncestorIds:   []string{"device"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
					},
				},
			},
		},
	}))

	t.Run("cpu temperature", testSnrgy1807Selectables(conf, []devicemodel.FilterCriteria{
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature", AspectId: "cpu"},
	}, nil, []model.Selectable{
		{
			Device: snrgy1807Device("pc_cooling_controller"),
			Services: []devicemodel.Service{
				{
					Id:          "getTemperatures",
					Interaction: interaction,
					Outputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:   "temperatures",
								Name: "temperatures",
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
						Path:             "temperatures.cpu",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "cpu",
							RootId:        "device",
							ParentId:      "device",
							ChildIds:      []string{},
							AncestorIds:   []string{"device"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
					},
				},
			},
		},
	}))

	t.Run("fan speed", testSnrgy1807Selectables(conf, []devicemodel.FilterCriteria{
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed", AspectId: "fan"},
	}, nil, []model.Selectable{
		{
			Device: snrgy1807Device("pc_cooling_controller"),
			Services: []devicemodel.Service{
				{
					Id:          "getFanSpeeds",
					Interaction: interaction,
					Outputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:   "speeds",
								Name: "speeds",
								SubContentVariables: []devicemodel.ContentVariable{
									{
										Id:         "cpu_fan",
										Name:       "cpu_fan",
										FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
										AspectId:   "cpu_fan",
									},
									{
										Id:         "gpu_fan",
										Name:       "gpu_fan",
										FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
										AspectId:   "gpu_fan",
									},
									{
										Id:         "case_fan_1",
										Name:       "case_fan_1",
										FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
										AspectId:   "case_fan_1",
									},
									{
										Id:         "case_fan_2",
										Name:       "case_fan_2",
										FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
										AspectId:   "case_fan_2",
									},
								},
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"getFanSpeeds": {
					{
						Path:             "speeds.case_fan_1",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "case_fan_1",
							RootId:        "fan",
							ParentId:      "case_fan",
							ChildIds:      []string{},
							AncestorIds:   []string{"case_fan", "fan"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
					},
					{
						Path:             "speeds.case_fan_2",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "case_fan_2",
							RootId:        "fan",
							ParentId:      "case_fan",
							ChildIds:      []string{},
							AncestorIds:   []string{"case_fan", "fan"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
					},
					{
						Path:             "speeds.cpu_fan",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "cpu_fan",
							RootId:        "fan",
							ParentId:      "fan",
							ChildIds:      []string{},
							AncestorIds:   []string{"fan"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
					},
					{
						Path:             "speeds.gpu_fan",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "gpu_fan",
							RootId:        "fan",
							ParentId:      "fan",
							ChildIds:      []string{},
							AncestorIds:   []string{"fan"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
					},
				},
			},
		},
	}))

	t.Run("cpu fan speed", testSnrgy1807Selectables(conf, []devicemodel.FilterCriteria{
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed", AspectId: "cpu_fan"},
	}, nil, []model.Selectable{
		{
			Device: snrgy1807Device("pc_cooling_controller"),
			Services: []devicemodel.Service{
				{
					Id:          "getFanSpeeds",
					Interaction: interaction,
					Outputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:   "speeds",
								Name: "speeds",
								SubContentVariables: []devicemodel.ContentVariable{
									{
										Id:         "cpu_fan",
										Name:       "cpu_fan",
										FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
										AspectId:   "cpu_fan",
									},
									{
										Id:         "gpu_fan",
										Name:       "gpu_fan",
										FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
										AspectId:   "gpu_fan",
									},
									{
										Id:         "case_fan_1",
										Name:       "case_fan_1",
										FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
										AspectId:   "case_fan_1",
									},
									{
										Id:         "case_fan_2",
										Name:       "case_fan_2",
										FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
										AspectId:   "case_fan_2",
									},
								},
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"getFanSpeeds": {
					{
						Path:             "speeds.cpu_fan",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "cpu_fan",
							RootId:        "fan",
							ParentId:      "fan",
							ChildIds:      []string{},
							AncestorIds:   []string{"fan"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
					},
				},
			},
		},
	}))

	t.Run("case fan speed", testSnrgy1807Selectables(conf, []devicemodel.FilterCriteria{
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed", AspectId: "case_fan"},
	}, nil, []model.Selectable{
		{
			Device: snrgy1807Device("pc_cooling_controller"),
			Services: []devicemodel.Service{
				{
					Id:          "getFanSpeeds",
					Interaction: interaction,
					Outputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:   "speeds",
								Name: "speeds",
								SubContentVariables: []devicemodel.ContentVariable{
									{
										Id:         "cpu_fan",
										Name:       "cpu_fan",
										FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
										AspectId:   "cpu_fan",
									},
									{
										Id:         "gpu_fan",
										Name:       "gpu_fan",
										FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
										AspectId:   "gpu_fan",
									},
									{
										Id:         "case_fan_1",
										Name:       "case_fan_1",
										FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
										AspectId:   "case_fan_1",
									},
									{
										Id:         "case_fan_2",
										Name:       "case_fan_2",
										FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
										AspectId:   "case_fan_2",
									},
								},
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"getFanSpeeds": {
					{
						Path:             "speeds.case_fan_1",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "case_fan_1",
							RootId:        "fan",
							ParentId:      "case_fan",
							ChildIds:      []string{},
							AncestorIds:   []string{"case_fan", "fan"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
					},
					{
						Path:             "speeds.case_fan_2",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "case_fan_2",
							RootId:        "fan",
							ParentId:      "case_fan",
							ChildIds:      []string{},
							AncestorIds:   []string{"case_fan", "fan"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
					},
				},
			},
		},
	}))

	t.Run("case fan speed", testSnrgy1807Selectables(conf, []devicemodel.FilterCriteria{
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed", AspectId: "case_fan_1"},
	}, nil, []model.Selectable{
		{
			Device: snrgy1807Device("pc_cooling_controller"),
			Services: []devicemodel.Service{
				{
					Id:          "getFanSpeeds",
					Interaction: interaction,
					Outputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:   "speeds",
								Name: "speeds",
								SubContentVariables: []devicemodel.ContentVariable{
									{
										Id:         "cpu_fan",
										Name:       "cpu_fan",
										FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
										AspectId:   "cpu_fan",
									},
									{
										Id:         "gpu_fan",
										Name:       "gpu_fan",
										FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
										AspectId:   "gpu_fan",
									},
									{
										Id:         "case_fan_1",
										Name:       "case_fan_1",
										FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
										AspectId:   "case_fan_1",
									},
									{
										Id:         "case_fan_2",
										Name:       "case_fan_2",
										FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
										AspectId:   "case_fan_2",
									},
								},
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"getFanSpeeds": {
					{
						Path:             "speeds.case_fan_1",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "case_fan_1",
							RootId:        "fan",
							ParentId:      "case_fan",
							ChildIds:      []string{},
							AncestorIds:   []string{"case_fan", "fan"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
					},
				},
			},
		},
	}))
}

func snrgy1807Device(s string) *model.PermSearchDevice {
	return &model.PermSearchDevice{
		Device: devicemodel.Device{
			Id:           s,
			LocalId:      s,
			Name:         s,
			DeviceTypeId: s,
		},
		Permissions: model.Permissions{
			R: true,
			W: true,
			X: true,
			A: true,
		},
		Shared:  false,
		Creator: helper.JwtSubject,
	}
}

func TestDeviceTypeControllingSelectables(t *testing.T) {
	//t.Skip("not implemented") //TODO

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

	t.Run("init metadata", createTestMetadata(devicemanager, interaction))

	t.Run("thermostat", testSnrgy1807Selectables(conf, []devicemodel.FilterCriteria{
		{FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature", DeviceClassId: "thermostat"},
	}, nil, []model.Selectable{
		{
			Device: snrgy1807Device("thermostat"),
			Services: []devicemodel.Service{{
				Id:          "setTargetTemperature",
				Interaction: interaction,
				Inputs: []devicemodel.Content{
					{
						ContentVariable: devicemodel.ContentVariable{
							Id:         "temperature",
							Name:       "temperature",
							FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
							AspectId:   "inside_air",
						},
					},
				},
			}},
			ServicePathOptions: map[string][]model.PathOption{
				"setTargetTemperature": {
					{
						Path:             "temperature",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "inside_air",
							RootId:        "air",
							ParentId:      "air",
							ChildIds:      []string{},
							AncestorIds:   []string{"air"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
					},
				},
			},
		},
		{
			Device: snrgy1807Device("thermostat_without_get"),
			Services: []devicemodel.Service{
				{
					Id:          "setTargetTemperature",
					Interaction: interaction,
					Inputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:         "temperature",
								Name:       "temperature",
								FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
								AspectId:   "inside_air",
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"setTargetTemperature": {
					{
						Path: "temperature",
						AspectNode: devicemodel.AspectNode{
							Id:            "inside_air",
							RootId:        "air",
							ParentId:      "air",
							ChildIds:      []string{},
							AncestorIds:   []string{"air"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
					},
				},
			},
		},
		{
			Device: snrgy1807Device("thermostat_without_get_base"),
			Services: []devicemodel.Service{
				{
					Id:          "setTargetTemperature",
					Interaction: interaction,
					Inputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:         "temperature",
								Name:       "temperature",
								FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
								AspectId:   "air",
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"setTargetTemperature": {
					{
						Path:             "temperature",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "air",
							RootId:        "air",
							ParentId:      "",
							ChildIds:      []string{"inside_air", "outside_air"},
							AncestorIds:   []string{},
							DescendentIds: []string{"evening_outside_air", "inside_air", "morning_outside_air", "outside_air"},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
					},
				},
			},
		},
		{
			Device: snrgy1807Device("thermostat_without_get_multiservice"),
			Services: []devicemodel.Service{
				{
					Id:          "setInsideTargetTemperature",
					Interaction: interaction,
					Inputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:         "temperature",
								Name:       "temperature",
								FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
								AspectId:   "inside_air",
							},
						},
					},
				},
				{
					Id:          "setOutsideTargetTemperature",
					Interaction: interaction,
					Inputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:         "temperature",
								Name:       "temperature",
								FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
								AspectId:   "outside_air",
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"setInsideTargetTemperature": {
					{
						Path:             "temperature",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "inside_air",
							RootId:        "air",
							ParentId:      "air",
							ChildIds:      []string{},
							AncestorIds:   []string{"air"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
					},
				},
				"setOutsideTargetTemperature": {
					{
						Path:             "temperature",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "outside_air",
							RootId:        "air",
							ParentId:      "air",
							ChildIds:      []string{"evening_outside_air", "morning_outside_air"},
							AncestorIds:   []string{"air"},
							DescendentIds: []string{"evening_outside_air", "morning_outside_air"},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
					},
				},
			},
		},
		{
			Device: snrgy1807Device("thermostat_without_get_multivalue"),
			Services: []devicemodel.Service{
				{
					Id:          "setTargetTemperature",
					Interaction: interaction,
					Inputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:   "temperature",
								Name: "temperature",
								SubContentVariables: []devicemodel.ContentVariable{
									{
										Id:         "inside",
										Name:       "inside",
										FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
										AspectId:   "inside_air",
									},
									{
										Id:         "outside",
										Name:       "outside",
										FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
										AspectId:   "outside_air",
									},
								},
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"setTargetTemperature": {
					{
						Path:             "temperature.inside",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "inside_air",
							RootId:        "air",
							ParentId:      "air",
							ChildIds:      []string{},
							AncestorIds:   []string{"air"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
					},
					{
						Path:             "temperature.outside",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "outside_air",
							RootId:        "air",
							ParentId:      "air",
							ChildIds:      []string{"evening_outside_air", "morning_outside_air"},
							AncestorIds:   []string{"air"},
							DescendentIds: []string{"evening_outside_air", "morning_outside_air"},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
					},
				},
			},
		},
		{
			Device: snrgy1807Device("thermostat_without_get_without_aspect"),
			Services: []devicemodel.Service{
				{
					Id:          "setTargetTemperature",
					Interaction: interaction,
					Inputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:         "temperature",
								Name:       "temperature",
								FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
								AspectId:   "",
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"setTargetTemperature": {
					{
						Path:             "temperature",
						CharacteristicId: "",
						AspectNode:       devicemodel.AspectNode{},
						FunctionId:       devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
					},
				},
			},
		},
	}))

	t.Run("thermostat air", testSnrgy1807Selectables(conf, []devicemodel.FilterCriteria{
		{FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature", DeviceClassId: "thermostat", AspectId: "air"},
	}, nil, []model.Selectable{
		{
			Device: snrgy1807Device("thermostat"),
			Services: []devicemodel.Service{{
				Id:          "setTargetTemperature",
				Interaction: interaction,
				Inputs: []devicemodel.Content{
					{
						ContentVariable: devicemodel.ContentVariable{
							Id:         "temperature",
							Name:       "temperature",
							FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
							AspectId:   "inside_air",
						},
					},
				},
			}},
			ServicePathOptions: map[string][]model.PathOption{
				"setTargetTemperature": {
					{
						Path:             "temperature",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "inside_air",
							RootId:        "air",
							ParentId:      "air",
							ChildIds:      []string{},
							AncestorIds:   []string{"air"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
					},
				},
			},
		},
		{
			Device: snrgy1807Device("thermostat_without_get"),
			Services: []devicemodel.Service{
				{
					Id:          "setTargetTemperature",
					Interaction: interaction,
					Inputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:         "temperature",
								Name:       "temperature",
								FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
								AspectId:   "inside_air",
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"setTargetTemperature": {
					{
						Path: "temperature",
						AspectNode: devicemodel.AspectNode{
							Id:            "inside_air",
							RootId:        "air",
							ParentId:      "air",
							ChildIds:      []string{},
							AncestorIds:   []string{"air"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
					},
				},
			},
		},
		{
			Device: snrgy1807Device("thermostat_without_get_base"),
			Services: []devicemodel.Service{
				{
					Id:          "setTargetTemperature",
					Interaction: interaction,
					Inputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:         "temperature",
								Name:       "temperature",
								FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
								AspectId:   "air",
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"setTargetTemperature": {
					{
						Path:             "temperature",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "air",
							RootId:        "air",
							ParentId:      "",
							ChildIds:      []string{"inside_air", "outside_air"},
							AncestorIds:   []string{},
							DescendentIds: []string{"evening_outside_air", "inside_air", "morning_outside_air", "outside_air"},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
					},
				},
			},
		},
		{
			Device: snrgy1807Device("thermostat_without_get_multiservice"),
			Services: []devicemodel.Service{
				{
					Id:          "setInsideTargetTemperature",
					Interaction: interaction,
					Inputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:         "temperature",
								Name:       "temperature",
								FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
								AspectId:   "inside_air",
							},
						},
					},
				},
				{
					Id:          "setOutsideTargetTemperature",
					Interaction: interaction,
					Inputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:         "temperature",
								Name:       "temperature",
								FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
								AspectId:   "outside_air",
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"setInsideTargetTemperature": {
					{
						Path:             "temperature",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "inside_air",
							RootId:        "air",
							ParentId:      "air",
							ChildIds:      []string{},
							AncestorIds:   []string{"air"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
					},
				},
				"setOutsideTargetTemperature": {
					{
						Path:             "temperature",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "outside_air",
							RootId:        "air",
							ParentId:      "air",
							ChildIds:      []string{"evening_outside_air", "morning_outside_air"},
							AncestorIds:   []string{"air"},
							DescendentIds: []string{"evening_outside_air", "morning_outside_air"},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
					},
				},
			},
		},
		{
			Device: snrgy1807Device("thermostat_without_get_multivalue"),
			Services: []devicemodel.Service{
				{
					Id:          "setTargetTemperature",
					Interaction: interaction,
					Inputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:   "temperature",
								Name: "temperature",
								SubContentVariables: []devicemodel.ContentVariable{
									{
										Id:         "inside",
										Name:       "inside",
										FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
										AspectId:   "inside_air",
									},
									{
										Id:         "outside",
										Name:       "outside",
										FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
										AspectId:   "outside_air",
									},
								},
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"setTargetTemperature": {
					{
						Path:             "temperature.inside",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "inside_air",
							RootId:        "air",
							ParentId:      "air",
							ChildIds:      []string{},
							AncestorIds:   []string{"air"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
					},
					{
						Path:             "temperature.outside",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "outside_air",
							RootId:        "air",
							ParentId:      "air",
							ChildIds:      []string{"evening_outside_air", "morning_outside_air"},
							AncestorIds:   []string{"air"},
							DescendentIds: []string{"evening_outside_air", "morning_outside_air"},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
					},
				},
			},
		},
	}))

	t.Run("thermostat inside air", testSnrgy1807Selectables(conf, []devicemodel.FilterCriteria{
		{FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature", DeviceClassId: "thermostat", AspectId: "inside_air"},
	}, nil, []model.Selectable{
		{
			Device: snrgy1807Device("thermostat"),
			Services: []devicemodel.Service{{
				Id:          "setTargetTemperature",
				Interaction: interaction,
				Inputs: []devicemodel.Content{
					{
						ContentVariable: devicemodel.ContentVariable{
							Id:         "temperature",
							Name:       "temperature",
							FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
							AspectId:   "inside_air",
						},
					},
				},
			}},
			ServicePathOptions: map[string][]model.PathOption{
				"setTargetTemperature": {
					{
						Path:             "temperature",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "inside_air",
							RootId:        "air",
							ParentId:      "air",
							ChildIds:      []string{},
							AncestorIds:   []string{"air"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
					},
				},
			},
		},
		{
			Device: snrgy1807Device("thermostat_without_get"),
			Services: []devicemodel.Service{
				{
					Id:          "setTargetTemperature",
					Interaction: interaction,
					Inputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:         "temperature",
								Name:       "temperature",
								FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
								AspectId:   "inside_air",
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"setTargetTemperature": {
					{
						Path: "temperature",
						AspectNode: devicemodel.AspectNode{
							Id:            "inside_air",
							RootId:        "air",
							ParentId:      "air",
							ChildIds:      []string{},
							AncestorIds:   []string{"air"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
					},
				},
			},
		},
		{
			Device: snrgy1807Device("thermostat_without_get_multiservice"),
			Services: []devicemodel.Service{
				{
					Id:          "setInsideTargetTemperature",
					Interaction: interaction,
					Inputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:         "temperature",
								Name:       "temperature",
								FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
								AspectId:   "inside_air",
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"setInsideTargetTemperature": {
					{
						Path:             "temperature",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "inside_air",
							RootId:        "air",
							ParentId:      "air",
							ChildIds:      []string{},
							AncestorIds:   []string{"air"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
					},
				},
			},
		},
		{
			Device: snrgy1807Device("thermostat_without_get_multivalue"),
			Services: []devicemodel.Service{
				{
					Id:          "setTargetTemperature",
					Interaction: interaction,
					Inputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:   "temperature",
								Name: "temperature",
								SubContentVariables: []devicemodel.ContentVariable{
									{
										Id:         "inside",
										Name:       "inside",
										FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
										AspectId:   "inside_air",
									},
									{
										Id:         "outside",
										Name:       "outside",
										FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
										AspectId:   "outside_air",
									},
								},
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"setTargetTemperature": {
					{
						Path:             "temperature.inside",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "inside_air",
							RootId:        "air",
							ParentId:      "air",
							ChildIds:      []string{},
							AncestorIds:   []string{"air"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
					},
				},
			},
		},
	}))

	t.Run("pc_cooling_controller fan_speed", testSnrgy1807Selectables(conf, []devicemodel.FilterCriteria{
		{FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed", DeviceClassId: "pc_cooling_controller"},
	}, nil, []model.Selectable{
		{
			Device: snrgy1807Device("pc_cooling_controller"),
			Services: []devicemodel.Service{
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
								Id:         "speed",
								Name:       "speed",
								FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
								AspectId:   "case_fan_2",
							},
						},
					},
				},
				{
					Id:          "setCpuSpeed",
					Interaction: interaction,
					Inputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:         "speed",
								Name:       "speed",
								FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
								AspectId:   "cpu_fan",
							},
						},
					},
				},
				{
					Id:          "setGpuSpeed",
					Interaction: interaction,
					Inputs: []devicemodel.Content{
						{
							ContentVariable: devicemodel.ContentVariable{
								Id:         "speed",
								Name:       "speed",
								FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
								AspectId:   "gpu_fan",
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"setCaseFan1Speed": {
					{
						Path:             "speed",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "case_fan_1",
							RootId:        "fan",
							ParentId:      "case_fan",
							ChildIds:      []string{},
							AncestorIds:   []string{"case_fan", "fan"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
					},
				},
				"setCaseFan2Speed": {
					{
						Path:             "speed",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "case_fan_2",
							RootId:        "fan",
							ParentId:      "case_fan",
							ChildIds:      []string{},
							AncestorIds:   []string{"case_fan", "fan"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
					},
				},
				"setCpuSpeed": {
					{
						Path:             "speed",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "cpu_fan",
							RootId:        "fan",
							ParentId:      "fan",
							ChildIds:      []string{},
							AncestorIds:   []string{"fan"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
					},
				},
				"setGpuSpeed": {
					{
						Path:             "speed",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "gpu_fan",
							RootId:        "fan",
							ParentId:      "fan",
							ChildIds:      []string{},
							AncestorIds:   []string{"fan"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
					},
				},
			},
		},
	}))

	t.Run("pc_cooling_controller fan_speed case_fan", testSnrgy1807Selectables(conf, []devicemodel.FilterCriteria{
		{FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed", DeviceClassId: "pc_cooling_controller", AspectId: "case_fan"},
	}, nil, []model.Selectable{
		{
			Device: snrgy1807Device("pc_cooling_controller"),
			Services: []devicemodel.Service{
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
								Id:         "speed",
								Name:       "speed",
								FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
								AspectId:   "case_fan_2",
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.PathOption{
				"setCaseFan1Speed": {
					{
						Path:             "speed",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "case_fan_1",
							RootId:        "fan",
							ParentId:      "case_fan",
							ChildIds:      []string{},
							AncestorIds:   []string{"case_fan", "fan"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
					},
				},
				"setCaseFan2Speed": {
					{
						Path:             "speed",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "case_fan_2",
							RootId:        "fan",
							ParentId:      "case_fan",
							ChildIds:      []string{},
							AncestorIds:   []string{"case_fan", "fan"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
					},
				},
			},
		},
	}))

	t.Run("pc_cooling_controller fan_speed case_fan_1", testSnrgy1807Selectables(conf, []devicemodel.FilterCriteria{
		{FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed", DeviceClassId: "pc_cooling_controller", AspectId: "case_fan_1"},
	}, nil, []model.Selectable{
		{
			Device: snrgy1807Device("pc_cooling_controller"),
			Services: []devicemodel.Service{
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
			},
			ServicePathOptions: map[string][]model.PathOption{
				"setCaseFan1Speed": {
					{
						Path:             "speed",
						CharacteristicId: "",
						AspectNode: devicemodel.AspectNode{
							Id:            "case_fan_1",
							RootId:        "fan",
							ParentId:      "case_fan",
							ChildIds:      []string{},
							AncestorIds:   []string{"case_fan", "fan"},
							DescendentIds: []string{},
						},
						FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
					},
				},
			},
		},
	}))
}

func testSnrgy1807Selectables(config configuration.Config, criteria []devicemodel.FilterCriteria, interactionsFilter []devicemodel.Interaction, expectedResult []model.Selectable) func(t *testing.T) {
	return func(t *testing.T) {
		result, err := GetSelectables(config, interactionsFilter, criteria)
		if err != nil {
			t.Error(err)
			return
		}
		result = sortServices(result)
		expectedResult = sortServices(expectedResult)
		normalizeTestSelectables(&result, true)
		normalizeTestSelectables(&expectedResult, true)
		if !reflect.DeepEqual(result, expectedResult) {
			resultJson, _ := json.Marshal(result)
			expectedJson, _ := json.Marshal(expectedResult)
			t.Error("\n", string(resultJson), "\n", string(expectedJson))
		}
	}
}

func sortServices(list []model.Selectable) (result []model.Selectable) {
	result = []model.Selectable{}
	for _, e := range list {
		sort.Slice(e.Services, func(i, j int) bool {
			return e.Services[i].Id < e.Services[j].Id
		})
		result = append(result, e)
	}
	return
}

func createTestMetadata(devicemanager string, interaction devicemodel.Interaction) func(t *testing.T) {
	return func(t *testing.T) {
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
		}
		devicetypes := []devicemodel.DeviceType{
			{
				Id:            "toggle",
				DeviceClassId: "toggle",
				Services: []devicemodel.Service{
					{
						Id:          "triggerToggle",
						Interaction: interaction,
						Inputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:         "void",
									IsVoid:     true,
									FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "toggle",
								},
							},
						},
					},
				},
			},
			{
				Id:            "thermostat_without_get",
				DeviceClassId: "thermostat",
				Services: []devicemodel.Service{
					{
						Id:          "setTargetTemperature",
						Interaction: interaction,
						Inputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
									AspectId:   "inside_air",
								},
							},
						},
					},
				},
			},
			{
				Id:            "thermostat_without_get_base",
				DeviceClassId: "thermostat",
				Services: []devicemodel.Service{
					{
						Id:          "setTargetTemperature",
						Interaction: interaction,
						Inputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
									AspectId:   "air",
								},
							},
						},
					},
				},
			},
			{
				Id:            "thermostat_without_get_without_aspect",
				DeviceClassId: "thermostat",
				Services: []devicemodel.Service{
					{
						Id:          "setTargetTemperature",
						Interaction: interaction,
						Inputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
									AspectId:   "",
								},
							},
						},
					},
				},
			},
			{
				Id:            "thermostat_without_get_multivalue",
				DeviceClassId: "thermostat",
				Services: []devicemodel.Service{
					{
						Id:          "setTargetTemperature",
						Interaction: interaction,
						Inputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:   "temperature",
									Name: "temperature",
									SubContentVariables: []devicemodel.ContentVariable{
										{
											Id:         "inside",
											Name:       "inside",
											FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
											AspectId:   "inside_air",
										},
										{
											Id:         "outside",
											Name:       "outside",
											FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
											AspectId:   "outside_air",
										},
									},
								},
							},
						},
					},
				},
			},
			{
				Id:            "thermostat_without_get_multiservice",
				DeviceClassId: "thermostat",
				Services: []devicemodel.Service{
					{
						Id:          "setInsideTargetTemperature",
						Interaction: interaction,
						Inputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
									AspectId:   "inside_air",
								},
							},
						},
					},
					{
						Id:          "setOutsideTargetTemperature",
						Interaction: interaction,
						Inputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
									AspectId:   "outside_air",
								},
							},
						},
					},
				},
			},
			{
				Id:            "thermostat",
				DeviceClassId: "thermostat",
				Services: []devicemodel.Service{
					{
						Id:          "setTargetTemperature",
						Interaction: interaction,
						Inputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
									AspectId:   "inside_air",
								},
							},
						},
					},
					{
						Id:          "getTargetTemperature",
						Interaction: interaction,
						Outputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
									AspectId:   "inside_air",
								},
							},
						},
					},
				},
			},
			{
				Id:            "thermometer",
				DeviceClassId: "thermometer",
				Services: []devicemodel.Service{
					{
						Id:          "getInsideTemperature",
						Interaction: interaction,
						Outputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
									AspectId:   "inside_air",
								},
							},
						},
					},
					{
						Id:          "getOutsideTemperature",
						Interaction: interaction,
						Outputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
									AspectId:   "outside_air",
								},
							},
						},
					},
				},
			},
			{
				Id:            "simple_thermometer",
				DeviceClassId: "thermometer",
				Services: []devicemodel.Service{
					{
						Id:          "getTemperature",
						Interaction: interaction,
						Outputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
									AspectId:   "air",
								},
							},
						},
					},
				},
			},
			{
				Id:            "water-probe",
				DeviceClassId: "thermometer",
				Services: []devicemodel.Service{
					{
						Id:          "getTemperature",
						Interaction: interaction,
						Outputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:               "temperature",
									Name:             "temperature",
									FunctionId:       devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature",
									AspectId:         "water",
									CharacteristicId: "water-probe-test-characteristic",
								},
							},
						},
					},
				},
			},
			{
				Id:            "pc_cooling_controller",
				DeviceClassId: "pc_cooling_controller",
				Services: []devicemodel.Service{
					{
						Id:          "getTemperatures",
						Interaction: interaction,
						Outputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:   "temperatures",
									Name: "temperatures",
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
						Id:          "getFanSpeeds",
						Interaction: interaction,
						Outputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:   "speeds",
									Name: "speeds",
									SubContentVariables: []devicemodel.ContentVariable{
										{
											Id:         "cpu_fan",
											Name:       "cpu_fan",
											FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
											AspectId:   "cpu_fan",
										},
										{
											Id:         "gpu_fan",
											Name:       "gpu_fan",
											FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
											AspectId:   "gpu_fan",
										},
										{
											Id:         "case_fan_1",
											Name:       "case_fan_1",
											FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
											AspectId:   "case_fan_1",
										},
										{
											Id:         "case_fan_2",
											Name:       "case_fan_2",
											FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
											AspectId:   "case_fan_2",
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
									Id:         "speed",
									Name:       "speed",
									FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
									AspectId:   "case_fan_2",
								},
							},
						},
					},
					{
						Id:          "setCpuSpeed",
						Interaction: interaction,
						Inputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:         "speed",
									Name:       "speed",
									FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
									AspectId:   "cpu_fan",
								},
							},
						},
					},
					{
						Id:          "setGpuSpeed",
						Interaction: interaction,
						Inputs: []devicemodel.Content{
							{
								ContentVariable: devicemodel.ContentVariable{
									Id:         "speed",
									Name:       "speed",
									FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
									AspectId:   "gpu_fan",
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

func GetSelectables(config configuration.Config, interactionsFilter []devicemodel.Interaction, descriptions model.FilterCriteriaAndSet) (result []model.Selectable, err error) {
	jsonStr, err := json.Marshal(descriptions)
	if err != nil {
		return result, err
	}
	interactionsQuery := ""
	if interactionsFilter != nil {
		interactions := []string{}
		for _, v := range interactionsFilter {
			interactions = append(interactions, string(v))
		}
		interactionsQuery = "&filter_interaction=" + url.QueryEscape(strings.Join(interactions, ","))
	}
	endpoint := "http://localhost:" + config.ApiPort + "/selectables?json=" + url.QueryEscape(string(jsonStr)) + interactionsQuery
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return result, err
	}
	req.Header.Set("Authorization", helper.AdminJwt)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, err
	}
	if resp.StatusCode != 200 {
		return result, errors.New("unexpected status code: " + resp.Status)
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return result, err
	}
	return result, err
}

func normalizeTestSelectables(selectables *[]model.Selectable, removeConfigurables bool) {
	for i, v := range *selectables {
		normalizeTestSelectable(&v, removeConfigurables)
		(*selectables)[i] = v
	}
	sort.SliceStable(*selectables, func(i, j int) bool {
		iName := ""
		if (*selectables)[i].Device != nil {
			iName = (*selectables)[i].Device.Name
		}
		if (*selectables)[i].DeviceGroup != nil {
			iName = (*selectables)[i].DeviceGroup.Name
		}

		jName := ""
		if (*selectables)[j].Device != nil {
			jName = (*selectables)[j].Device.Name
		}
		if (*selectables)[j].DeviceGroup != nil {
			jName = (*selectables)[j].DeviceGroup.Name
		}
		return iName < jName
	})
}

func normalizeTestSelectable(selectable *model.Selectable, removeConfigurables bool) {
	if selectable.Device != nil {
		selectable.Device.Id = ""
		selectable.Device.LocalId = ""
		selectable.Device.Creator = ""
		selectable.Device.Permissions = model.Permissions{}
		selectable.Device.Shared = false
		selectable.Device.DisplayName = ""
		if removeConfigurables {
			for sid, options := range selectable.ServicePathOptions {
				for i, option := range options {
					temp := option
					temp.Configurables = []devicemodel.Configurable{}
					options[i] = temp
				}
				selectable.ServicePathOptions[sid] = options
			}
		}

		for i, v := range selectable.Services {
			normalizeTestService(&v)
			selectable.Services[i] = v
		}
		sort.SliceStable(selectable.Services, func(i, j int) bool {
			iName := selectable.Services[i].Name
			jName := selectable.Services[j].Name
			return iName < jName
		})
	}
}

func normalizeTestService(service *devicemodel.Service) {
	for i, v := range service.Inputs {
		v.Id = ""
		normalizeTestContentVariable(&v.ContentVariable)
		service.Inputs[i] = v
	}
	for i, v := range service.Outputs {
		v.Id = ""
		normalizeTestContentVariable(&v.ContentVariable)
		service.Outputs[i] = v
	}
}

func normalizeTestContentVariable(variable *devicemodel.ContentVariable) {
	variable.Id = ""
	for i, v := range variable.SubContentVariables {
		normalizeTestContentVariable(&v)
		variable.SubContentVariables[i] = v
	}
}
