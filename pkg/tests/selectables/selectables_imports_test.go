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
	"github.com/SENERGY-Platform/device-selection/pkg/tests/environment"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/environment/kafka"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/helper"
	kafka2 "github.com/segmentio/kafka-go"
	"io/ioutil"
	"net/http"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestSelectableImports(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkabroker, deviceManagerUrl, deviceRepoUrl, permSearchUrl, importRepoUrl, importDeployUrl, err := environment.NewWithImport(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	c := &configuration.ConfigStruct{
		PermSearchUrl:   permSearchUrl,
		DeviceRepoUrl:   deviceRepoUrl,
		ImportRepoUrl:   importRepoUrl,
		ImportDeployUrl: importDeployUrl,
		Debug:           true,
	}

	ctrl, err := controller.New(ctx, c)
	if err != nil {
		t.Error(err)
		return
	}

	deviceAspect := "urn:infai:ses:aspect:deviceAspect"
	deviceAspectNode := devicemodel.AspectNode{
		Id:            deviceAspect,
		Name:          "",
		RootId:        deviceAspect,
		ParentId:      "",
		ChildIds:      []string{},
		AncestorIds:   []string{},
		DescendentIds: []string{},
	}
	airAspect := "urn:infai:ses:aspect:airAspect"
	aspects := []devicemodel.Aspect{
		{Id: deviceAspect},
		{Id: airAspect},
	}
	for _, a := range aspects {
		err = helper.SetAspect(deviceManagerUrl, a)
		if err != nil {
			t.Error(err)
			return
		}
	}

	getColorFunction := devicemodel.MEASURING_FUNCTION_PREFIX + "getColorFunction"
	getHumidityFunction := devicemodel.MEASURING_FUNCTION_PREFIX + "getHumidityFunction"

	colorConcept := "urn:infai:ses:concept:color"
	humidityConcept := "urn:infai:ses:concept:humidity"

	testCharacteristic := "urn:infai:ses:characteristic:test"

	importTypes := []model.ImportType{
		{
			Id:   "lamp",
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
								AspectId:         deviceAspect,
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
			Owner: helper.JwtSubject,
		},
		{
			Id:   "never",
			Name: "never",
			Output: model.ImportContentVariable{
				Name: "output",
			},
			Owner: helper.JwtSubject,
		},
	}

	importInstances := []model.Import{
		{
			Id:           "lamp-instance",
			Name:         "lamp-instance",
			ImportTypeId: "lamp",
		},
		{
			Id:           "never-instance",
			Name:         "never-instance",
			ImportTypeId: "never",
		},
	}

	functionProducer, err := kafka.GetProducer([]string{kafkabroker}, environment.FunctionTopic)
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

	conceptProducer, err := kafka.GetProducer([]string{kafkabroker}, environment.ConceptTopic)
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

	t.Run("create import-types", testCreateImportTypes(kafkabroker, importRepoUrl, importTypes))
	t.Run("create imports", testCreateImports(importDeployUrl, importInstances))

	time.Sleep(10 * time.Second)

	criteria := model.FilterCriteriaAndSet{
		{FunctionId: getColorFunction, AspectId: deviceAspect},
	}

	t.Run("filter imports", testCheckImportSelection(ctrl, criteria, []model.Selectable{
		{
			Import: &model.Import{
				Id:           "lamp-instance",
				Name:         "lamp-instance",
				ImportTypeId: "lamp",
				Image:        "",
				KafkaTopic:   "",
				Configs:      nil,
				Restart:      nil,
			},
			ImportType: &model.ImportType{
				Id:   "lamp",
				Name: "lamp",
			},
		},
	}))

	t.Run("complete imports types", func(t *testing.T) {
		selectables := []model.Selectable{
			{
				Import: &model.Import{
					Id:           "lamp-instance",
					Name:         "lamp-instance",
					ImportTypeId: "lamp",
					Image:        "",
					KafkaTopic:   "",
					Configs:      nil,
					Restart:      nil,
				},
				ImportType: &model.ImportType{
					Id:   "lamp",
					Name: "lamp",
				},
			},
		}
		selectables, err := ctrl.CompleteServices(token, selectables, criteria)
		if err != nil {
			t.Error(err)
			return
		}
		selectables, err = normalizeImportSelectable(selectables)
		if err != nil {
			t.Error(err)
			return
		}
		expectedSelectables := []model.Selectable{
			{
				Import: &model.Import{
					Id:           "lamp-instance",
					Name:         "lamp-instance",
					ImportTypeId: "lamp",
					Image:        "",
					KafkaTopic:   "",
					Configs:      nil,
					Restart:      nil,
				},
				ImportType: &model.ImportType{
					Id:    "lamp",
					Name:  "lamp",
					Owner: helper.JwtSubject,
				},
				ServicePathOptions: map[string][]model.PathOption{
					"lamp": {{
						Path:             "value.value",
						CharacteristicId: testCharacteristic,
						AspectNode:       deviceAspectNode,
						FunctionId:       getColorFunction,
					}},
				},
			},
		}
		expectedSelectables, err = normalizeImportSelectable(expectedSelectables)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(selectables, expectedSelectables) {
			resultJson, _ := json.MarshalIndent(selectables, "", "    ")
			expectedJson, _ := json.MarshalIndent(expectedSelectables, "", "    ")
			t.Error("\n", string(resultJson), "\n", string(expectedJson))
		}
	})
}

func testCheckImportSelection(ctrl *controller.Controller, criteria model.FilterCriteriaAndSet, expectedResult []model.Selectable) func(t *testing.T) {
	return func(t *testing.T) {
		result, err, _ := ctrl.GetFilteredDevices(token, criteria, nil, "", false, true, nil)
		if err != nil {
			t.Error(err)
			return
		}
		result, err = normalizeImportSelectable(result)
		if err != nil {
			t.Error(err)
			return
		}
		expectedResult, err = normalizeImportSelectable(expectedResult)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(result, expectedResult) {
			resultJson, _ := json.MarshalIndent(result, "", "    ")
			expectedJson, _ := json.MarshalIndent(expectedResult, "", "    ")
			t.Error("\n", string(resultJson), "\n", string(expectedJson))
		}
	}
}

func testCreateImports(deployUrl string, imports []model.Import) func(t *testing.T) {
	return func(t *testing.T) {
		buff := new(bytes.Buffer)
		err := json.NewEncoder(buff).Encode(imports)
		if err != nil {
			t.Error(err)
			return
		}
		req, err := http.NewRequest("PUT", deployUrl+"/instances", buff)
		if err != nil {
			t.Error(err)
			return
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != 200 {
			temp, _ := ioutil.ReadAll(resp.Body)
			t.Error(resp.StatusCode, string(temp))
			return
		}
	}

}

func testCreateImportTypes(kafkabroker string, repoUrl string, importTypes []model.ImportType) func(t *testing.T) {
	return func(t *testing.T) {
		//ensure existing ids
		producer, err := kafka.GetProducer([]string{kafkabroker}, environment.ImportTypeTopic)
		if err != nil {
			t.Error(err)
			return
		}
		for _, importType := range importTypes {
			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			err = producer.WriteMessages(ctx, kafka2.Message{Key: []byte(importType.Id), Value: []byte(
				`{"command": "PUT", "id": "` + importType.Id + `", "owner": "` + helper.JwtSubject + `", "import_type":{"id": "` + importType.Id + `", "owner": "` + helper.JwtSubject + `"}}`)})
			if err != nil {
				t.Error(err)
				return
			}
		}
		time.Sleep(2 * time.Second)
		//set values
		for _, importType := range importTypes {
			buff := new(bytes.Buffer)
			err := json.NewEncoder(buff).Encode(importType)
			if err != nil {
				t.Error(err)
				return
			}
			req, err := http.NewRequest("PUT", repoUrl+"/import-types/"+importType.Id, buff)
			if err != nil {
				t.Error(err)
				return
			}
			req.Header.Set("Authorization", token)
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Error(err)
				return
			}
			if resp.StatusCode != 200 {
				temp, _ := ioutil.ReadAll(resp.Body)
				t.Error(resp.StatusCode, string(temp), importType.Id, importType)
				return
			}
		}
	}
}

func normalizeImportSelectable(selectable []model.Selectable) (out []model.Selectable, err error) {
	for _, v := range selectable {
		if v.ImportType != nil {
			v.ImportType.Output = model.ImportContentVariable{}
		}
	}
	tmp, err := json.Marshal(selectable)
	if err != nil {
		return []model.Selectable{}, err
	}
	err = json.Unmarshal(tmp, &out)
	return
}
