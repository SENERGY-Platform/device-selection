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

package helper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/device-selection/pkg/api"
	"github.com/SENERGY-Platform/device-selection/pkg/configuration"
	"github.com/SENERGY-Platform/device-selection/pkg/controller"
	"github.com/SENERGY-Platform/device-selection/pkg/model"
	"github.com/SENERGY-Platform/device-selection/pkg/model/devicemodel"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/environment/docker"
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

func GroupHelper(selectionurl string, maintainUsability bool, deviceIds []string, expectedResult model.DeviceGroupHelperResult) func(t *testing.T) {
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
		req.Header.Set("Authorization", AdminJwt)
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
		option.Device.DisplayName = ""
		sort.SliceStable(option.RemovesCriteria, func(i, j int) bool {
			return option.RemovesCriteria[i].AspectId < option.RemovesCriteria[j].AspectId
		})
		sort.SliceStable(option.RemovesCriteria, func(i, j int) bool {
			return option.RemovesCriteria[i].FunctionId < option.RemovesCriteria[j].FunctionId
		})
		sort.SliceStable(option.RemovesCriteria, func(i, j int) bool {
			return option.RemovesCriteria[i].Interaction < option.RemovesCriteria[j].Interaction
		})
		option.Device.Permissions = model.Permissions{}
		option.Device.Creator = ""
		result.Options[i] = option
	}
	return result
}

func EnvWithDevices(ctx context.Context, wg *sync.WaitGroup, deviceTypes []devicemodel.DeviceType, deviceInstances []devicemodel.Device) (kafkaUrl string, managerurl string, repourl string, searchurl string, err error) {
	kafkaUrl, managerurl, repourl, searchurl, err = docker.DeviceManagerWithDependenciesAndKafka(ctx, wg)
	if err != nil {
		return kafkaUrl, managerurl, repourl, searchurl, err
	}

	for _, dt := range deviceTypes {
		err = SetDeviceType(managerurl, dt)
		if err != nil {
			return kafkaUrl, managerurl, repourl, searchurl, err
		}
	}

	for _, d := range deviceInstances {
		err = SetDevice(managerurl, d)
		if err != nil {
			return kafkaUrl, managerurl, repourl, searchurl, err
		}
	}

	time.Sleep(2 * time.Second)

	return
}

func EnvWithMetadata(ctx context.Context, wg *sync.WaitGroup, deviceTypes []devicemodel.DeviceType, deviceInstances []devicemodel.Device, aspects []devicemodel.Aspect, functions []devicemodel.Function) (managerurl string, repourl string, searchurl string, selectionurl string, err error) {
	var kafkaUrl string
	kafkaUrl, managerurl, repourl, searchurl, err = docker.DeviceManagerWithDependenciesAndKafka(ctx, wg)
	if err != nil {
		return managerurl, repourl, searchurl, selectionurl, err
	}

	for _, f := range functions {
		err = SetFunction(managerurl, f)
		if err != nil {
			return managerurl, repourl, searchurl, selectionurl, err
		}
	}

	for _, aspect := range aspects {
		err = SetAspect(managerurl, aspect)
		if err != nil {
			return managerurl, repourl, searchurl, selectionurl, err
		}
	}

	for _, dt := range deviceTypes {
		err = SetDeviceType(managerurl, dt)
		if err != nil {
			return managerurl, repourl, searchurl, selectionurl, err
		}
	}

	for _, d := range deviceInstances {
		err = SetDevice(managerurl, d)
		if err != nil {
			return managerurl, repourl, searchurl, selectionurl, err
		}
	}

	time.Sleep(2 * time.Second)

	c := &configuration.ConfigStruct{
		PermSearchUrl:                   searchurl,
		DeviceRepoUrl:                   repourl,
		Debug:                           true,
		KafkaUrl:                        kafkaUrl,
		KafkaConsumerGroup:              "device_selection",
		KafkaTopicsForCacheInvalidation: []string{"device-types", "aspects", "functions"},
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

func EnvWithApi(ctx context.Context, wg *sync.WaitGroup, deviceTypes []devicemodel.DeviceType, deviceInstances []devicemodel.Device) (managerurl string, repourl string, searchurl string, selectionurl string, err error) {
	var kafkaUrl string
	kafkaUrl, managerurl, repourl, searchurl, err = EnvWithDevices(ctx, wg, deviceTypes, deviceInstances)
	if err != nil {
		return managerurl, repourl, searchurl, selectionurl, err
	}

	c := &configuration.ConfigStruct{
		PermSearchUrl:                   searchurl,
		DeviceRepoUrl:                   repourl,
		Debug:                           true,
		KafkaUrl:                        kafkaUrl,
		KafkaConsumerGroup:              "device_selection",
		KafkaTopicsForCacheInvalidation: []string{"device-types", "aspects", "functions"},
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

func SetFunction(devicemanagerUrl string, f devicemodel.Function) error {
	resp, err := Jwtput(AdminJwt, devicemanagerUrl+"/functions/"+url.PathEscape(f.Id), f)
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

func SetAspect(devicemanagerUrl string, a devicemodel.Aspect) error {
	resp, err := Jwtput(AdminJwt, devicemanagerUrl+"/aspects/"+url.PathEscape(a.Id), a)
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

func SetConcept(devicemanagerUrl string, c devicemodel.Concept) error {
	resp, err := Jwtput(AdminJwt, devicemanagerUrl+"/concepts/"+url.PathEscape(c.Id), c)
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

func SetDeviceType(devicemanagerUrl string, dt devicemodel.DeviceType) error {
	temp, _ := json.Marshal(dt)
	log.Println("test create device-type:", string(temp))
	var resp *http.Response
	var err error
	if dt.Id != "" {
		resp, err = Jwtput(AdminJwt, devicemanagerUrl+"/device-types/"+url.PathEscape(dt.Id), dt)
	} else {
		resp, err = Jwtpost(AdminJwt, devicemanagerUrl+"/device-types", dt)
	}
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

func SetDevice(devicemanagerUrl string, d devicemodel.Device) error {
	d.LocalId = d.Id
	resp, err := Jwtput(AdminJwt, devicemanagerUrl+"/devices/"+url.PathEscape(d.Id), d)
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

const AdminJwt = "Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICIzaUtabW9aUHpsMmRtQnBJdS1vSkY4ZVVUZHh4OUFIckVOcG5CcHM5SjYwIn0.eyJqdGkiOiJiOGUyNGZkNy1jNjJlLTRhNWQtOTQ4ZC1mZGI2ZWVkM2JmYzYiLCJleHAiOjE1MzA1MzIwMzIsIm5iZiI6MCwiaWF0IjoxNTMwNTI4NDMyLCJpc3MiOiJodHRwczovL2F1dGguc2VwbC5pbmZhaS5vcmcvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJkZDY5ZWEwZC1mNTUzLTQzMzYtODBmMy03ZjQ1NjdmODVjN2IiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiMjJlMGVjZjgtZjhhMS00NDQ1LWFmMjctNGQ1M2JmNWQxOGI5IiwiYXV0aF90aW1lIjoxNTMwNTI4NDIzLCJzZXNzaW9uX3N0YXRlIjoiMWQ3NWE5ODQtNzM1OS00MWJlLTgxYjktNzMyZDgyNzRjMjNlIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJjcmVhdGUtcmVhbG0iLCJhZG1pbiIsImRldmVsb3BlciIsInVtYV9hdXRob3JpemF0aW9uIiwidXNlciJdfSwicmVzb3VyY2VfYWNjZXNzIjp7Im1hc3Rlci1yZWFsbSI6eyJyb2xlcyI6WyJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsInZpZXctcmVhbG0iLCJtYW5hZ2UtaWRlbnRpdHktcHJvdmlkZXJzIiwiaW1wZXJzb25hdGlvbiIsImNyZWF0ZS1jbGllbnQiLCJtYW5hZ2UtdXNlcnMiLCJxdWVyeS1yZWFsbXMiLCJ2aWV3LWF1dGhvcml6YXRpb24iLCJxdWVyeS1jbGllbnRzIiwicXVlcnktdXNlcnMiLCJtYW5hZ2UtZXZlbnRzIiwibWFuYWdlLXJlYWxtIiwidmlldy1ldmVudHMiLCJ2aWV3LXVzZXJzIiwidmlldy1jbGllbnRzIiwibWFuYWdlLWF1dGhvcml6YXRpb24iLCJtYW5hZ2UtY2xpZW50cyIsInF1ZXJ5LWdyb3VwcyJdfSwiYWNjb3VudCI6eyJyb2xlcyI6WyJtYW5hZ2UtYWNjb3VudCIsIm1hbmFnZS1hY2NvdW50LWxpbmtzIiwidmlldy1wcm9maWxlIl19fSwicm9sZXMiOlsidW1hX2F1dGhvcml6YXRpb24iLCJhZG1pbiIsImNyZWF0ZS1yZWFsbSIsImRldmVsb3BlciIsInVzZXIiLCJvZmZsaW5lX2FjY2VzcyJdLCJuYW1lIjoiZGYgZGZmZmYiLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJzZXBsIiwiZ2l2ZW5fbmFtZSI6ImRmIiwiZmFtaWx5X25hbWUiOiJkZmZmZiIsImVtYWlsIjoic2VwbEBzZXBsLmRlIn0.eOwKV7vwRrWr8GlfCPFSq5WwR_p-_rSJURXCV1K7ClBY5jqKQkCsRL2V4YhkP1uS6ECeSxF7NNOLmElVLeFyAkvgSNOUkiuIWQpMTakNKynyRfH0SrdnPSTwK2V1s1i4VjoYdyZWXKNjeT2tUUX9eCyI5qOf_Dzcai5FhGCSUeKpV0ScUj5lKrn56aamlW9IdmbFJ4VwpQg2Y843Vc0TqpjK9n_uKwuRcQd9jkKHkbwWQ-wyJEbFWXHjQ6LnM84H0CQ2fgBqPPfpQDKjGSUNaCS-jtBcbsBAWQSICwol95BuOAqVFMucx56Wm-OyQOuoQ1jaLt2t-Uxtr-C9wKJWHQ"
const JwtSubject = "dd69ea0d-f553-4336-80f3-7f4567f85c7b"

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

func TestRequest(serviceUrl string, method string, path string, body interface{}, expectedStatusCode int, expected interface{}) func(t *testing.T) {
	return TestRequestWithToken(serviceUrl, AdminJwt, method, path, body, expectedStatusCode, expected)
}

func TestRequestWithToken(serviceUrl string, token string, method string, path string, body interface{}, expectedStatusCode int, expected interface{}) func(t *testing.T) {
	return func(t *testing.T) {
		var requestBody io.Reader
		if body != nil {
			temp := new(bytes.Buffer)
			err := json.NewEncoder(temp).Encode(body)
			if err != nil {
				t.Error(err)
				return
			}
			requestBody = temp
		}

		req, err := http.NewRequest(method, serviceUrl+path, requestBody)
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
		defer resp.Body.Close()
		defer io.ReadAll(resp.Body) // ensure reuse of connection
		if resp.StatusCode != expectedStatusCode {
			temp, _ := io.ReadAll(resp.Body)
			t.Error(resp.StatusCode, string(temp))
			return
		}

		if expected != nil {
			temp, err := json.Marshal(expected)
			if err != nil {
				t.Error(err)
				return
			}
			var normalizedExpected interface{}
			err = json.Unmarshal(temp, &normalizedExpected)
			if err != nil {
				t.Error(err)
				return
			}

			var actual interface{}
			err = json.NewDecoder(resp.Body).Decode(&actual)
			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(actual, normalizedExpected) {
				a, _ := json.Marshal(actual)
				e, _ := json.Marshal(normalizedExpected)
				t.Error("\n", string(a), "\n", string(e))
				return
			}
		}
	}
}
