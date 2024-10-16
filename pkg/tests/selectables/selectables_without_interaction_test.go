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
	"github.com/SENERGY-Platform/device-selection/pkg/configuration"
	"github.com/SENERGY-Platform/device-selection/pkg/controller"
	"github.com/SENERGY-Platform/device-selection/pkg/model"
	"github.com/SENERGY-Platform/device-selection/pkg/model/devicemodel"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/environment"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/environment/kafka"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/environment/legacy"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/helper"
	kafka2 "github.com/segmentio/kafka-go"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"
)

func TestSelectableWithoutInteractionFilter(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkaUrl, deviceManagerUrl, deviceRepoUrl, permSearchUrl, _, importRepoUrl, importDeployUrl, err := environment.NewWithImport(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	c := &configuration.ConfigStruct{
		PermSearchUrl:                   permSearchUrl,
		DeviceRepoUrl:                   deviceRepoUrl,
		ImportRepoUrl:                   importRepoUrl,
		ImportDeployUrl:                 importDeployUrl,
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
	airAspect := "urn:infai:ses:aspect:airAspect"
	aspects := []devicemodel.Aspect{
		{Id: deviceAspect},
		{Id: lightAspect},
		{Id: airAspect},
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

	getHumidityFunction := devicemodel.MEASURING_FUNCTION_PREFIX + "getHumidityFunction"

	colorConcept := "urn:infai:ses:concept:color"
	humidityConcept := "urn:infai:ses:concept:humidity"

	testCharacteristic := "urn:infai:ses:characteristic:test"

	functionProducer, err := kafka.GetProducer([]string{kafkaUrl}, environment.FunctionTopic)
	if err != nil {
		t.Error(err)
		return
	}

	err = functionProducer.WriteMessages(ctx, kafka2.Message{Key: []byte(getColorFunction), Value: []byte("" +
		"{\"command\":\"PUT\",\"id\":\"" + getColorFunction + "\",\"owner\":\"1234567890\",\"function\":{\"id\":\"" + getColorFunction + "\",\"name\":\"getColorFunction\",\"description\":\"\"," +
		"\"concept_id\":\"" + colorConcept + "\",\"rdf_type\":\"https://senergy.infai.org/ontology/MeasuringFunction\"}}")})
	if err != nil {
		t.Error(err)
		return
	}
	err = functionProducer.WriteMessages(ctx, kafka2.Message{Key: []byte(getHumidityFunction), Value: []byte("" +
		"{\"command\":\"PUT\",\"id\":\"" + getHumidityFunction + "\",\"owner\":\"1234567890\",\"function\":{\"id\":\"" + getHumidityFunction + "\",\"name\":\"getColorFunction\",\"description\":\"\"," +
		"\"concept_id\":\"" + humidityConcept + "\",\"rdf_type\":\"https://senergy.infai.org/ontology/MeasuringFunction\"}}")})
	if err != nil {
		t.Error(err)
		return
	}

	conceptProducer, err := kafka.GetProducer([]string{kafkaUrl}, environment.ConceptTopic)
	if err != nil {
		t.Error(err)
		return
	}
	err = conceptProducer.WriteMessages(ctx, kafka2.Message{Key: []byte(colorConcept), Value: []byte("" +
		"{\"command\":\"PUT\",\"id\":\"" + colorConcept + "\",\"owner\":\"1234567890\",\"concept\":{\"id\":\"" + colorConcept + "\",\"name\":\"\",\"characteristic_ids\":[\"" + testCharacteristic + "\"],\"base_characteristic_id\": \"\",\"rdf_type\": \"https://senergy.infai.org/ontology/Concept\"   } }")})
	if err != nil {
		t.Error(err)
		return
	}

	err = conceptProducer.WriteMessages(ctx, kafka2.Message{Key: []byte(humidityConcept), Value: []byte("" +
		"{\"command\":\"PUT\",\"id\":\"" + humidityConcept + "\",\"owner\":\"1234567890\",\"concept\":{\"id\":\"" + humidityConcept + "\",\"name\":\"\",\"characteristic_ids\":[],\"base_characteristic_id\": \"\",\"rdf_type\": \"https://senergy.infai.org/ontology/Concept\"   } }")})
	if err != nil {
		t.Error(err)
		return
	}

	lamp := model.ImportType{}
	t.Run("create import-types lamp", testCreateImportTypes(helper.AdminJwt, kafkaUrl, importRepoUrl, model.ImportType{
		Name: "lamp",
		Output: model.ImportContentVariable{
			Name: "output",
			SubContentVariables: []model.ImportContentVariable{
				{
					Name: "value",
					SubContentVariables: []model.ImportContentVariable{
						{
							Name:             "value",
							CharacteristicId: testCharacteristic,
							AspectId:         lightAspect,
							FunctionId:       getColorFunction,
						},
						{
							Name:       "foo",
							AspectId:   airAspect,
							FunctionId: getHumidityFunction,
						},
					},
				},
			},
		},
	}, &lamp))

	never := model.ImportType{}
	t.Run("create import-types lamp", testCreateImportTypes(helper.AdminJwt, kafkaUrl, importRepoUrl, model.ImportType{
		Name: "never",
		Output: model.ImportContentVariable{
			Name: "output",
		},
	}, &never))

	importInstances := []model.Import{
		{
			Id:           "lamp-instance",
			Name:         "lamp-instance",
			ImportTypeId: lamp.Id,
		},
		{
			Id:           "never-instance",
			Name:         "never-instance",
			ImportTypeId: never.Id,
		},
	}

	t.Run("create imports", testCreateImports(importDeployUrl, importInstances))

	time.Sleep(5 * time.Second)

	t.Run("selection 1", func(t *testing.T) {
		result, err, _ := ctrl.GetFilteredDevicesV2(helper.AdminJwt, model.FilterCriteriaAndSet{{
			FunctionId: getColorFunction,
			AspectId:   lightAspect,
		}}, true, true, true, nil, false, false)
		if err != nil {
			t.Error(err)
		}

		expectedIds := []string{"colorlamp1_s8", "colorlamp2_s8", "dg_colorlamp", "lamp-instance"}
		resultIds := selectableToStringList(result)
		if !reflect.DeepEqual(resultIds, expectedIds) {
			r, _ := json.Marshal(resultIds)
			e, _ := json.Marshal(expectedIds)
			t.Errorf("%#v", resultIds)
			t.Errorf("\n%v\n%v", string(r), string(e))
			return
		}
	})

	t.Run("import path options", func(t *testing.T) {
		untrimmed, err, _ := ctrl.GetFilteredDevicesV2(helper.AdminJwt, model.FilterCriteriaAndSet{{
			FunctionId: getColorFunction,
			AspectId:   lightAspect,
		}}, false, false, true, nil, false, false)
		if err != nil {
			t.Error(err)
		}
		pathsUntrimmed := []string{}
		for _, element := range untrimmed {
			for _, pathOptions := range element.ServicePathOptions {
				for _, option := range pathOptions {
					pathsUntrimmed = append(pathsUntrimmed, option.Path)
				}
			}
		}
		trimmed, err, _ := ctrl.GetFilteredDevicesV2(helper.AdminJwt, model.FilterCriteriaAndSet{{
			FunctionId: getColorFunction,
			AspectId:   lightAspect,
		}}, false, false, true, nil, false, true)
		if err != nil {
			t.Error(err)
		}
		pathsTrimmed := []string{}
		for _, element := range trimmed {
			for _, pathOptions := range element.ServicePathOptions {
				for _, option := range pathOptions {
					pathsTrimmed = append(pathsTrimmed, option.Path)
				}
			}
		}
		sort.Strings(pathsUntrimmed)
		sort.Strings(pathsTrimmed)
		t.Logf("\n%#v\n%#v\n", pathsUntrimmed, pathsTrimmed)
		if !reflect.DeepEqual(pathsUntrimmed, []string{"output.value.value"}) {
			t.Errorf("%#v", pathsUntrimmed)
			return
		}
		if !reflect.DeepEqual(pathsTrimmed, []string{"value.value"}) {
			t.Errorf("%#v", pathsTrimmed)
			return
		}
	})
}

func selectableToStringList(selectables []model.Selectable) (result []string) {
	for _, selectable := range selectables {
		if selectable.Device != nil {
			for _, service := range selectable.Services {
				result = append(result, selectable.Device.Name+"_"+service.Id)
			}
		}
		if selectable.DeviceGroup != nil {
			result = append(result, selectable.DeviceGroup.Id)
		}
		if selectable.Import != nil {
			result = append(result, selectable.Import.Id)
		}
	}
	sort.Strings(result)
	return
}
