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
	"bytes"
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/device-selection/pkg/configuration"
	"github.com/SENERGY-Platform/device-selection/pkg/controller"
	"github.com/SENERGY-Platform/device-selection/pkg/model"
	"github.com/SENERGY-Platform/device-selection/pkg/model/devicemodel"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/environment/docker"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/environment/legacy"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/helper"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"
)

func TestSelectableGroups(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkaUrl, deviceManagerUrl, deviceRepoUrl, _, err := docker.DeviceManagerWithDependenciesAndKafka(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	c := &configuration.ConfigStruct{
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

	deviceGroups := []devicemodel.DeviceGroup{
		{
			Id:   "dg_lamp",
			Name: "dg_lamp",
			Criteria: []devicemodel.DeviceGroupFilterCriteria{
				{FunctionId: setOnFunction, DeviceClassId: lampDeviceClass, AspectId: "", Interaction: devicemodel.REQUEST},
				{FunctionId: setOffFunction, DeviceClassId: lampDeviceClass, AspectId: "", Interaction: devicemodel.REQUEST},
				{FunctionId: getStateFunction, DeviceClassId: "", AspectId: deviceAspect, Interaction: devicemodel.REQUEST},
				{FunctionId: getStateFunction, DeviceClassId: "", AspectId: lightAspect, Interaction: devicemodel.REQUEST},
			},
			DeviceIds: []string{"lamp1", "colorlamp1"},
		},
		{
			Id:   "dg_colorlamp",
			Name: "dg_colorlamp",
			Criteria: []devicemodel.DeviceGroupFilterCriteria{
				{FunctionId: setColorFunction, DeviceClassId: lampDeviceClass, AspectId: "", Interaction: devicemodel.REQUEST},
				{FunctionId: setOnFunction, DeviceClassId: lampDeviceClass, AspectId: "", Interaction: devicemodel.REQUEST},
				{FunctionId: setOffFunction, DeviceClassId: lampDeviceClass, AspectId: "", Interaction: devicemodel.REQUEST},
				{FunctionId: getStateFunction, DeviceClassId: "", AspectId: deviceAspect, Interaction: devicemodel.REQUEST},
				{FunctionId: getStateFunction, DeviceClassId: "", AspectId: lightAspect, Interaction: devicemodel.REQUEST},
				{FunctionId: getColorFunction, DeviceClassId: "", AspectId: lightAspect, Interaction: devicemodel.REQUEST},
			},
			DeviceIds: []string{"colorlamp1"},
		},
		{
			Id:   "dg_plug",
			Name: "dg_plug",
			Criteria: []devicemodel.DeviceGroupFilterCriteria{
				{FunctionId: setOnFunction, DeviceClassId: plugDeviceClass, AspectId: "", Interaction: devicemodel.REQUEST},
				{FunctionId: setOffFunction, DeviceClassId: plugDeviceClass, AspectId: "", Interaction: devicemodel.REQUEST},
				{FunctionId: getStateFunction, DeviceClassId: "", AspectId: deviceAspect, Interaction: devicemodel.REQUEST},
			},
			DeviceIds: []string{"plug1", "plug2"},
		},
		{
			Id:   "dg_event_lamp",
			Name: "eventlamps",
			Criteria: []devicemodel.DeviceGroupFilterCriteria{
				{FunctionId: setOnFunction, DeviceClassId: lampDeviceClass, AspectId: "", Interaction: devicemodel.REQUEST},
				{FunctionId: setOffFunction, DeviceClassId: lampDeviceClass, AspectId: "", Interaction: devicemodel.REQUEST},
				{FunctionId: getStateFunction, DeviceClassId: "", AspectId: lightAspect, Interaction: devicemodel.EVENT},
				{FunctionId: getStateFunction, DeviceClassId: "", AspectId: deviceAspect, Interaction: devicemodel.EVENT},
			},
			DeviceIds: []string{"elamp"},
		},
		{
			Id:   "dg_both_lamp",
			Name: "bothlamps",
			Criteria: []devicemodel.DeviceGroupFilterCriteria{
				{FunctionId: getStateFunction, DeviceClassId: "", AspectId: lightAspect, Interaction: devicemodel.EVENT},
				{FunctionId: getStateFunction, DeviceClassId: "", AspectId: deviceAspect, Interaction: devicemodel.EVENT},
				{FunctionId: getStateFunction, DeviceClassId: "", AspectId: lightAspect, Interaction: devicemodel.REQUEST},
				{FunctionId: getStateFunction, DeviceClassId: "", AspectId: deviceAspect, Interaction: devicemodel.REQUEST},
				{FunctionId: setOnFunction, DeviceClassId: lampDeviceClass, AspectId: "", Interaction: devicemodel.REQUEST},
				{FunctionId: setOffFunction, DeviceClassId: lampDeviceClass, AspectId: "", Interaction: devicemodel.REQUEST},
			},
			DeviceIds: []string{"blamp"},
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
	t.Run("create devices-groups", testCreateDeviceGroups(deviceManagerUrl, deviceGroups))

	time.Sleep(5 * time.Second)

	t.Run("lamp on/off", testCheckSelectionWithoutOptions(ctrl, model.FilterCriteriaAndSet{
		{FunctionId: setOnFunction, DeviceClassId: lampDeviceClass, AspectId: ""},
		{FunctionId: setOffFunction, DeviceClassId: lampDeviceClass, AspectId: ""},
	}, devicemodel.EVENT, false, []model.Selectable{
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
					Id:           "colorlamp2",
					Name:         "colorlamp2",
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
					Id:           "lamp1",
					Name:         "lamp1",
					DeviceTypeId: "lamp",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "s1", Name: "s1", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOnFunction}},
				{Id: "s2", Name: "s2", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOffFunction}},
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
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "se1", Name: "se1", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOnFunction}},
				{Id: "se2", Name: "se2", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOffFunction}},
			}),
		},
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "sb1", Name: "sb1", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOnFunction}},
				{Id: "sb2", Name: "sb2", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOffFunction}},
			}),
		},
	}))

	t.Run("get lamp state", testCheckSelectionWithoutOptions(ctrl, model.FilterCriteriaAndSet{
		{FunctionId: getStateFunction, DeviceClassId: "", AspectId: lightAspect},
	}, devicemodel.EVENT, false, []model.Selectable{
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
				{Id: "s6", Name: "s6", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{getStateFunction}},
			}),
		},
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "s6", Name: "s6", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{getStateFunction}},
			}),
		},
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "lamp1",
					Name:         "lamp1",
					DeviceTypeId: "lamp",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "s3", Name: "s3", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{getStateFunction}},
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
				{Id: "s3", Name: "s3", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{getStateFunction}},
			}),
		},
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "sb3", Name: "sb3", Interaction: devicemodel.EVENT_AND_REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{getStateFunction}},
			}),
		},
	}))

	t.Run("get lamp state as event", testCheckSelectionWithoutOptions(ctrl, model.FilterCriteriaAndSet{
		{FunctionId: getStateFunction, DeviceClassId: "", AspectId: lightAspect},
	}, devicemodel.REQUEST, false, []model.Selectable{
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "sb3", Name: "sb3", Interaction: devicemodel.EVENT_AND_REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{getStateFunction}},
			}),
		},
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "se3", Name: "se3", Interaction: devicemodel.EVENT, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{getStateFunction}},
			}),
		},
	}))

	t.Run("lamp on/off with group", testCheckSelectionWithoutOptions(ctrl, model.FilterCriteriaAndSet{
		{FunctionId: setOnFunction, DeviceClassId: lampDeviceClass, AspectId: ""},
		{FunctionId: setOffFunction, DeviceClassId: lampDeviceClass, AspectId: ""},
	}, devicemodel.EVENT, true, []model.Selectable{
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
					Id:           "colorlamp2",
					Name:         "colorlamp2",
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
					Id:           "lamp1",
					Name:         "lamp1",
					DeviceTypeId: "lamp",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "s1", Name: "s1", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOnFunction}},
				{Id: "s2", Name: "s2", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOffFunction}},
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
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "se1", Name: "se1", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOnFunction}},
				{Id: "se2", Name: "se2", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOffFunction}},
			}),
		},
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "sb1", Name: "sb1", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOnFunction}},
				{Id: "sb2", Name: "sb2", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{setOffFunction}},
			}),
		},
		{
			DeviceGroup: &model.DeviceGroup{
				Id:   "dg_colorlamp",
				Name: "dg_colorlamp",
			},
		},
		{
			DeviceGroup: &model.DeviceGroup{
				Id:   "dg_event_lamp",
				Name: "eventlamps",
			},
		},
		{
			DeviceGroup: &model.DeviceGroup{
				Id:   "dg_both_lamp",
				Name: "bothlamps",
			},
		},
		{
			DeviceGroup: &model.DeviceGroup{
				Id:   "dg_lamp",
				Name: "dg_lamp",
			},
		},
	}))

	t.Run("lamp get color with group", testCheckSelectionWithoutOptions(ctrl, model.FilterCriteriaAndSet{
		{FunctionId: getColorFunction, DeviceClassId: "", AspectId: lightAspect},
	}, devicemodel.EVENT, true, []model.Selectable{
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
				{Id: "s8", Name: "s8", Interaction: devicemodel.REQUEST, AspectIds: []string{lightAspect}, FunctionIds: []string{getColorFunction}},
			}),
		},
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "s8", Name: "s8", Interaction: devicemodel.REQUEST, AspectIds: []string{lightAspect}, FunctionIds: []string{getColorFunction}},
			}),
		},
		{
			DeviceGroup: &model.DeviceGroup{
				Id:   "dg_colorlamp",
				Name: "dg_colorlamp",
			},
		},
	}))

	t.Run("plug on/off with group", testCheckSelectionWithoutOptions(ctrl, model.FilterCriteriaAndSet{
		{FunctionId: setOnFunction, DeviceClassId: plugDeviceClass, AspectId: ""},
		{FunctionId: setOffFunction, DeviceClassId: plugDeviceClass, AspectId: ""},
	}, devicemodel.EVENT, true, []model.Selectable{
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "plug1",
					Name:         "plug1",
					DeviceTypeId: "plug",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "s9", Name: "s9", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect}, FunctionIds: []string{setOnFunction}},
				{Id: "s10", Name: "s10", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect}, FunctionIds: []string{setOffFunction}},
			}),
		},
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "plug2",
					Name:         "plug2",
					DeviceTypeId: "plug",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "s9", Name: "s9", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect}, FunctionIds: []string{setOnFunction}},
				{Id: "s10", Name: "s10", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect}, FunctionIds: []string{setOffFunction}},
			}),
		},
		{
			DeviceGroup: &model.DeviceGroup{
				Id:   "dg_plug",
				Name: "dg_plug",
			},
		},
	}))

	t.Run("get lamp state with groups", testCheckSelectionWithoutOptions(ctrl, model.FilterCriteriaAndSet{
		{FunctionId: getStateFunction, DeviceClassId: "", AspectId: lightAspect},
	}, devicemodel.EVENT, true, []model.Selectable{
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
				{Id: "s6", Name: "s6", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{getStateFunction}},
			}),
		},
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "colorlamp2",
					Name:         "colorlamp2",
					DeviceTypeId: "colorlamp",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "s6", Name: "s6", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{getStateFunction}},
			}),
		},
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "lamp1",
					Name:         "lamp1",
					DeviceTypeId: "lamp",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "s3", Name: "s3", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{getStateFunction}},
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
				{Id: "s3", Name: "s3", Interaction: devicemodel.REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{getStateFunction}},
			}),
		},
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "sb3", Name: "sb3", Interaction: devicemodel.EVENT_AND_REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{getStateFunction}},
			}),
		},
		{
			DeviceGroup: &model.DeviceGroup{
				Id:   "dg_lamp",
				Name: "dg_lamp",
			},
		},
		{
			DeviceGroup: &model.DeviceGroup{
				Id:   "dg_colorlamp",
				Name: "dg_colorlamp",
			},
		},
		{
			DeviceGroup: &model.DeviceGroup{
				Id:   "dg_both_lamp",
				Name: "bothlamps",
			},
		},
	}))

	t.Run("get lamp state as event with groups", testCheckSelectionWithoutOptions(ctrl, model.FilterCriteriaAndSet{
		{FunctionId: getStateFunction, DeviceClassId: "", AspectId: lightAspect},
	}, devicemodel.REQUEST, true, []model.Selectable{
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "blamp",
					Name:         "blamp",
					DeviceTypeId: "both_lamp",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "sb3", Name: "sb3", Interaction: devicemodel.EVENT_AND_REQUEST, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{getStateFunction}},
			}),
		},
		{
			Device: &model.PermSearchDevice{
				Device: devicemodel.Device{
					Id:           "elamp",
					Name:         "elamp",
					DeviceTypeId: "event_lamp",
					OwnerId:      helper.JwtSubject,
				},
			},
			Services: legacy.FromLegacyServices([]legacy.Service{
				{Id: "se3", Name: "se3", Interaction: devicemodel.EVENT, AspectIds: []string{deviceAspect, lightAspect}, FunctionIds: []string{getStateFunction}},
			}),
		},
		{
			DeviceGroup: &model.DeviceGroup{
				Id:   "dg_both_lamp",
				Name: "bothlamps",
			},
		},
		{
			DeviceGroup: &model.DeviceGroup{
				Id:   "dg_event_lamp",
				Name: "eventlamps",
			},
		},
	}))
}

const token = helper.AdminJwt

func testCheckSelectionWithoutOptions(ctrl *controller.Controller, criteria model.FilterCriteriaAndSet, interaction devicemodel.Interaction, includeGroups bool, expectedResult []model.Selectable) func(t *testing.T) {
	return func(t *testing.T) {
		result, err, _ := ctrl.GetFilteredDevices(token, criteria, nil, interaction, includeGroups, false, nil)
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
			t.Errorf("\n%v\n%v\n", string(resultJson), string(expectedJson))
		}
	}
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

func testCreateDeviceGroups(managerUrl string, groups []devicemodel.DeviceGroup) func(t *testing.T) {
	return func(t *testing.T) {
		for _, group := range groups {
			buff := new(bytes.Buffer)
			err := json.NewEncoder(buff).Encode(group)
			if err != nil {
				t.Error(err)
				return
			}
			var req *http.Request
			if group.Id != "" {
				req, err = http.NewRequest("PUT", managerUrl+"/device-groups/"+url.PathEscape(group.Id), buff)
			} else {
				req, err = http.NewRequest("POST", managerUrl+"/device-groups", buff)
			}
			if err != nil {
				t.Error(err)
				return
			}
			req.Header.Set("Authorization", token)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Error(err)
				return
			}
			if resp.StatusCode != 200 {
				temp, _ := io.ReadAll(resp.Body)
				t.Error(resp.StatusCode, string(temp))
			}
		}
	}
}

func testCreateDevices(managerUrl string, devices []devicemodel.Device) func(t *testing.T) {
	return func(t *testing.T) {
		for _, device := range devices {
			buff := new(bytes.Buffer)
			err := json.NewEncoder(buff).Encode(device)
			if err != nil {
				t.Error(err)
				return
			}

			var req *http.Request
			if device.Id != "" {
				req, err = http.NewRequest("PUT", managerUrl+"/devices/"+url.PathEscape(device.Id), buff)
			} else {
				req, err = http.NewRequest("POST", managerUrl+"/devices", buff)
			}
			if err != nil {
				t.Error(err)
				return
			}
			req.Header.Set("Authorization", token)
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
		}
	}
}

func testCreateDeviceTypes(managerUrl string, deviceTypes []devicemodel.DeviceType) func(t *testing.T) {
	return func(t *testing.T) {
		for _, deviceType := range deviceTypes {
			buff := new(bytes.Buffer)
			err := json.NewEncoder(buff).Encode(deviceType)
			if err != nil {
				t.Error(err)
				return
			}
			var req *http.Request
			if deviceType.Id != "" {
				req, err = http.NewRequest("PUT", managerUrl+"/device-types/"+url.PathEscape(deviceType.Id), buff)
			} else {
				req, err = http.NewRequest("POST", managerUrl+"/device-types", buff)
			}
			if err != nil {
				t.Error(err)
				return
			}
			req.Header.Set("Authorization", token)
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
		}
	}
}
