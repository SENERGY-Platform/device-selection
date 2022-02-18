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
	"context"
	"device-selection/pkg/model"
	"device-selection/pkg/model/devicemodel"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"runtime/debug"
	"sync"
	"testing"
)

func TestApiSimpleGet(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, _, selectionurl, err := testenv(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	result := []model.Selectable{}

	t.Run("send simple request", sendSimpleRequest(selectionurl, &result, devicemodel.MEASURING_FUNCTION_PREFIX+"_1", "dc1", "a1", "mqtt"))

	t.Run("check result", func(t *testing.T) {
		if len(result) != 1 {
			t.Error(len(result), result)
			return
		}
		if result[0].Device.Name != "1" ||
			result[0].Device.Id != "1" ||
			len(result[0].Services) != 1 ||
			result[0].Services[0].Id != "11" {
			temp, _ := json.Marshal(result[0])
			t.Error(string(temp))
			return
		}
	})
}

func TestApiCompleteSimpledGet(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, _, selectionurl, err := testenv(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	result := []model.Selectable{}

	t.Run("send simple request", sendCompletedSimpleRequest(selectionurl, &result, devicemodel.MEASURING_FUNCTION_PREFIX+"_1", "dc1", "a1", "mqtt"))

	t.Run("check result", func(t *testing.T) {
		if len(result) != 1 ||
			result[0].Device.Name != "1" ||
			result[0].Device.Id != "1" ||
			len(result[0].Services) != 1 ||
			result[0].Services[0].Id != "11" ||
			len(result[0].Services[0].Outputs) != 1 ||
			result[0].Services[0].Outputs[0].Id != "content1" {
			t.Error(result)
			return
		}
	})
}

func TestApiJsonGet(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, _, selectionurl, err := testenv(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	result := []model.Selectable{}

	t.Run("send json request", sendJsonRequest(selectionurl, &result, devicemodel.MEASURING_FUNCTION_PREFIX+"_1", "dc1", "a1", "mqtt"))

	t.Run("check result", func(t *testing.T) {
		if len(result) != 1 ||
			result[0].Device.Name != "1" ||
			result[0].Device.Id != "1" ||
			len(result[0].Services) != 1 ||
			result[0].Services[0].Id != "11" {
			t.Error(result)
			return
		}
	})

}

func TestApiBase64Get(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, _, selectionurl, err := testenv(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	result := []model.Selectable{}

	t.Run("send base64 request", sendBase64Request(selectionurl, &result, devicemodel.MEASURING_FUNCTION_PREFIX+"_1", "dc1", "a1", "mqtt"))

	t.Run("check result", func(t *testing.T) {
		if len(result) != 1 ||
			result[0].Device.Name != "1" ||
			result[0].Device.Id != "1" ||
			len(result[0].Services) != 1 ||
			result[0].Services[0].Id != "11" {
			t.Error(result)
			return
		}
	})
}

func sendSimpleRequest(apiurl string, result interface{}, functionId string, deviceClassId string, aspectId string, blockList string) func(t *testing.T) {
	return func(t *testing.T) {
		endpoint := apiurl + "/selectables?function_id=" + url.QueryEscape(functionId) + "&device_class_id=" + url.QueryEscape(deviceClassId) + "&aspect_id=" + url.QueryEscape(aspectId) + "&filter_protocols=" + url.QueryEscape(blockList)
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", adminjwt)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != 200 {
			t.Error(resp.StatusCode)
			return
		}
		err = json.NewDecoder(resp.Body).Decode(result)
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func sendCompletedSimpleRequest(apiurl string, result interface{}, functionId string, deviceClassId string, aspectId string, blockList string) func(t *testing.T) {
	return func(t *testing.T) {
		endpoint := apiurl + "/selectables?complete_services=true&function_id=" + url.QueryEscape(functionId) + "&device_class_id=" + url.QueryEscape(deviceClassId) + "&aspect_id=" + url.QueryEscape(aspectId) + "&filter_protocols=" + url.QueryEscape(blockList)
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", adminjwt)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != 200 {
			t.Error(resp.StatusCode)
			return
		}
		err = json.NewDecoder(resp.Body).Decode(result)
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func sendJsonRequest(apiurl string, result interface{}, functionId string, deviceClassId string, aspectId string, blockList string) func(t *testing.T) {
	return func(t *testing.T) {
		jsonStr, err := json.Marshal(model.FilterCriteriaAndSet{{
			FunctionId:    functionId,
			DeviceClassId: deviceClassId,
			AspectId:      aspectId,
		}})
		if err != nil {
			t.Error(err)
			return
		}
		endpoint := apiurl + "/selectables?json=" + url.QueryEscape(string(jsonStr)) + "&filter_protocols=" + url.QueryEscape(blockList)
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", adminjwt)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != 200 {
			t.Error(resp.StatusCode)
			return
		}
		err = json.NewDecoder(resp.Body).Decode(result)
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func sendBase64Request(apiurl string, result interface{}, functionId string, deviceClassId string, aspectId string, blockList string) func(t *testing.T) {
	return func(t *testing.T) {
		jsonStr, err := json.Marshal(model.FilterCriteriaAndSet{{
			FunctionId:    functionId,
			DeviceClassId: deviceClassId,
			AspectId:      aspectId,
		}})
		if err != nil {
			t.Error(err)
			return
		}
		b64Str := base64.StdEncoding.EncodeToString(jsonStr)
		resp, err := http.Get(apiurl + "/selectables?base64=" + b64Str + "&filter_protocols=" + url.QueryEscape(blockList))
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != 200 {
			t.Error(resp.StatusCode)
			return
		}
		err = json.NewDecoder(resp.Body).Decode(result)
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func testenv(ctx context.Context, wg *sync.WaitGroup) (managerurl string, repourl string, searchurl string, selectionurl string, err error) {
	deviceTypes := []devicemodel.DeviceType{
		{Id: "dt1", Name: "dt1name", DeviceClassId: "dc1", Services: []devicemodel.Service{
			testTechnicalService("11", "pid", nil, []devicemodel.Content{{
				Id:            "content1",
				Serialization: "json",
				ContentVariable: devicemodel.ContentVariable{
					Id:         "variable1",
					Name:       "variable1",
					AspectId:   "a1",
					FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "_1",
				},
			}}, devicemodel.REQUEST),
			testTechnicalService("11_b", "mqtt", nil, []devicemodel.Content{{
				Id:            "content2",
				Serialization: "json",
				ContentVariable: devicemodel.ContentVariable{
					Id:         "variable2",
					Name:       "variable2",
					AspectId:   "a1",
					FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "_1",
				},
			}}, devicemodel.EVENT),
			testTechnicalService("12", "pid", []devicemodel.Content{{
				Id:            "content3",
				Serialization: "json",
				ContentVariable: devicemodel.ContentVariable{
					Id:         "variable3",
					Name:       "variable3",
					AspectId:   "a1",
					FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "_1",
				},
			}}, nil, devicemodel.REQUEST),
		}},
		{Id: "dt2", Name: "dt2name", DeviceClassId: "dc1", Services: []devicemodel.Service{
			testTechnicalService("21", "pid", []devicemodel.Content{{
				Id:            "content4",
				Serialization: "json",
				ContentVariable: devicemodel.ContentVariable{
					Id:         "variable4",
					Name:       "variable4",
					AspectId:   "a1",
					FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "_1",
				},
			}}, nil, devicemodel.REQUEST),
			testTechnicalService("22", "pid", []devicemodel.Content{{
				Id:            "content5",
				Serialization: "json",
				ContentVariable: devicemodel.ContentVariable{
					Id:         "variable5",
					Name:       "variable5",
					AspectId:   "a1",
					FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "_1",
				},
			}}, nil, devicemodel.REQUEST),
		}},
		{Id: "dt3", Name: "dt1name", DeviceClassId: "dc1", Services: []devicemodel.Service{
			testService("31", "mqtt", devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION, devicemodel.EVENT),
			testService("32", "mqtt", devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION, devicemodel.EVENT),
		}},
		{Id: "dt4", Name: "dt2name", DeviceClassId: "dc1", Services: []devicemodel.Service{
			testService("41", "mqtt", devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION, devicemodel.EVENT),
			testService("42", "mqtt", devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION, devicemodel.EVENT),
		}},
	}
	deviceInstances := []devicemodel.Device{
		{
			Id:           "1",
			Name:         "1",
			DeviceTypeId: "dt1",
		},
		{
			Id:           "2",
			Name:         "2",
			DeviceTypeId: "dt2",
		},
		{
			Id:           "3",
			Name:         "3",
			DeviceTypeId: "dt3",
		},
		{
			Id:           "4",
			Name:         "4",
			DeviceTypeId: "dt4",
		},
	}

	concepts := []devicemodel.Concept{
		{
			Id: "concept",
		},
	}

	functions := []devicemodel.Function{
		{
			Id:        devicemodel.MEASURING_FUNCTION_PREFIX + "_1",
			ConceptId: "concept",
		},
		{
			Id:        devicemodel.CONTROLLING_FUNCTION_PREFIX + "_1",
			ConceptId: "concept",
		},
	}

	aspects := []devicemodel.Aspect{
		{
			Id:   "a1",
			Name: "a1",
		},
	}

	managerurl, repourl, searchurl, selectionurl, err = grouphelpertestenv(ctx, wg, deviceTypes, deviceInstances)
	if err != nil {
		return managerurl, repourl, searchurl, selectionurl, err
	}

	for _, concept := range concepts {
		err = testSetConcept(managerurl, concept)
		if err != nil {
			return managerurl, repourl, searchurl, selectionurl, err
		}
	}

	for _, f := range functions {
		err = testSetFunction(managerurl, f)
		if err != nil {
			return managerurl, repourl, searchurl, selectionurl, err
		}
	}

	for _, a := range aspects {
		err = testSetAspect(managerurl, a)
		if err != nil {
			return managerurl, repourl, searchurl, selectionurl, err
		}
	}
	return
}

func testService(id string, protocolId string, functionType string, interaction devicemodel.Interaction) devicemodel.Service {
	result := devicemodel.Service{
		Id:          id,
		LocalId:     id + "_l",
		Name:        id + "_name",
		ProtocolId:  protocolId,
		Interaction: interaction,
	}
	if functionType == devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION {
		result.Outputs = append(result.Outputs, devicemodel.Content{
			ContentVariable: devicemodel.ContentVariable{
				FunctionId: devicemodel.MEASURING_FUNCTION_PREFIX + "_1",
				AspectId:   "a1",
			},
			Serialization:     "json",
			ProtocolSegmentId: "ProtocolSegmentId",
		})
	} else {
		result.Inputs = append(result.Outputs, devicemodel.Content{
			ContentVariable: devicemodel.ContentVariable{
				FunctionId: devicemodel.CONTROLLING_FUNCTION_PREFIX + "_1",
				AspectId:   "a1",
			},
			Serialization:     "json",
			ProtocolSegmentId: "ProtocolSegmentId",
		})
	}
	return result
}

func testTechnicalService(id string, protocolId string, inputs, outputs []devicemodel.Content, interaction devicemodel.Interaction) devicemodel.Service {
	return devicemodel.Service{
		Id:          id,
		LocalId:     id + "_l",
		Name:        id + "_name",
		ProtocolId:  protocolId,
		Outputs:     outputs,
		Inputs:      inputs,
		Interaction: interaction,
	}
}

type DeviceDescriptions []DeviceDescription
type DeviceDescription struct {
	CharacteristicId string                   `json:"characteristic_id"`
	Function         devicemodel.Function     `json:"function"`
	DeviceClass      *devicemodel.DeviceClass `json:"device_class,omitempty"`
	Aspect           *devicemodel.Aspect      `json:"aspect,omitempty"`
}

func (this DeviceDescriptions) ToFilter() (result model.FilterCriteriaAndSet) {
	for _, element := range this {
		newElement := devicemodel.FilterCriteria{
			FunctionId: element.Function.Id,
		}
		if element.DeviceClass != nil {
			newElement.DeviceClassId = element.DeviceClass.Id
		}
		if element.Aspect != nil {
			newElement.AspectId = element.Aspect.Id
		}
		if !IsZero(element) {
			result = append(result, newElement)
		}
	}
	return
}

func IsZero(x interface{}) bool {
	return x == reflect.Zero(reflect.TypeOf(x)).Interface()
}

type TestPermSearchDevice struct {
	Id          string            `json:"id"`
	LocalId     string            `json:"local_id,omitempty"`
	Name        string            `json:"name,omitempty"`
	DeviceType  string            `json:"device_type_id,omitempty"`
	Permissions model.Permissions `json:"permissions"`
	Shared      bool              `json:"shared"`
	Creator     string            `json:"creator"`
}

func testSetFunction(devicemanagerUrl string, f devicemodel.Function) error {
	resp, err := Jwtput(adminjwt, devicemanagerUrl+"/functions/"+url.PathEscape(f.Id), f)
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

func testSetAspect(devicemanagerUrl string, a devicemodel.Aspect) error {
	resp, err := Jwtput(adminjwt, devicemanagerUrl+"/aspects/"+url.PathEscape(a.Id), a)
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

func testSetConcept(devicemanagerUrl string, c devicemodel.Concept) error {
	resp, err := Jwtput(adminjwt, devicemanagerUrl+"/concepts/"+url.PathEscape(c.Id), c)
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
