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
	"device-selection/pkg/configuration"
	"device-selection/pkg/devices"
	"device-selection/pkg/model"
	"device-selection/pkg/model/devicemodel"
	"encoding/base64"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
)

func init() {
	endpoints = append(endpoints, SelectablesEndpoints)
}

func SelectablesEndpoints(router *httprouter.Router, config configuration.Config, ctrl *devices.Devices) {

	router.GET("/selectables", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token := request.Header.Get("Authorization")
		criteria, err := getCriteriaFromRequest(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		blocklist, err, code := getBlockListFromRequest(ctrl, request, token)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		result, err, code := ctrl.GetFilteredDevices(token, criteria, blocklist)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		if request.URL.Query().Get("complete_services") == "true" {
			result, err = ctrl.CompleteServices(token, result)
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
}

func getBlockListFromRequest(ctrl *devices.Devices, request *http.Request, token string) (protocolBlockList []string, err error, code int) {
	if filterProtocols := request.URL.Query().Get("filter_protocols"); filterProtocols != "" {
		protocolBlockList = strings.Split(filterProtocols, ",")
		for i, protocol := range protocolBlockList {
			protocolBlockList[i] = strings.TrimSpace(protocol)
		}
	}

	if interaction := request.URL.Query().Get("filter_interaction"); interaction != "" {
		protocolBlockList, err, code = ctrl.GetBlockedProtocols(token, devicemodel.Interaction(interaction))
		if err != nil {
			return
		}
	}
	return
}

func getCriteriaFromRequest(request *http.Request) (descriptions model.FilterCriteriaAndSet, err error) {
	if b64 := request.URL.Query().Get("base64"); b64 != "" {
		descriptions, err = getCriteriaFromBase64(b64)
		return
	}

	if jsonStr := request.URL.Query().Get("json"); jsonStr != "" {
		err = json.Unmarshal([]byte(jsonStr), &descriptions)
		return
	}

	descriptions = []model.FilterCriteria{{
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
