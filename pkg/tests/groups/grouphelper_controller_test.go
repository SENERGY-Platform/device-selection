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

package groups

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/device-selection/pkg/configuration"
	"github.com/SENERGY-Platform/device-selection/pkg/controller"
	"github.com/SENERGY-Platform/device-selection/pkg/model/devicemodel"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/environment/legacy"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/helper"
	"reflect"
	"sort"
	"sync"
	"testing"
)

func TestGroupHelperCriteria(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deviceTypes := []devicemodel.DeviceType{
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
				{Id: "sb1", Name: "sb1", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOn"}},
				{Id: "sb2", Name: "sb2", Interaction: devicemodel.REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{"setOff"}},
				{Id: "sb3", Name: "sb3", Interaction: devicemodel.EVENT_AND_REQUEST, AspectIds: []string{"device", "light"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getState"}},
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

		{
			Id:            "aspect-hierarchy-check-parent",
			Name:          "aspect-hierarchy-check-parent",
			DeviceClassId: "plug",
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "s12", Name: "s12", Interaction: devicemodel.REQUEST, AspectIds: []string{"device"}, FunctionIds: []string{"setOn"}},
				{Id: "s13", Name: "s13", Interaction: devicemodel.REQUEST, AspectIds: []string{"device"}, FunctionIds: []string{"setOff"}},
				{Id: "s14", Name: "s14", Interaction: devicemodel.REQUEST, AspectIds: []string{"device"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getState"}},
			}),
		},

		{
			Id:            "aspect-hierarchy-check-child",
			Name:          "aspect-hierarchy-check-child",
			DeviceClassId: "plug",
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "s15", Name: "s15", Interaction: devicemodel.REQUEST, AspectIds: []string{"horn"}, FunctionIds: []string{"setOn"}},
				{Id: "s16", Name: "s16", Interaction: devicemodel.REQUEST, AspectIds: []string{"horn"}, FunctionIds: []string{"setOff"}},
				{Id: "s17", Name: "s17", Interaction: devicemodel.REQUEST, AspectIds: []string{"horn"}, FunctionIds: []string{devicemodel.MEASURING_FUNCTION_PREFIX + "getState"}},
			}),
		},
	}

	devicesInstances := []devicemodel.Device{
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

		{
			Id:           "aspect-hierarchy-check-parent",
			Name:         "aspect-hierarchy-check-parent",
			DeviceTypeId: "aspect-hierarchy-check-parent",
		},
		{
			Id:           "aspect-hierarchy-check-child",
			Name:         "aspect-hierarchy-check-child",
			DeviceTypeId: "aspect-hierarchy-check-child",
		},
	}

	repo, err := grouphelpertestenv(ctx, wg, deviceTypes, devicesInstances)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("empty list", testGroupHelper(repo, []string{}, []devicemodel.DeviceGroupFilterCriteria{}))

	t.Run("lamp1", testGroupHelper(repo, []string{"lamp1"}, []devicemodel.DeviceGroupFilterCriteria{
		{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
	}))

	t.Run("elamp", testGroupHelper(repo, []string{"elamp"}, []devicemodel.DeviceGroupFilterCriteria{
		{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
	}))

	t.Run("blamp", testGroupHelper(repo, []string{"blamp"}, []devicemodel.DeviceGroupFilterCriteria{
		{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.EVENT},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.EVENT},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
	}))

	t.Run("lamp1 blamp", testGroupHelper(repo, []string{"lamp1", "blamp"}, []devicemodel.DeviceGroupFilterCriteria{
		{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
	}))

	t.Run("lamp1 elamp", testGroupHelper(repo, []string{"lamp1", "elamp"}, []devicemodel.DeviceGroupFilterCriteria{
		{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
	}))

	t.Run("colorlamp1", testGroupHelper(repo, []string{"colorlamp1"}, []devicemodel.DeviceGroupFilterCriteria{
		{FunctionId: "setColor", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getColor", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
	}))

	t.Run("plug2", testGroupHelper(repo, []string{"plug2"}, []devicemodel.DeviceGroupFilterCriteria{
		{FunctionId: "setOn", DeviceClassId: "plug", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: "setOff", DeviceClassId: "plug", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
	}))

	t.Run("lamp1 colorlamp1", testGroupHelper(repo, []string{"lamp1", "colorlamp1"}, []devicemodel.DeviceGroupFilterCriteria{
		{FunctionId: "setOn", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: "setOff", DeviceClassId: "lamp", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "light", Interaction: devicemodel.REQUEST},
	}))

	t.Run("lamp1 colorlamp1 plug1", testGroupHelper(repo, []string{"lamp1", "colorlamp1", "plug1"}, []devicemodel.DeviceGroupFilterCriteria{
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
	}))

	t.Run("aspect-hierarchy-check-parent", testGroupHelper(repo, []string{"aspect-hierarchy-check-parent"}, []devicemodel.DeviceGroupFilterCriteria{
		{FunctionId: "setOn", DeviceClassId: "plug", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: "setOff", DeviceClassId: "plug", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
	}))

	t.Run("aspect-hierarchy-check-child", testGroupHelper(repo, []string{"aspect-hierarchy-check-child"}, []devicemodel.DeviceGroupFilterCriteria{
		{FunctionId: "setOn", DeviceClassId: "plug", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: "setOff", DeviceClassId: "plug", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "components", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "horn", Interaction: devicemodel.REQUEST},
	}))

	t.Run("aspect-hierarchy-check", testGroupHelper(repo, []string{"aspect-hierarchy-check-parent", "aspect-hierarchy-check-child"}, []devicemodel.DeviceGroupFilterCriteria{
		{FunctionId: "setOn", DeviceClassId: "plug", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: "setOff", DeviceClassId: "plug", AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "getState", DeviceClassId: "", AspectId: "device", Interaction: devicemodel.REQUEST},
	}))
}

func testGroupHelper(repo *controller.Controller, deviceIds []string, expectedResult []devicemodel.DeviceGroupFilterCriteria) func(t *testing.T) {
	return func(t *testing.T) {
		dtCache := &map[string]devicemodel.DeviceType{}
		dCache := &map[string]devicemodel.Device{}
		result, err, code := repo.GetDeviceGroupCriteria(helper.AdminJwt, dtCache, dCache, deviceIds)
		if err != nil {
			t.Error(err, code)
			return
		}
		result = normalizeCriteria(result)
		expectedResult = normalizeCriteria(expectedResult)
		if !reflect.DeepEqual(result, expectedResult) {
			resultJson, _ := json.Marshal(result)
			expectedJson, _ := json.Marshal(expectedResult)
			t.Error(string(resultJson), string(expectedJson))
		}
	}
}

func normalizeCriteria(criteria []devicemodel.DeviceGroupFilterCriteria) []devicemodel.DeviceGroupFilterCriteria {
	sort.SliceStable(criteria, func(i, j int) bool {
		return criteria[i].AspectId < criteria[j].AspectId
	})
	sort.SliceStable(criteria, func(i, j int) bool {
		return criteria[i].FunctionId < criteria[j].FunctionId
	})
	sort.SliceStable(criteria, func(i, j int) bool {
		return criteria[i].Interaction < criteria[j].Interaction
	})
	return criteria
}

func grouphelpertestenv(ctx context.Context, wg *sync.WaitGroup, deviceTypes []devicemodel.DeviceType, deviceInstances []devicemodel.Device) (repo *controller.Controller, err error) {
	kafkaUrl, managerurl, repoUrl, searchurl, err := helper.EnvWithDevices(ctx, wg, deviceTypes, deviceInstances)
	if err != nil {
		return nil, err
	}
	err = helper.SetAspect(managerurl, devicemodel.Aspect{
		Id:   "device",
		Name: "device",
		SubAspects: []devicemodel.Aspect{
			{
				Id:   "components",
				Name: "components",
				SubAspects: []devicemodel.Aspect{
					{
						Id:   "horn",
						Name: "horn",
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	err = helper.SetAspect(managerurl, devicemodel.Aspect{
		Id:   "light",
		Name: "light",
	})
	if err != nil {
		return nil, err
	}
	c := &configuration.ConfigStruct{
		PermSearchUrl:                   searchurl,
		DeviceRepoUrl:                   repoUrl,
		KafkaUrl:                        kafkaUrl,
		KafkaConsumerGroup:              "device_selection",
		KafkaTopicsForCacheInvalidation: []string{"device-types", "aspects", "functions"},
	}
	return controller.New(ctx, c)
}
