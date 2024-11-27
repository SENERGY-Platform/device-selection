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

package _807

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-selection/pkg/configuration"
	"github.com/SENERGY-Platform/device-selection/pkg/controller"
	"github.com/SENERGY-Platform/device-selection/pkg/model"
	"github.com/SENERGY-Platform/device-selection/pkg/model/devicemodel"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/environment/legacy"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/helper"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestGetFilteredDeviceTypes(t *testing.T) {

	mux := sync.Mutex{}
	calls := []string{}

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.Lock()
		defer mux.Unlock()
		calls = append(calls, r.URL.Path+"?"+r.URL.RawQuery)
		json.NewEncoder(w).Encode([]devicemodel.DeviceType{{Id: "dt1", Name: "dt1name"}})
	}))

	defer mock.Close()

	c := &configuration.ConfigStruct{
		DeviceRepoUrl: mock.URL,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	repo, err := controller.New(ctx, c)
	if err != nil {
		t.Error(err)
		return
	}

	_, err, _ = repo.GetFilteredDeviceTypes(helper.AdminJwt, []client.FilterCriteria{{
		FunctionId: "fid",
	}})

	if err != nil {
		t.Error(err)
		return
	}

	dt, err, _ := repo.GetFilteredDeviceTypes(helper.AdminJwt, []client.FilterCriteria{{
		FunctionId:    "fid",
		DeviceClassId: "dc1",
		AspectId:      "a1",
	}})

	if err != nil {
		t.Error(err)
		return
	}

	if len(dt) != 1 || dt[0].Name != "dt1name" || dt[0].Id != "dt1" {
		t.Error(dt)
		return
	}

	mux.Lock()
	defer mux.Unlock()
	expected := []string{
		"/v3/device-types?criteria=" + url.QueryEscape(`[{"interaction":"","function_id":"fid","device_class_id":"","aspect_id":""}]`) + "&include-modified=true&limit=10000&sort=name.asc",
		"/v3/device-types?criteria=" + url.QueryEscape(`[{"interaction":"","function_id":"fid","device_class_id":"dc1","aspect_id":"a1"}]`) + "&include-modified=true&limit=10000&sort=name.asc",
	}
	if !reflect.DeepEqual(calls, expected) {
		t.Errorf("\na=%#v\ne=%#v\n", calls, expected)
		for _, path := range calls {
			parts := strings.SplitN(path, "?", 2)
			if len(parts) != 2 {
				t.Error("unexpected path", parts)
				return
			}
			query, err := url.ParseQuery(parts[1])
			if err != nil {
				t.Error(err)
				return
			}
			t.Logf("%#v\n", query)
		}
	}
}

func TestGetFilteredDevices(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	managerurl, repourl, _, err := legacy.TestenvWithoutApi(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	for _, concept := range concepts {
		err = helper.SetConcept(managerurl, concept)
		if err != nil {
			t.Error(err)
			return
		}
	}

	for _, f := range functions {
		err = helper.SetFunction(managerurl, f)
		if err != nil {
			t.Error(err)
			return
		}
	}

	for _, a := range aspects {
		err = helper.SetAspect(managerurl, a)
		if err != nil {
			t.Error(err)
			return
		}
	}

	c := &configuration.ConfigStruct{
		DeviceRepoUrl: repourl,
	}

	repo, err := controller.New(ctx, c)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(2 * time.Second)

	d, err, _ := repo.GetFilteredDevices(helper.AdminJwt, DeviceDescriptions{{
		CharacteristicId: "chid1",
		Function:         devicemodel.Function{Id: devicemodel.MEASURING_FUNCTION_PREFIX + "_1"},
		DeviceClass:      &devicemodel.DeviceClass{Id: "dc1"},
		Aspect:           &devicemodel.Aspect{Id: "a1"},
	}}.ToFilter(), []string{"mqtt"}, "", false, false, nil)

	if err != nil {
		t.Error(err)
		return
	}

	if len(d) != 1 || d[0].Device.Name != "1" || d[0].Device.Id != "1" || len(d[0].Services) != 1 || d[0].Services[0].Id != "11" {
		t.Error(len(d), d)
		return
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
