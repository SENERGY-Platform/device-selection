/*
 * Copyright 2026 InfAI (CC SES)
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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/SENERGY-Platform/device-selection/pkg/model"
	"github.com/SENERGY-Platform/device-selection/pkg/model/devicemodel"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/helper"
)

// The aspects of one filter criteria are ANDed on the content variable: the variable has to
// carry all of them, so a criteria over two aspects finds only the device whose variable
// carries both, not the one that carries them in two separate variables.
func TestSelectablesAspectIds(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	getTemperature := devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature"
	air := "urn:infai:ses:aspect:air"
	insideAir := "urn:infai:ses:aspect:inside_air"
	water := "urn:infai:ses:aspect:water"

	aspects := []devicemodel.Aspect{
		{Id: air, Name: "air", SubAspects: []devicemodel.Aspect{{Id: insideAir, Name: "inside_air"}}},
		{Id: water, Name: "water"},
	}
	functions := []devicemodel.Function{{Id: getTemperature}}

	temperatureVariable := func(id string, aspectIds []string) devicemodel.ContentVariable {
		return devicemodel.ContentVariable{
			Id:               id,
			Name:             "temperature",
			FunctionId:       getTemperature,
			AspectIds:        aspectIds,
			CharacteristicId: "temperature-characteristic",
		}
	}

	deviceTypes := []devicemodel.DeviceType{
		{
			//one variable carrying both aspects
			Id:            "combined",
			Name:          "combined",
			DeviceClassId: "thermometer",
			Services: []devicemodel.Service{{
				Id:          "combined_service",
				Name:        "combined_service",
				Interaction: devicemodel.REQUEST,
				Outputs: []devicemodel.Content{{
					Id:              "combined_output",
					ContentVariable: temperatureVariable("combined_variable", []string{insideAir, water}),
				}},
			}},
		},
		{
			//the same two aspects, but in two separate variables
			Id:            "separate",
			Name:          "separate",
			DeviceClassId: "thermometer",
			Services: []devicemodel.Service{{
				Id:          "separate_service",
				Name:        "separate_service",
				Interaction: devicemodel.REQUEST,
				Outputs: []devicemodel.Content{
					{
						Id:              "separate_output_air",
						ContentVariable: temperatureVariable("separate_variable_air", []string{insideAir}),
					},
					{
						Id:              "separate_output_water",
						ContentVariable: temperatureVariable("separate_variable_water", []string{water}),
					},
				},
			}},
		},
	}

	devices := []devicemodel.Device{
		{Id: "combined_device", LocalId: "combined_device", Name: "combined_device", DeviceTypeId: "combined", OwnerId: helper.JwtSubject},
		{Id: "separate_device", LocalId: "separate_device", Name: "separate_device", DeviceTypeId: "separate", OwnerId: helper.JwtSubject},
	}

	_, _, _, selectionurl, err := helper.EnvWithMetadata(ctx, wg, deviceTypes, devices, aspects, functions)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("single aspect", testAspectIdsSelection(selectionurl, model.FilterCriteriaAndSet{
		{FunctionId: getTemperature, AspectIds: []string{insideAir}},
	}, []string{"combined_device", "separate_device"}))

	t.Run("deprecated alias", testAspectIdsSelection(selectionurl, model.FilterCriteriaAndSet{
		{FunctionId: getTemperature, AspectId: insideAir},
	}, []string{"combined_device", "separate_device"}))

	t.Run("both aspects on one variable", testAspectIdsSelection(selectionurl, model.FilterCriteriaAndSet{
		{FunctionId: getTemperature, AspectIds: []string{insideAir, water}},
	}, []string{"combined_device"}))

	//every entry of the list covers its own subtree, so the parent of the variable aspect matches it
	t.Run("aspect list covers the subtree", testAspectIdsSelection(selectionurl, model.FilterCriteriaAndSet{
		{FunctionId: getTemperature, AspectIds: []string{air, water}},
	}, []string{"combined_device"}))

	t.Run("aspect the variable does not carry", testAspectIdsSelection(selectionurl, model.FilterCriteriaAndSet{
		{FunctionId: getTemperature, AspectIds: []string{water, "urn:infai:ses:aspect:unknown"}},
	}, []string{}))

	//the query parameter shortcut of the single criteria form
	t.Run("aspect_ids query parameter", func(t *testing.T) {
		actual, err := getSelectableDeviceIds(selectionurl, "/v2/selectables?include_devices=true&function_id="+url.QueryEscape(getTemperature)+
			"&aspect_ids="+url.QueryEscape(insideAir+","+water))
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual, []string{"combined_device"}) {
			t.Error(actual)
		}
	})
}

func testAspectIdsSelection(selectionurl string, criteria model.FilterCriteriaAndSet, expectedDeviceIds []string) func(t *testing.T) {
	return func(t *testing.T) {
		criteriaJson, err := json.Marshal(criteria)
		if err != nil {
			t.Error(err)
			return
		}
		actual, err := getSelectableDeviceIds(selectionurl, "/v2/selectables?include_devices=true&json="+url.QueryEscape(string(criteriaJson)))
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual, expectedDeviceIds) {
			t.Error(actual, expectedDeviceIds)
		}
	}
}

func getSelectableDeviceIds(selectionurl string, path string) (deviceIds []string, err error) {
	req, err := http.NewRequest("GET", selectionurl+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", helper.AdminJwt)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		temp, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected statuscode %v: %v", resp.StatusCode, string(temp))
	}
	result := []model.Selectable{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	deviceIds = []string{}
	for _, selectable := range result {
		if selectable.Device != nil {
			deviceIds = append(deviceIds, selectable.Device.Id)
		}
	}
	sort.Strings(deviceIds)
	return deviceIds, nil
}
