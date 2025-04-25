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

package selectables

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/device-selection/pkg/api"
	"github.com/SENERGY-Platform/device-selection/pkg/client"
	"github.com/SENERGY-Platform/device-selection/pkg/configuration"
	"github.com/SENERGY-Platform/device-selection/pkg/controller"
	"github.com/SENERGY-Platform/device-selection/pkg/model"
	"github.com/SENERGY-Platform/device-selection/pkg/model/devicemodel"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/environment/docker"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/environment/legacy"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/helper"
	"github.com/SENERGY-Platform/models/go/models"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestSelectableLocalId(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkaUrl, deviceManagerUrl, deviceRepoUrl, _, err := docker.DeviceManagerWithDependenciesAndKafka(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	port, err := docker.GetFreePort()
	if err != nil {
		t.Error(err)
		return
	}

	c := &configuration.ConfigStruct{
		ApiPort:                         strconv.Itoa(port),
		DeviceRepoUrl:                   deviceRepoUrl,
		Debug:                           true,
		KafkaUrl:                        kafkaUrl,
		KafkaConsumerGroup:              "device_selection",
		KafkaTopicsForCacheInvalidation: []string{"device-types", "aspects", "functions"},
	}

	ctrl, err := controller.New(ctx, c)
	if err != nil {
		t.Error(err)
		return
	}

	err = api.Start(ctx, wg, c, ctrl)
	if err != nil {
		t.Error(err)
		return
	}

	deviceAspect := "urn:infai:ses:aspect:deviceAspect"
	lightAspect := "urn:infai:ses:aspect:ligthAspect"
	aspects := []devicemodel.Aspect{
		{Id: deviceAspect},
		{Id: lightAspect},
	}

	setOnFunction := devicemodel.CONTROLLING_FUNCTION_PREFIX + "setOnFunction"
	setOffFunction := devicemodel.CONTROLLING_FUNCTION_PREFIX + "setOffFunction"
	setColorFunction := devicemodel.CONTROLLING_FUNCTION_PREFIX + "setColorFunction"
	getStateFunction := devicemodel.MEASURING_FUNCTION_PREFIX + "getStateFunction"
	getColorFunction := devicemodel.MEASURING_FUNCTION_PREFIX + "getColorFunction"
	functions := []devicemodel.Function{
		{Id: setOnFunction},
		{Id: setOffFunction},
		{Id: setColorFunction},
		{Id: getStateFunction},
		{Id: getColorFunction},
	}

	lampDeviceClass := "urn:infai:ses:device-class:lampClass"
	plugDeviceClass := "urn:infai:ses:device-class:plugClass"

	deviceTypes := []devicemodel.DeviceType{
		{
			Id:            "lamp",
			Name:          "lamp",
			DeviceClassId: lampDeviceClass,
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "s1", Name: "s1", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOnFunction}},
				{Id: "s2", Name: "s2", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOffFunction}},
				{Id: "s3", Name: "s3", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{getStateFunction}},
			}),
		},
		{
			Id:            "both_lamp",
			Name:          "both_lamp",
			DeviceClassId: lampDeviceClass,
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "sb1", Name: "sb1", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOnFunction}},
				{Id: "sb2", Name: "sb2", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOffFunction}},
				{Id: "sb3", Name: "sb3", Interaction: devicemodel.EVENT_AND_REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{getStateFunction}},
			}),
		},
		{
			Id:            "event_lamp",
			Name:          "event_lamp",
			DeviceClassId: lampDeviceClass,
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "se1", Name: "se1", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOnFunction}},
				{Id: "se2", Name: "se2", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOffFunction}},
				{Id: "se3", Name: "se3", Interaction: devicemodel.EVENT, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{getStateFunction}},
			}),
		},
		{
			Id:            "colorlamp",
			Name:          "colorlamp",
			DeviceClassId: lampDeviceClass,
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "s4", Name: "s4", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOnFunction}},
				{Id: "s5", Name: "s5", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOffFunction}},
				{Id: "s6", Name: "s6", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{getStateFunction}},
				{Id: "s7", Name: "s7", Interaction: devicemodel.REQUEST, AspectIds: []string{lightAspect}, FunctionIds: []string{setColorFunction}},
				{Id: "s8", Name: "s8", Interaction: devicemodel.REQUEST, AspectIds: []string{lightAspect}, FunctionIds: []string{getColorFunction}},
			}),
		},
		{
			Id:            "plug",
			Name:          "plug",
			DeviceClassId: plugDeviceClass,
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "s9", Name: "s9", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect}, FunctionIds: []string{setOnFunction}},
				{Id: "s10", Name: "s10", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect}, FunctionIds: []string{setOffFunction}},
				{Id: "s11", Name: "s11", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect}, FunctionIds: []string{getStateFunction}},
			}),
		},
	}

	devicesInstances := []devicemodel.Device{
		{
			Id:           "elamp",
			Name:         "elamp",
			LocalId:      "elamp",
			DeviceTypeId: "event_lamp",
		},
		{
			Id:           "blamp",
			Name:         "blamp",
			LocalId:      "blamp",
			DeviceTypeId: "both_lamp",
		},
		{
			Id:           "lamp1",
			Name:         "lamp1",
			LocalId:      "lamp1",
			DeviceTypeId: "lamp",
		},
		{
			Id:           "lamp2",
			Name:         "lamp2",
			LocalId:      "lamp2",
			DeviceTypeId: "lamp",
		},
		{
			Id:           "colorlamp1",
			Name:         "colorlamp1",
			LocalId:      "colorlamp1",
			DeviceTypeId: "colorlamp",
		},
		{
			Id:           "colorlamp2",
			Name:         "colorlamp2",
			LocalId:      "colorlamp2",
			DeviceTypeId: "colorlamp",
		},
		{
			Id:           "plug1",
			Name:         "plug1",
			LocalId:      "plug1",
			DeviceTypeId: "plug",
		},
		{
			Id:           "plug2",
			Name:         "plug2",
			LocalId:      "plug2",
			DeviceTypeId: "plug",
		},
	}

	for _, a := range aspects {
		err = helper.SetAspect(deviceManagerUrl, a)
		if err != nil {
			t.Error(err)
			return
		}
	}
	for _, f := range functions {
		err = helper.SetFunction(deviceManagerUrl, f)
		if err != nil {
			t.Error(err)
			return
		}
	}

	t.Run("create device-types", testCreateDeviceTypes(deviceManagerUrl, deviceTypes))
	t.Run("create devices", testCreateDevices(deviceManagerUrl, devicesInstances))

	time.Sleep(5 * time.Second)

	t.Run("lamp on/off", testCheckSelectionWithLocalIdsWithoutOptions(ctrl, model.FilterCriteriaAndSet{
		{FunctionId: setOnFunction, DeviceClassId: lampDeviceClass, AspectId: ""},
		{FunctionId: setOffFunction, DeviceClassId: lampDeviceClass, AspectId: ""},
	}, devicemodel.EVENT, false, []string{"colorlamp1", "lamp2"}, []model.Selectable{
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "colorlamp1",
					Name:         "colorlamp1",
					DeviceTypeId: "colorlamp",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "s4", Name: "s4", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOnFunction}},
				{Id: "s5", Name: "s5", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOffFunction}},
			}),
		},
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "s1", Name: "s1", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOnFunction}},
				{Id: "s2", Name: "s2", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOffFunction}},
			}),
		},
	}))

	t.Run("lamp on/off with id", testCheckSelectionV2WithDeviceIds(c, []models.DeviceGroupFilterCriteria{
		{FunctionId: setOnFunction, DeviceClassId: lampDeviceClass, AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: setOffFunction, DeviceClassId: lampDeviceClass, AspectId: "", Interaction: devicemodel.REQUEST},
	}, false, []string{"colorlamp1", "lamp2"}, []model.Selectable{
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "colorlamp1",
					Name:         "colorlamp1",
					DeviceTypeId: "colorlamp",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "s4", Name: "s4", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOnFunction}},
				{Id: "s5", Name: "s5", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOffFunction}},
			}),
		},
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "s1", Name: "s1", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOnFunction}},
				{Id: "s2", Name: "s2", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOffFunction}},
			}),
		},
	}))

	t.Run("lamp on/off with local id and owner", testCheckSelectionV2WithLocalDeviceIds(c, []models.DeviceGroupFilterCriteria{
		{FunctionId: setOnFunction, DeviceClassId: lampDeviceClass, AspectId: "", Interaction: devicemodel.REQUEST},
		{FunctionId: setOffFunction, DeviceClassId: lampDeviceClass, AspectId: "", Interaction: devicemodel.REQUEST},
	}, false, []string{"colorlamp1", "lamp2"}, helper.JwtSubject, []model.Selectable{
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "colorlamp1",
					Name:         "colorlamp1",
					DeviceTypeId: "colorlamp",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "s4", Name: "s4", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOnFunction}},
				{Id: "s5", Name: "s5", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOffFunction}},
			}),
		},
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "lamp2",
					Name:         "lamp2",
					DeviceTypeId: "lamp",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "s1", Name: "s1", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOnFunction}},
				{Id: "s2", Name: "s2", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOffFunction}},
			}),
		},
	}))
}

func testCheckSelectionWithLocalIdsWithoutOptions(ctrl *controller.Controller, criteria model.FilterCriteriaAndSet, interaction devicemodel.Interaction, includeGroups bool, localIds []string, expectedResult []model.Selectable) func(t *testing.T) {
	return func(t *testing.T) {
		result, err, _ := ctrl.GetFilteredDevices(token, criteria, nil, interaction, includeGroups, false, localIds)
		if err != nil {
			t.Error(err)
			return
		}
		for i, e := range result {
			e.ServicePathOptions = nil
			result[i] = e
		}
		for i, e := range expectedResult {
			e.ServicePathOptions = nil
			expectedResult[i] = e
		}
		normalizeTestSelectables(&result, true)
		normalizeTestSelectables(&expectedResult, true)
		if !reflect.DeepEqual(result, expectedResult) {
			resultJson, _ := json.Marshal(result)
			expectedJson, _ := json.Marshal(expectedResult)
			t.Error("\n", string(resultJson), "\n", string(expectedJson))
		}
	}
}

func testCheckSelectionV2WithDeviceIds(config configuration.Config, criteria []models.DeviceGroupFilterCriteria, includeGroups bool, ids []string, expectedResult []model.Selectable) func(t *testing.T) {
	return func(t *testing.T) {
		c := client.NewClient("http://localhost:" + config.ApiPort)
		result, _, err := c.GetSelectables(token, criteria, &client.GetSelectablesOptions{
			IncludeGroups:  includeGroups,
			IncludeDevices: true,
			WithDeviceIds:  ids,
		})
		if err != nil {
			t.Error(err)
			return
		}
		for i, e := range result {
			e.ServicePathOptions = nil
			result[i] = e
		}
		for i, e := range expectedResult {
			e.ServicePathOptions = nil
			expectedResult[i] = e
		}
		normalizeTestSelectables(&result, true)
		normalizeTestSelectables(&expectedResult, true)
		if !reflect.DeepEqual(result, expectedResult) {
			resultJson, _ := json.Marshal(result)
			expectedJson, _ := json.Marshal(expectedResult)
			t.Error("\n", string(resultJson), "\n", string(expectedJson))
		}
	}
}

func testCheckSelectionV2WithLocalDeviceIds(config configuration.Config, criteria []models.DeviceGroupFilterCriteria, includeGroups bool, localIds []string, owner string, expectedResult []model.Selectable) func(t *testing.T) {
	return func(t *testing.T) {
		//use different admin token to test LocalDeviceOwner field
		testAdminToken := `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjEwMDAwMDAwMDAsImlhdCI6MTAwMDAwMDAwMCwiYXV0aF90aW1lIjoxMDAwMDAwMDAwLCJpc3MiOiJpbnRlcm5hbCIsImF1ZCI6W10sInN1YiI6InRlc3QiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJhZG1pbiIsImRldmVsb3BlciIsInVzZXIiXX0sInJlc291cmNlX2FjY2VzcyI6eyJtYXN0ZXItcmVhbG0iOnsicm9sZXMiOltdfSwiQmFja2VuZC1yZWFsbSI6eyJyb2xlcyI6W119LCJhY2NvdW50Ijp7InJvbGVzIjpbXX19LCJyb2xlcyI6WyJhZG1pbiIsImRldmVsb3BlciIsInVzZXIiXSwibmFtZSI6IlNlcGwgQWRtaW4iLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJzZXBsIiwiZ2l2ZW5fbmFtZSI6IlNlcGwiLCJsb2NhbGUiOiJlbiIsImZhbWlseV9uYW1lIjoiQWRtaW4iLCJlbWFpbCI6InNlcGxAc2VwbC5kZSJ9.-A5JKZptW3UrvaK5vD06eVuzoa1snijdneUOjDCKt6w`
		c := client.NewClient("http://localhost:" + config.ApiPort)
		result, _, err := c.GetSelectables(testAdminToken, criteria, &client.GetSelectablesOptions{
			IncludeGroups:      includeGroups,
			IncludeDevices:     true,
			WithLocalDeviceIds: localIds,
			LocalDeviceOwner:   owner,
		})
		if err != nil {
			t.Error(err)
			return
		}
		for i, e := range result {
			e.ServicePathOptions = nil
			result[i] = e
		}
		for i, e := range expectedResult {
			e.ServicePathOptions = nil
			expectedResult[i] = e
		}
		normalizeTestSelectables(&result, true)
		normalizeTestSelectables(&expectedResult, true)
		if !reflect.DeepEqual(result, expectedResult) {
			resultJson, _ := json.Marshal(result)
			expectedJson, _ := json.Marshal(expectedResult)
			t.Error("\n", string(resultJson), "\n", string(expectedJson))
		}
	}
}
