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

package devices

import (
	"bytes"
	"context"
	"device-selection/pkg/api"
	"device-selection/pkg/configuration"
	"device-selection/pkg/controller"
	"device-selection/pkg/model"
	"device-selection/pkg/model/devicemodel"
	"device-selection/pkg/tests/environment/docker"
	"device-selection/pkg/tests/environment/mock"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"runtime/debug"
	"sort"
	"sync"
	"testing"
	"time"
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
			Services: mock.FromLegacyServices([]mock.Service{
				{Id: "st1", Name: "st1", Interaction: devicemodel.REQUEST, AspectIds: []string{"air"}, FunctionIds: []string{"getTemperature"}},
			}),
		},
		{
			Id:            "lamp",
			Name:          "lamp",
			DeviceClassId: "lamp",
			Services: mock.FromLegacyServices([]mock.Service{
				{Id: "s1", Name: "s1", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "s2", Name: "s2", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
				{Id: "s3", Name: "s3", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getState"}},
			}),
		},
		{
			Id:            "event_lamp",
			Name:          "event_lamp",
			DeviceClassId: "lamp",
			Services: mock.FromLegacyServices([]mock.Service{
				{Id: "se1", Name: "se1", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "se2", Name: "se2", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
				{Id: "se3", Name: "se3", Interaction: devicemodel.EVENT, AspectIds: []string{"device", "light"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getState"}},
			}),
		},
		{
			Id:            "both_lamp",
			Name:          "both_lamp",
			DeviceClassId: "lamp",
			Services: mock.FromLegacyServices([]mock.Service{
				{Id: "se1", Name: "se1", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "se2", Name: "se2", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
				{Id: "se3", Name: "se3", Interaction: devicemodel.EVENT_AND_REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getState"}},
			}),
		},
		{
			Id:            "colorlamp",
			Name:          "colorlamp",
			DeviceClassId: "lamp",
			Services: mock.FromLegacyServices([]mock.Service{
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
			Services: mock.FromLegacyServices([]mock.Service{
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

	_, _, _, selectionurl, err := grouphelpertestenv(ctx, wg, deviceTypes, devicesInstances)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("empty list", testGroupHelper(selectionurl, false, []string{}, model.DeviceGroupHelperResult{
		Criteria: []devicemodel.DeviceGroupFilterCriteria{},
		Options: []model.DeviceGroupOption{
			{
				Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "lamp1",
					Name:         "lamp1",
					DeviceTypeId: "lamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "colorlamp1",
					Name:         "colorlamp1",
					DeviceTypeId: "colorlamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "plug1",
					Name:         "plug1",
					DeviceTypeId: "plug",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "t1",
					Name:         "t1",
					DeviceTypeId: "temperature",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
		},
	}))

	t.Run("blamp", testGroupHelper(selectionurl, false, []string{"blamp"}, model.DeviceGroupHelperResult{
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
				Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "lamp1",
					Name:         "lamp1",
					DeviceTypeId: "lamp",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "colorlamp1",
					Name:         "colorlamp1",
					DeviceTypeId: "colorlamp",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "plug1",
					Name:         "plug1",
					DeviceTypeId: "plug",
				},
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
				Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				},
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
				Device: devicemodel.Device{
					Id:           "t1",
					Name:         "t1",
					DeviceTypeId: "temperature",
				},
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

	t.Run("lamp1", testGroupHelper(selectionurl, false, []string{"lamp1"}, model.DeviceGroupHelperResult{
		Criteria: []devicemodel.DeviceGroupFilterCriteria{
			{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
		},
		Options: []model.DeviceGroupOption{
			{
				Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "colorlamp1",
					Name:         "colorlamp1",
					DeviceTypeId: "colorlamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "plug1",
					Name:         "plug1",
					DeviceTypeId: "plug",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "t1",
					Name:         "t1",
					DeviceTypeId: "temperature",
				},
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

	t.Run("colorlamp1", testGroupHelper(selectionurl, false, []string{"colorlamp1"}, model.DeviceGroupHelperResult{
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
				Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "lamp1",
					Name:         "lamp1",
					DeviceTypeId: "lamp",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "plug1",
					Name:         "plug1",
					DeviceTypeId: "plug",
				},
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
				Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				},
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
				Device: devicemodel.Device{
					Id:           "t1",
					Name:         "t1",
					DeviceTypeId: "temperature",
				},
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

	t.Run("lamp1 colorlamp1", testGroupHelper(selectionurl, false, []string{"lamp1", "colorlamp1"}, model.DeviceGroupHelperResult{
		Criteria: []devicemodel.DeviceGroupFilterCriteria{
			{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
		},
		Options: []model.DeviceGroupOption{
			{
				Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "plug1",
					Name:         "plug1",
					DeviceTypeId: "plug",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "t1",
					Name:         "t1",
					DeviceTypeId: "temperature",
				},
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

	t.Run("lamp1 colorlamp1 plug1", testGroupHelper(selectionurl, false, []string{"lamp1", "colorlamp1", "plug1"}, model.DeviceGroupHelperResult{
		Criteria: []devicemodel.DeviceGroupFilterCriteria{
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
		},
		Options: []model.DeviceGroupOption{
			{
				Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: false,
			},
			{
				Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "t1",
					Name:         "t1",
					DeviceTypeId: "temperature",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: false,
			},
		},
	}))

	t.Run("maintainUsability empty list", testGroupHelper(selectionurl, true, []string{}, model.DeviceGroupHelperResult{
		Criteria: []devicemodel.DeviceGroupFilterCriteria{},
		Options: []model.DeviceGroupOption{
			{
				Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "lamp1",
					Name:         "lamp1",
					DeviceTypeId: "lamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "colorlamp1",
					Name:         "colorlamp1",
					DeviceTypeId: "colorlamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "plug1",
					Name:         "plug1",
					DeviceTypeId: "plug",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "t1",
					Name:         "t1",
					DeviceTypeId: "temperature",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
		},
	}))

	t.Run("maintainUsability blamp", testGroupHelper(selectionurl, true, []string{"blamp"}, model.DeviceGroupHelperResult{
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
				Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "lamp1",
					Name:         "lamp1",
					DeviceTypeId: "lamp",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "colorlamp1",
					Name:         "colorlamp1",
					DeviceTypeId: "colorlamp",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "plug1",
					Name:         "plug1",
					DeviceTypeId: "plug",
				},
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
				Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				},
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

	t.Run("maintainUsability lamp1", testGroupHelper(selectionurl, true, []string{"lamp1"}, model.DeviceGroupHelperResult{
		Criteria: []devicemodel.DeviceGroupFilterCriteria{
			{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
		},
		Options: []model.DeviceGroupOption{
			{
				Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "colorlamp1",
					Name:         "colorlamp1",
					DeviceTypeId: "colorlamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "plug1",
					Name:         "plug1",
					DeviceTypeId: "plug",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
		},
	}))

	t.Run("maintainUsability colorlamp1", testGroupHelper(selectionurl, true, []string{"colorlamp1"}, model.DeviceGroupHelperResult{
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
				Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "lamp1",
					Name:         "lamp1",
					DeviceTypeId: "lamp",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "plug1",
					Name:         "plug1",
					DeviceTypeId: "plug",
				},
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
				Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				},
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

	t.Run("maintainUsability lamp1 colorlamp1", testGroupHelper(selectionurl, true, []string{"lamp1", "colorlamp1"}, model.DeviceGroupHelperResult{
		Criteria: []devicemodel.DeviceGroupFilterCriteria{
			{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
		},
		Options: []model.DeviceGroupOption{
			{
				Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "plug1",
					Name:         "plug1",
					DeviceTypeId: "plug",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				},
				RemovesCriteria: []devicemodel.DeviceGroupFilterCriteria{
					{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
					{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
				},
				MaintainsGroupUsability: true,
			},
		},
	}))

	t.Run("maintainUsability lamp1 colorlamp1 plug1", testGroupHelper(selectionurl, true, []string{"lamp1", "colorlamp1", "plug1"}, model.DeviceGroupHelperResult{
		Criteria: []devicemodel.DeviceGroupFilterCriteria{
			{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
		},
		Options: []model.DeviceGroupOption{
			{
				Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
			{
				Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
				},
				RemovesCriteria:         []devicemodel.DeviceGroupFilterCriteria{},
				MaintainsGroupUsability: true,
			},
		},
	}))
}

func testGroupHelper(selectionurl string, maintainUsability bool, deviceIds []string, expectedResult model.DeviceGroupHelperResult) func(t *testing.T) {
	return func(t *testing.T) {
		buff := new(bytes.Buffer)
		err := json.NewEncoder(buff).Encode(deviceIds)
		if err != nil {
			t.Error(err)
			return
		}
		query := ""
		if maintainUsability {
			query = "?maintains_group_usability=true"
		}
		req, err := http.NewRequest("POST", selectionurl+"/device-group-helper"+query, buff)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", adminjwt)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != 200 {
			temp, _ := io.ReadAll(resp.Body)
			t.Error(resp.StatusCode, string(temp))
			return
		}
		result := model.DeviceGroupHelperResult{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Error(err)
			return
		}
		result = normalizeGroupHelperResult(result)
		expectedResult = normalizeGroupHelperResult(expectedResult)
		if !reflect.DeepEqual(result, expectedResult) {
			resultJson, _ := json.Marshal(result)
			expectedJson, _ := json.Marshal(expectedResult)
			t.Error(string(resultJson), "\n", string(expectedJson))
		}
	}
}

func normalizeGroupHelperResult(result model.DeviceGroupHelperResult) model.DeviceGroupHelperResult {
	sort.SliceStable(result.Criteria, func(i, j int) bool {
		return result.Criteria[i].AspectId < result.Criteria[j].AspectId
	})
	sort.SliceStable(result.Criteria, func(i, j int) bool {
		return result.Criteria[i].FunctionId < result.Criteria[j].FunctionId
	})
	sort.SliceStable(result.Criteria, func(i, j int) bool {
		return result.Criteria[i].Interaction < result.Criteria[j].Interaction
	})
	sort.SliceStable(result.Options, func(i, j int) bool {
		return result.Options[i].Device.Id < result.Options[j].Device.Id
	})
	for i, option := range result.Options {
		option.Device.LocalId = option.Device.Id
		sort.SliceStable(option.RemovesCriteria, func(i, j int) bool {
			return option.RemovesCriteria[i].AspectId < option.RemovesCriteria[j].AspectId
		})
		sort.SliceStable(option.RemovesCriteria, func(i, j int) bool {
			return option.RemovesCriteria[i].FunctionId < option.RemovesCriteria[j].FunctionId
		})
		sort.SliceStable(option.RemovesCriteria, func(i, j int) bool {
			return option.RemovesCriteria[i].Interaction < option.RemovesCriteria[j].Interaction
		})
		result.Options[i] = option
	}
	return result
}

func grouphelpertestenv(ctx context.Context, wg *sync.WaitGroup, deviceTypes []devicemodel.DeviceType, deviceInstances []devicemodel.Device) (managerurl string, repourl string, searchurl string, selectionurl string, err error) {
	managerurl, repourl, searchurl, err = docker.DeviceManagerWithDependencies(ctx, wg)
	if err != nil {
		return managerurl, repourl, searchurl, selectionurl, err
	}

	for _, dt := range deviceTypes {
		err = testSetDeviceType(managerurl, dt)
		if err != nil {
			return managerurl, repourl, searchurl, selectionurl, err
		}
	}

	for _, d := range deviceInstances {
		err = testSetDevice(managerurl, d)
		if err != nil {
			return managerurl, repourl, searchurl, selectionurl, err
		}
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

	return
}

func testSetDeviceType(devicemanagerUrl string, dt devicemodel.DeviceType) error {
	resp, err := Jwtpost(adminjwt, devicemanagerUrl+"/device-types", dt)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		temp, _ := io.ReadAll(resp.Body)
		err = errors.New(string(temp))
		log.Println("ERROR:", err)
		debug.PrintStack()
		return err
	}
	return nil
}

func testSetDevice(devicemanagerUrl string, d devicemodel.Device) error {
	d.LocalId = d.Id
	resp, err := Jwtput(adminjwt, devicemanagerUrl+"/devices/"+url.PathEscape(d.Id), d)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		temp, _ := io.ReadAll(resp.Body)
		err = errors.New(string(temp))
		log.Println("ERROR:", err)
		debug.PrintStack()
		return err
	}
	return nil
}

var SleepAfterEdit = 2 * time.Second

func Jwtpost(token string, url string, msg interface{}) (resp *http.Response, err error) {
	body := new(bytes.Buffer)
	err = json.NewEncoder(body).Encode(msg)
	if err != nil {
		return resp, err
	}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if SleepAfterEdit != 0 {
		time.Sleep(SleepAfterEdit)
	}
	return
}

func Jwtput(token string, url string, msg interface{}) (resp *http.Response, err error) {
	body := new(bytes.Buffer)
	err = json.NewEncoder(body).Encode(msg)
	if err != nil {
		return resp, err
	}
	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if SleepAfterEdit != 0 {
		time.Sleep(SleepAfterEdit)
	}
	return
}

const userjwt = "Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICIzaUtabW9aUHpsMmRtQnBJdS1vSkY4ZVVUZHh4OUFIckVOcG5CcHM5SjYwIn0.eyJqdGkiOiJiOGUyNGZkNy1jNjJlLTRhNWQtOTQ4ZC1mZGI2ZWVkM2JmYzYiLCJleHAiOjE1MzA1MzIwMzIsIm5iZiI6MCwiaWF0IjoxNTMwNTI4NDMyLCJpc3MiOiJodHRwczovL2F1dGguc2VwbC5pbmZhaS5vcmcvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJkZDY5ZWEwZC1mNTUzLTQzMzYtODBmMy03ZjQ1NjdmODVjN2IiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiMjJlMGVjZjgtZjhhMS00NDQ1LWFmMjctNGQ1M2JmNWQxOGI5IiwiYXV0aF90aW1lIjoxNTMwNTI4NDIzLCJzZXNzaW9uX3N0YXRlIjoiMWQ3NWE5ODQtNzM1OS00MWJlLTgxYjktNzMyZDgyNzRjMjNlIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJjcmVhdGUtcmVhbG0iLCJhZG1pbiIsImRldmVsb3BlciIsInVtYV9hdXRob3JpemF0aW9uIiwidXNlciJdfSwicmVzb3VyY2VfYWNjZXNzIjp7Im1hc3Rlci1yZWFsbSI6eyJyb2xlcyI6WyJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsInZpZXctcmVhbG0iLCJtYW5hZ2UtaWRlbnRpdHktcHJvdmlkZXJzIiwiaW1wZXJzb25hdGlvbiIsImNyZWF0ZS1jbGllbnQiLCJtYW5hZ2UtdXNlcnMiLCJxdWVyeS1yZWFsbXMiLCJ2aWV3LWF1dGhvcml6YXRpb24iLCJxdWVyeS1jbGllbnRzIiwicXVlcnktdXNlcnMiLCJtYW5hZ2UtZXZlbnRzIiwibWFuYWdlLXJlYWxtIiwidmlldy1ldmVudHMiLCJ2aWV3LXVzZXJzIiwidmlldy1jbGllbnRzIiwibWFuYWdlLWF1dGhvcml6YXRpb24iLCJtYW5hZ2UtY2xpZW50cyIsInF1ZXJ5LWdyb3VwcyJdfSwiYWNjb3VudCI6eyJyb2xlcyI6WyJtYW5hZ2UtYWNjb3VudCIsIm1hbmFnZS1hY2NvdW50LWxpbmtzIiwidmlldy1wcm9maWxlIl19fSwicm9sZXMiOlsidW1hX2F1dGhvcml6YXRpb24iLCJhZG1pbiIsImNyZWF0ZS1yZWFsbSIsImRldmVsb3BlciIsInVzZXIiLCJvZmZsaW5lX2FjY2VzcyJdLCJuYW1lIjoiZGYgZGZmZmYiLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJzZXBsIiwiZ2l2ZW5fbmFtZSI6ImRmIiwiZmFtaWx5X25hbWUiOiJkZmZmZiIsImVtYWlsIjoic2VwbEBzZXBsLmRlIn0.eOwKV7vwRrWr8GlfCPFSq5WwR_p-_rSJURXCV1K7ClBY5jqKQkCsRL2V4YhkP1uS6ECeSxF7NNOLmElVLeFyAkvgSNOUkiuIWQpMTakNKynyRfH0SrdnPSTwK2V1s1i4VjoYdyZWXKNjeT2tUUX9eCyI5qOf_Dzcai5FhGCSUeKpV0ScUj5lKrn56aamlW9IdmbFJ4VwpQg2Y843Vc0TqpjK9n_uKwuRcQd9jkKHkbwWQ-wyJEbFWXHjQ6LnM84H0CQ2fgBqPPfpQDKjGSUNaCS-jtBcbsBAWQSICwol95BuOAqVFMucx56Wm-OyQOuoQ1jaLt2t-Uxtr-C9wKJWHQ"
const userid = "dd69ea0d-f553-4336-80f3-7f4567f85c7b"
const adminjwt = "Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICIzaUtabW9aUHpsMmRtQnBJdS1vSkY4ZVVUZHh4OUFIckVOcG5CcHM5SjYwIn0.eyJqdGkiOiJiOGUyNGZkNy1jNjJlLTRhNWQtOTQ4ZC1mZGI2ZWVkM2JmYzYiLCJleHAiOjE1MzA1MzIwMzIsIm5iZiI6MCwiaWF0IjoxNTMwNTI4NDMyLCJpc3MiOiJodHRwczovL2F1dGguc2VwbC5pbmZhaS5vcmcvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJkZDY5ZWEwZC1mNTUzLTQzMzYtODBmMy03ZjQ1NjdmODVjN2IiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiMjJlMGVjZjgtZjhhMS00NDQ1LWFmMjctNGQ1M2JmNWQxOGI5IiwiYXV0aF90aW1lIjoxNTMwNTI4NDIzLCJzZXNzaW9uX3N0YXRlIjoiMWQ3NWE5ODQtNzM1OS00MWJlLTgxYjktNzMyZDgyNzRjMjNlIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJjcmVhdGUtcmVhbG0iLCJhZG1pbiIsImRldmVsb3BlciIsInVtYV9hdXRob3JpemF0aW9uIiwidXNlciJdfSwicmVzb3VyY2VfYWNjZXNzIjp7Im1hc3Rlci1yZWFsbSI6eyJyb2xlcyI6WyJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsInZpZXctcmVhbG0iLCJtYW5hZ2UtaWRlbnRpdHktcHJvdmlkZXJzIiwiaW1wZXJzb25hdGlvbiIsImNyZWF0ZS1jbGllbnQiLCJtYW5hZ2UtdXNlcnMiLCJxdWVyeS1yZWFsbXMiLCJ2aWV3LWF1dGhvcml6YXRpb24iLCJxdWVyeS1jbGllbnRzIiwicXVlcnktdXNlcnMiLCJtYW5hZ2UtZXZlbnRzIiwibWFuYWdlLXJlYWxtIiwidmlldy1ldmVudHMiLCJ2aWV3LXVzZXJzIiwidmlldy1jbGllbnRzIiwibWFuYWdlLWF1dGhvcml6YXRpb24iLCJtYW5hZ2UtY2xpZW50cyIsInF1ZXJ5LWdyb3VwcyJdfSwiYWNjb3VudCI6eyJyb2xlcyI6WyJtYW5hZ2UtYWNjb3VudCIsIm1hbmFnZS1hY2NvdW50LWxpbmtzIiwidmlldy1wcm9maWxlIl19fSwicm9sZXMiOlsidW1hX2F1dGhvcml6YXRpb24iLCJhZG1pbiIsImNyZWF0ZS1yZWFsbSIsImRldmVsb3BlciIsInVzZXIiLCJvZmZsaW5lX2FjY2VzcyJdLCJuYW1lIjoiZGYgZGZmZmYiLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJzZXBsIiwiZ2l2ZW5fbmFtZSI6ImRmIiwiZmFtaWx5X25hbWUiOiJkZmZmZiIsImVtYWlsIjoic2VwbEBzZXBsLmRlIn0.eOwKV7vwRrWr8GlfCPFSq5WwR_p-_rSJURXCV1K7ClBY5jqKQkCsRL2V4YhkP1uS6ECeSxF7NNOLmElVLeFyAkvgSNOUkiuIWQpMTakNKynyRfH0SrdnPSTwK2V1s1i4VjoYdyZWXKNjeT2tUUX9eCyI5qOf_Dzcai5FhGCSUeKpV0ScUj5lKrn56aamlW9IdmbFJ4VwpQg2Y843Vc0TqpjK9n_uKwuRcQd9jkKHkbwWQ-wyJEbFWXHjQ6LnM84H0CQ2fgBqPPfpQDKjGSUNaCS-jtBcbsBAWQSICwol95BuOAqVFMucx56Wm-OyQOuoQ1jaLt2t-Uxtr-C9wKJWHQ"
