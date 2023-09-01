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

package api

import (
	"encoding/base64"
	"encoding/json"
	"github.com/SENERGY-Platform/device-selection/pkg/configuration"
	"github.com/SENERGY-Platform/device-selection/pkg/controller"
	"github.com/SENERGY-Platform/device-selection/pkg/model"
	"github.com/SENERGY-Platform/device-selection/pkg/model/devicemodel"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
)

func init() {
	endpoints = append(endpoints, SelectablesEndpoints)
}

func SelectablesEndpoints(router *httprouter.Router, config configuration.Config, ctrl *controller.Controller) {

	router.GET("/selectables", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token := request.Header.Get("Authorization")
		criteria, blockedProtocols, blockedInteraction, err := getCriteriaFromRequest(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		includeGroups, _ := strconv.ParseBool(request.URL.Query().Get("include_groups"))
		includeImports, _ := strconv.ParseBool(request.URL.Query().Get("include_imports"))

		var withLocalDeviceIds []string
		localDevicesQueryParam := request.URL.Query().Get("local_devices")
		if localDevicesQueryParam != "" {
			for _, localId := range strings.Split(localDevicesQueryParam, ",") {
				withLocalDeviceIds = append(withLocalDeviceIds, strings.TrimSpace(localId))
			}
		}

		result, err, code := ctrl.GetFilteredDevices(token, criteria, blockedProtocols, blockedInteraction, includeGroups, includeImports, withLocalDeviceIds)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		if request.URL.Query().Get("complete_services") == "true" {
			result, err = ctrl.CompleteServices(token, result, criteria)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			log.Println("ERROR:", err)
			debug.PrintStack()
		}
	})

	router.GET("/v2/selectables", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token := request.Header.Get("Authorization")
		criteria, err := getCriteriaFromRequestV2(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		includeGroups, _ := strconv.ParseBool(request.URL.Query().Get("include_groups"))
		includeImports, _ := strconv.ParseBool(request.URL.Query().Get("include_imports"))
		includeDevices, _ := strconv.ParseBool(request.URL.Query().Get("include_devices"))
		includeIdModified, _ := strconv.ParseBool(request.URL.Query().Get("include_id_modified"))

		var withLocalDeviceIds []string
		localDevicesQueryParam := request.URL.Query().Get("local_devices")
		if localDevicesQueryParam != "" {
			for _, localId := range strings.Split(localDevicesQueryParam, ",") {
				withLocalDeviceIds = append(withLocalDeviceIds, strings.TrimSpace(localId))
			}
		}

		result, err, code := ctrl.GetFilteredDevicesV2(token, criteria, includeDevices, includeGroups, includeImports, withLocalDeviceIds, includeIdModified)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			log.Println("ERROR:", err)
			debug.PrintStack()
		}
	})

	router.POST("/v2/query/selectables", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token := request.Header.Get("Authorization")

		var criteria model.FilterCriteriaAndSet
		err := json.NewDecoder(request.Body).Decode(&criteria)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		includeGroups, _ := strconv.ParseBool(request.URL.Query().Get("include_groups"))
		includeImports, _ := strconv.ParseBool(request.URL.Query().Get("include_imports"))
		includeDevices, _ := strconv.ParseBool(request.URL.Query().Get("include_devices"))
		includeIdModified, _ := strconv.ParseBool(request.URL.Query().Get("include_id_modified"))

		var withLocalDeviceIds []string
		localDevicesQueryParam := request.URL.Query().Get("local_devices")
		if localDevicesQueryParam != "" {
			for _, localId := range strings.Split(localDevicesQueryParam, ",") {
				withLocalDeviceIds = append(withLocalDeviceIds, strings.TrimSpace(localId))
			}
		}

		result, err, code := ctrl.GetFilteredDevicesV2(token, criteria, includeDevices, includeGroups, includeImports, withLocalDeviceIds, includeIdModified)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			log.Println("ERROR:", err)
			debug.PrintStack()
		}
	})
}

func getCriteriaFromRequest(request *http.Request) (criteria model.FilterCriteriaAndSet, protocolBlockList []string, blockedInteraction devicemodel.Interaction, err error) {
	if filterProtocols := request.URL.Query().Get("filter_protocols"); filterProtocols != "" {
		protocolBlockList = strings.Split(filterProtocols, ",")
		for i, protocol := range protocolBlockList {
			protocolBlockList[i] = strings.TrimSpace(protocol)
		}
	}

	if interaction := request.URL.Query().Get("filter_interaction"); interaction != "" {
		blockedInteraction = devicemodel.Interaction(interaction)
	}

	if b64 := request.URL.Query().Get("base64"); b64 != "" {
		criteria, err = getCriteriaFromBase64(b64)
		return
	}

	if jsonStr := request.URL.Query().Get("json"); jsonStr != "" {
		err = json.Unmarshal([]byte(jsonStr), &criteria)
		return
	}

	criteria = []devicemodel.FilterCriteria{{
		FunctionId:    request.URL.Query().Get("function_id"),
		DeviceClassId: request.URL.Query().Get("device_class_id"),
		AspectId:      request.URL.Query().Get("aspect_id"),
	}}
	return
}

func getCriteriaFromRequestV2(request *http.Request) (criteria model.FilterCriteriaAndSet, err error) {
	if b64 := request.URL.Query().Get("base64"); b64 != "" {
		criteria, err = getCriteriaFromBase64(b64)
		return
	}

	if jsonStr := request.URL.Query().Get("json"); jsonStr != "" {
		err = json.Unmarshal([]byte(jsonStr), &criteria)
		return
	}

	criteria = []devicemodel.FilterCriteria{{
		Interaction:   request.URL.Query().Get("interaction"),
		FunctionId:    request.URL.Query().Get("function_id"),
		DeviceClassId: request.URL.Query().Get("device_class_id"),
		AspectId:      request.URL.Query().Get("aspect_id"),
	}}
	return
}

func getCriteriaFromBase64(b64 string) (descriptions model.FilterCriteriaAndSet, err error) {
	var jsonByte []byte
	jsonByte, err = base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return
	}

	err = json.Unmarshal(jsonByte, &descriptions)
	if err != nil {
		return
	}

	return
}
