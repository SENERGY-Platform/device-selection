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
	"device-selection/pkg/configuration"
	"device-selection/pkg/controller"
	"device-selection/pkg/model"
	"device-selection/pkg/model/devicemodel"
	"device-selection/pkg/tests/environment"
	"encoding/json"
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

	_, semanticUrl, deviceRepoUrl, permSearchUrl, importRepoUrl, importDeployUrl, err := environment.New(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	c := &configuration.ConfigStruct{
		SemanticRepoUrl: semanticUrl,
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
	airAspect := "urn:infai:ses:aspect:airAspect"

	getColorFunction := devicemodel.MEASURING_FUNCTION_PREFIX + "getColorFunction"
	getHumidityFunction := devicemodel.MEASURING_FUNCTION_PREFIX + "getHumidityFunction"

	importTypes := []model.ImportType{
		{
			Id:          "lamp",
			Name:        "lamp",
			AspectIds:   []string{deviceAspect, airAspect},
			FunctionIds: []string{getColorFunction, getHumidityFunction},
			Output: model.ImportContentVariable{
				Name: "output",
			},
			Owner: "1234567890",
		},
		{
			Id:          "never",
			Name:        "never",
			AspectIds:   []string{},
			FunctionIds: []string{},
			Output: model.ImportContentVariable{
				Name: "output",
			},
			Owner: "1234567890",
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

	t.Run("create import-types", testCreateImportTypes(importRepoUrl, importTypes))
	t.Run("create imports", testCreateImports(importDeployUrl, importInstances))

	time.Sleep(10 * time.Second)

	t.Run("filter imports", testCheckImportSelection(ctrl, model.FilterCriteriaAndSet{
		{FunctionId: getColorFunction, AspectId: deviceAspect},
	}, []model.Selectable{
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
				Id:          "lamp",
				Name:        "lamp",
				AspectIds:   []string{deviceAspect, airAspect},
				FunctionIds: []string{getColorFunction, getHumidityFunction},
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
					Id:          "lamp",
					Name:        "lamp",
					AspectIds:   []string{deviceAspect, airAspect},
					FunctionIds: []string{getColorFunction, getHumidityFunction},
					Output: model.ImportContentVariable{
						Name: "output",
					},
				},
			},
		}
		selectables, err := ctrl.CompleteServices(token, selectables)
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
					Id:          "lamp",
					Name:        "lamp",
					AspectIds:   []string{deviceAspect, airAspect},
					FunctionIds: []string{getColorFunction, getHumidityFunction},
					Output: model.ImportContentVariable{
						Name: "output",
					},
					Owner: "1234567890",
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
		result, err, _ := ctrl.GetFilteredDevices(token, criteria, nil, "", false, true)
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

func testCreateImportTypes(repoUrl string, importTypes []model.ImportType) func(t *testing.T) {
	return func(t *testing.T) {
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
}

func normalizeImportSelectable(selectable []model.Selectable) (out []model.Selectable, err error) {
	tmp, err := json.Marshal(selectable)
	if err != nil {
		return []model.Selectable{}, err
	}
	err = json.Unmarshal(tmp, &out)
	return
}
