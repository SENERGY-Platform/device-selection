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
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"runtime/debug"
	"sort"
	"strings"
	"testing"
)

func TestGroupHelper(t *testing.T) {
	deviceTypes := []devicemodel.DeviceType{
		{
			Id:            "temperature",
			Name:          "temperature",
			DeviceClassId: "temperature",
			Services: []devicemodel.Service{
				{Id: "st1", Name: "st1", Interaction: devicemodel.REQUEST, AspectIds: []string{"air"}, FunctionIds: []string{"getTemperature"}},
			},
		},
		{
			Id:            "lamp",
			Name:          "lamp",
			DeviceClassId: "lamp",
			Services: []devicemodel.Service{
				{Id: "s1", Name: "s1", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "s2", Name: "s2", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
				{Id: "s3", Name: "s3", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getState"}},
			},
		},
		{
			Id:            "event_lamp",
			Name:          "event_lamp",
			DeviceClassId: "lamp",
			Services: []devicemodel.Service{
				{Id: "se1", Name: "se1", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "se2", Name: "se2", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
				{Id: "se3", Name: "se3", Interaction: devicemodel.EVENT, AspectIds: []string{"device", "light"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getState"}},
			},
		},
		{
			Id:            "both_lamp",
			Name:          "both_lamp",
			DeviceClassId: "lamp",
			Services: []devicemodel.Service{
				{Id: "se1", Name: "se1", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "se2", Name: "se2", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
				{Id: "se3", Name: "se3", Interaction: devicemodel.EVENT_AND_REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getState"}},
			},
		},
		{
			Id:            "colorlamp",
			Name:          "colorlamp",
			DeviceClassId: "lamp",
			Services: []devicemodel.Service{
				{Id: "s4", Name: "s4", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "s5", Name: "s5", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
				{Id: "s6", Name: "s6", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getState"}},
				{Id: "s7", Name: "s7", Interaction: devicemodel.REQUEST, AspectIds: []string{"light"}, FunctionIds: []string{"setColor"}},
				{Id: "s8", Name: "s8", Interaction: devicemodel.REQUEST, AspectIds: []string{"light"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getColor"}},
			},
		},
		{
			Id:            "plug",
			Name:          "plug",
			DeviceClassId: "plug",
			Services: []devicemodel.Service{
				{Id: "s9", Name: "s9", Interaction: devicemodel.REQUEST, AspectIds: []string{"device"}, FunctionIds: []string{"setOn"}},
				{Id: "s10", Name: "s10", Interaction: devicemodel.REQUEST, AspectIds: []string{"device"}, FunctionIds: []string{"setOff"}},
				{Id: "s11", Name: "s11", Interaction: devicemodel.REQUEST, AspectIds: []string{"device"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getState"}},
			},
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

	semanticmock, searchmock, devicerepomock, selectionApi, err := grouphelpertestenv(deviceTypes, devicesInstances)
	if err != nil {
		t.Error(err)
		return
	}
	defer selectionApi.Close()
	defer searchmock.Close()
	defer devicerepomock.Close()
	defer semanticmock.Close()

	t.Run("empty list", testGroupHelper(selectionApi, false, []string{}, model.DeviceGroupHelperResult{
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

	t.Run("blamp", testGroupHelper(selectionApi, false, []string{"blamp"}, model.DeviceGroupHelperResult{
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

	t.Run("lamp1", testGroupHelper(selectionApi, false, []string{"lamp1"}, model.DeviceGroupHelperResult{
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

	t.Run("colorlamp1", testGroupHelper(selectionApi, false, []string{"colorlamp1"}, model.DeviceGroupHelperResult{
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

	t.Run("lamp1 colorlamp1", testGroupHelper(selectionApi, false, []string{"lamp1", "colorlamp1"}, model.DeviceGroupHelperResult{
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

	t.Run("lamp1 colorlamp1 plug1", testGroupHelper(selectionApi, false, []string{"lamp1", "colorlamp1", "plug1"}, model.DeviceGroupHelperResult{
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

	t.Run("maintainUsability empty list", testGroupHelper(selectionApi, true, []string{}, model.DeviceGroupHelperResult{
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

	t.Run("maintainUsability blamp", testGroupHelper(selectionApi, true, []string{"blamp"}, model.DeviceGroupHelperResult{
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

	t.Run("maintainUsability lamp1", testGroupHelper(selectionApi, true, []string{"lamp1"}, model.DeviceGroupHelperResult{
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

	t.Run("maintainUsability colorlamp1", testGroupHelper(selectionApi, true, []string{"colorlamp1"}, model.DeviceGroupHelperResult{
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

	t.Run("maintainUsability lamp1 colorlamp1", testGroupHelper(selectionApi, true, []string{"lamp1", "colorlamp1"}, model.DeviceGroupHelperResult{
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

	t.Run("maintainUsability lamp1 colorlamp1 plug1", testGroupHelper(selectionApi, true, []string{"lamp1", "colorlamp1", "plug1"}, model.DeviceGroupHelperResult{
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

func testGroupHelper(selectionApi *httptest.Server, maintainUsability bool, deviceIds []string, expectedResult model.DeviceGroupHelperResult) func(t *testing.T) {
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
		req, err := http.NewRequest("POST", selectionApi.URL+"/device-group-helper"+query, buff)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", "test-token")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != 200 {
			t.Error(resp.StatusCode)
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

func grouphelpertestenv(deviceTypes []devicemodel.DeviceType, deviceInstances []devicemodel.Device) (semanticRepoMock *httptest.Server, searchmock *httptest.Server, devicerepomock *httptest.Server, selectionApi *httptest.Server, err error) {

	devicerepomock = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, dt := range deviceTypes {
			if r.URL.Path == "/device-types/"+url.PathEscape(dt.Id) {
				json.NewEncoder(w).Encode(dt)
				return
			}
		}
		for _, d := range deviceInstances {
			if r.URL.Path == "/devices/"+url.PathEscape(d.Id) {
				json.NewEncoder(w).Encode(d)
				return
			}
		}
		log.Println("DEBUG: devicerepo call: " + r.URL.Path + "?" + r.URL.RawQuery)
		http.Error(w, "not implemented", http.StatusNotImplemented)
	}))

	semanticRepoMock = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(deviceTypes)
	}))

	searchmock = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/jwt/select/devices/device_type_id/") {
			result := []TestPermSearchDevice{}
			for _, d := range deviceInstances {
				if r.URL.Path == "/jwt/select/devices/device_type_id/"+d.DeviceTypeId+"/x" {
					result = append(result, TestPermSearchDevice{Id: d.Id, Name: d.Name, DeviceType: d.DeviceTypeId, Permissions: model.Permissions{R: true, W: true, X: true, A: true}})
				}
			}
			json.NewEncoder(w).Encode(result)
			return
		}
		if r.URL.Path == "/v2/query" {
			query := model.QueryMessage{}
			err := json.NewDecoder(r.Body).Decode(&query)
			if err != nil {
				debug.PrintStack()
				http.Error(w, "not implemented", http.StatusNotImplemented)
				return
			}
			if query.Resource != "devices" ||
				query.Find == nil ||
				query.Find.Filter == nil ||
				query.Find.Filter.And == nil ||
				len(query.Find.Filter.And) == 0 ||
				query.Find.Filter.And[0].Not == nil ||
				query.Find.Filter.And[0].Not.Condition.Feature != "id" ||
				query.Find.Filter.And[0].Not.Condition.Operation != model.QueryAnyValueInFeatureOperation {
				debug.PrintStack()
				http.Error(w, "not implemented", http.StatusNotImplemented)
				return
			}
			notIdList, ok := query.Find.Filter.And[0].Not.Condition.Value.([]interface{})
			if !ok {
				debug.PrintStack()
				http.Error(w, "not implemented", http.StatusNotImplemented)
				return
			}

			var filterByDeviceTypeId []interface{}
			if len(query.Find.Filter.And) >= 2 {
				if query.Find.Filter.And[1].Condition.Feature != "features.device_type_id" ||
					query.Find.Filter.And[1].Condition.Operation != model.QueryAnyValueInFeatureOperation {
					debug.PrintStack()
					http.Error(w, "not implemented", http.StatusNotImplemented)
					return
				}
				filterByDeviceTypeId = query.Find.Filter.And[1].Condition.Value.([]interface{})
			}

			result := []TestPermSearchDevice{}
			for _, d := range deviceInstances {
				if strings.Contains(d.Name, query.Find.Search) {
					foundIdInNotIdList := false
					for _, notIdInterface := range notIdList {
						if notId, ok := notIdInterface.(string); ok && d.Id == notId {
							foundIdInNotIdList = true
							break
						}
					}

					foundTypeIdInList := filterByDeviceTypeId == nil
					if !foundTypeIdInList {
						for _, typeIdInterface := range filterByDeviceTypeId {
							if typeId, ok := typeIdInterface.(string); ok && d.DeviceTypeId == typeId {
								foundTypeIdInList = true
								break
							}
						}
					}

					if !foundIdInNotIdList && foundTypeIdInList {
						result = append(result, TestPermSearchDevice{Id: d.Id, Name: d.Name, DeviceType: d.DeviceTypeId, Permissions: model.Permissions{R: true, W: true, X: true, A: true}})
					}
				}
			}
			json.NewEncoder(w).Encode(result)
			return
		}
		log.Println("DEBUG: search call: " + r.URL.Path + "?" + r.URL.RawQuery)
		http.Error(w, "not implemented", http.StatusNotImplemented)
	}))

	c := &configuration.ConfigStruct{
		PermSearchUrl:   searchmock.URL,
		DeviceRepoUrl:   devicerepomock.URL,
		SemanticRepoUrl: semanticRepoMock.URL,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	repo, err := controller.New(ctx, c)
	if err != nil {
		searchmock.Close()
		selectionApi.Close()
		devicerepomock.Close()
		semanticRepoMock.Close()
		return semanticRepoMock, searchmock, devicerepomock, selectionApi, err
	}

	router := api.Router(c, repo)
	selectionApi = httptest.NewServer(router)

	return
}
