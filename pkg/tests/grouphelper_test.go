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

package tests

import (
	"context"
	"device-selection/pkg/model"
	"device-selection/pkg/model/devicemodel"
	"device-selection/pkg/tests/environment/legacy"
	"device-selection/pkg/tests/helper"
	"sync"
	"testing"
)

func TestGroupHelper(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deviceTypes := []devicemodel.DeviceType{
		{
			Id:            "temperature",
			Name:          "temperature",
			DeviceClassId: "temperature",
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "st1", Name: "st1", Interaction: devicemodel.REQUEST, AspectIds: []string{"air"}, FunctionIds: []string{"getTemperature"}},
			}),
		},
		{
			Id:            "lamp",
			Name:          "lamp",
			DeviceClassId: "lamp",
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "s1", Name: "s1", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "s2", Name: "s2", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
				{Id: "s3", Name: "s3", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getState"}},
			}),
		},
		{
			Id:            "event_lamp",
			Name:          "event_lamp",
			DeviceClassId: "lamp",
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "se1", Name: "se1", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "se2", Name: "se2", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
				{Id: "se3", Name: "se3", Interaction: devicemodel.EVENT, AspectIds: []string{"device", "light"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getState"}},
			}),
		},
		{
			Id:            "both_lamp",
			Name:          "both_lamp",
			DeviceClassId: "lamp",
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "se1", Name: "se1", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "se2", Name: "se2", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
				{Id: "se3", Name: "se3", Interaction: devicemodel.EVENT_AND_REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getState"}},
			}),
		},
		{
			Id:            "colorlamp",
			Name:          "colorlamp",
			DeviceClassId: "lamp",
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "s4", Name: "s4", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "s5", Name: "s5", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
				{Id: "s6", Name: "s6", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getState"}},
				{Id: "s7", Name: "s7", Interaction: devicemodel.REQUEST, AspectIds: []string{"light"}, FunctionIds: []string{"setColor"}},
				{Id: "s8", Name: "s8", Interaction: devicemodel.REQUEST, AspectIds: []string{"light"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getColor"}},
			}),
		},
		{
			Id:            "plug",
			Name:          "plug",
			DeviceClassId: "plug",
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "s9", Name: "s9", Interaction: devicemodel.REQUEST, AspectIds: []string{"device"}, FunctionIds: []string{"setOn"}},
				{Id: "s10", Name: "s10", Interaction: devicemodel.REQUEST, AspectIds: []string{"device"}, FunctionIds: []string{"setOff"}},
				{Id: "s11", Name: "s11", Interaction: devicemodel.REQUEST, AspectIds: []string{"device"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getState"}},
			}),
		},
	}

	devicesInstances := []devicemodel.Device{
		{
			Id:           "t1",
			Name:         "t1",
			DeviceTypeId: "temperature",
		},
		{
			Id:           "blamp",
			Name:         "blamp",
			DeviceTypeId: "both_lamp",
		},
		{
			Id:           "elamp",
			Name:         "elamp",
			DeviceTypeId: "event_lamp",
		},
		{
			Id:           "lamp1",
			Name:         "lamp1",
			DeviceTypeId: "lamp",
		},
		{
			Id:           "lamp2",
			Name:         "lamp2",
			DeviceTypeId: "lamp",
		},
		{
			Id:           "colorlamp1",
			Name:         "colorlamp1",
			DeviceTypeId: "colorlamp",
		},
		{
			Id:           "colorlamp2",
			Name:         "colorlamp2",
			DeviceTypeId: "colorlamp",
		},
		{
			Id:           "plug1",
			Name:         "plug1",
			DeviceTypeId: "plug",
		},
		{
			Id:           "plug2",
			Name:         "plug2",
			DeviceTypeId: "plug",
		},
	}

	_, _, _, selectionurl, err := helper.Grouphelpertestenv(ctx, wg, deviceTypes, devicesInstances)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("empty list", helper.GroupHelper(selectionurl, false, []string{}, model.DeviceGroupHelperResult{
		Criteria: []devicemodel.DeviceGroupFilterCriteria{},
		Options: []model.DeviceGroupOption{
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "lamp1",
					Name:         "lamp1",
					DeviceTypeId: "lamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "colorlamp1",
					Name:         "colorlamp1",
					DeviceTypeId: "colorlamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "plug1",
					Name:         "plug1",
					DeviceTypeId: "plug",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "t1",
					Name:         "t1",
					DeviceTypeId: "temperature",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
		},
	}))

	t.Run("blamp", helper.GroupHelper(selectionurl, false, []string{"blamp"}, model.DeviceGroupHelperResult{
		Criteria: []devicemodel.DeviceGroupFilterCriteria{
			{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
		},
		Options: []model.DeviceGroupOption{
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "lamp1",
					Name:         "lamp1",
					DeviceTypeId: "lamp",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "colorlamp1",
					Name:         "colorlamp1",
					DeviceTypeId: "colorlamp",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "plug1",
					Name:         "plug1",
					DeviceTypeId: "plug",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "t1",
					Name:         "t1",
					DeviceTypeId: "temperature",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
				},
				MaintainsGroupUsability: false,
			},
		},
	}))

	t.Run("lamp1", helper.GroupHelper(selectionurl, false, []string{"lamp1"}, model.DeviceGroupHelperResult{
		Criteria: []devicemodel.DeviceGroupFilterCriteria{
			{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
		},
		Options: []model.DeviceGroupOption{
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "colorlamp1",
					Name:         "colorlamp1",
					DeviceTypeId: "colorlamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "plug1",
					Name:         "plug1",
					DeviceTypeId: "plug",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "t1",
					Name:         "t1",
					DeviceTypeId: "temperature",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: false,
			},
		},
	}))

	t.Run("colorlamp1", helper.GroupHelper(selectionurl, false, []string{"colorlamp1"}, model.DeviceGroupHelperResult{
		Criteria: []devicemodel.DeviceGroupFilterCriteria{
			{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
		},
		Options: []model.DeviceGroupOption{
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "lamp1",
					Name:         "lamp1",
					DeviceTypeId: "lamp",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "plug1",
					Name:         "plug1",
					DeviceTypeId: "plug",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "t1",
					Name:         "t1",
					DeviceTypeId: "temperature",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: false,
			},
		},
	}))

	t.Run("lamp1 colorlamp1", helper.GroupHelper(selectionurl, false, []string{"lamp1", "colorlamp1"}, model.DeviceGroupHelperResult{
		Criteria: []devicemodel.DeviceGroupFilterCriteria{
			{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
		},
		Options: []model.DeviceGroupOption{
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "plug1",
					Name:         "plug1",
					DeviceTypeId: "plug",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "t1",
					Name:         "t1",
					DeviceTypeId: "temperature",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: false,
			},
		},
	}))

	t.Run("lamp1 colorlamp1 plug1", helper.GroupHelper(selectionurl, false, []string{"lamp1", "colorlamp1", "plug1"}, model.DeviceGroupHelperResult{
		Criteria: []devicemodel.DeviceGroupFilterCriteria{
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
		},
		Options: []model.DeviceGroupOption{
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: false,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "t1",
					Name:         "t1",
					DeviceTypeId: "temperature",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: false,
			},
		},
	}))

	t.Run("maintainUsability empty list", helper.GroupHelper(selectionurl, true, []string{}, model.DeviceGroupHelperResult{
		Criteria: []devicemodel.DeviceGroupFilterCriteria{},
		Options: []model.DeviceGroupOption{
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "lamp1",
					Name:         "lamp1",
					DeviceTypeId: "lamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "colorlamp1",
					Name:         "colorlamp1",
					DeviceTypeId: "colorlamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "plug1",
					Name:         "plug1",
					DeviceTypeId: "plug",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "t1",
					Name:         "t1",
					DeviceTypeId: "temperature",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
		},
	}))

	t.Run("maintainUsability blamp", helper.GroupHelper(selectionurl, true, []string{"blamp"}, model.DeviceGroupHelperResult{
		Criteria: []devicemodel.DeviceGroupFilterCriteria{
			{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
		},
		Options: []model.DeviceGroupOption{
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "lamp1",
					Name:         "lamp1",
					DeviceTypeId: "lamp",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "colorlamp1",
					Name:         "colorlamp1",
					DeviceTypeId: "colorlamp",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "plug1",
					Name:         "plug1",
					DeviceTypeId: "plug",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
				},
				MaintainsGroupUsability: true,
			},
		},
	}))

	t.Run("maintainUsability lamp1", helper.GroupHelper(selectionurl, true, []string{"lamp1"}, model.DeviceGroupHelperResult{
		Criteria: []devicemodel.DeviceGroupFilterCriteria{
			{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
		},
		Options: []model.DeviceGroupOption{
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "colorlamp1",
					Name:         "colorlamp1",
					DeviceTypeId: "colorlamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "plug1",
					Name:         "plug1",
					DeviceTypeId: "plug",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
		},
	}))

	t.Run("maintainUsability colorlamp1", helper.GroupHelper(selectionurl, true, []string{"colorlamp1"}, model.DeviceGroupHelperResult{
		Criteria: []devicemodel.DeviceGroupFilterCriteria{
			{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
		},
		Options: []model.DeviceGroupOption{
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "lamp1",
					Name:         "lamp1",
					DeviceTypeId: "lamp",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "plug1",
					Name:         "plug1",
					DeviceTypeId: "plug",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
		},
	}))

	t.Run("maintainUsability lamp1 colorlamp1", helper.GroupHelper(selectionurl, true, []string{"lamp1", "colorlamp1"}, model.DeviceGroupHelperResult{
		Criteria: []devicemodel.DeviceGroupFilterCriteria{
			{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
		},
		Options: []model.DeviceGroupOption{
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "plug1",
					Name:         "plug1",
					DeviceTypeId: "plug",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				}},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
		},
	}))

	t.Run("maintainUsability lamp1 colorlamp1 plug1", helper.GroupHelper(selectionurl, true, []string{"lamp1", "colorlamp1", "plug1"}, model.DeviceGroupHelperResult{
		Criteria: []devicemodel.DeviceGroupFilterCriteria{
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
		},
		Options: []model.DeviceGroupOption{
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: model.PermSearchDevice{Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				}},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
		},
	}))
}
