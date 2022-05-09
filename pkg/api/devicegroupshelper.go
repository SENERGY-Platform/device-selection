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
	"device-selection/pkg/controller"
	"device-selection/pkg/model"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
)

func init() {
	endpoints = append(endpoints, DeviceGroupsHelper)
}

func DeviceGroupsHelper(router *httprouter.Router, config configuration.Config, ctrl *controller.Controller) {
	router.POST("/device-group-helper", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token := request.Header.Get("Authorization")

		deviceIds := []string{}
		err := json.NewDecoder(request.Body).Decode(&deviceIds)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		search := model.QueryFind{}
		search.Search = request.URL.Query().Get("search")
		search.Limit, _ = strconv.Atoi(request.URL.Query().Get("limit"))
		search.Offset, _ = strconv.Atoi(request.URL.Query().Get("offset"))
		search.Rights = "rx"

		filterMaintainsGroupUsability := false
		maintainsGroupUsability := request.URL.Query().Get("maintains_group_usability")
		if maintainsGroupUsability != "" {
			filterMaintainsGroupUsability, _ = strconv.ParseBool(maintainsGroupUsability)
		}

		functionBlockList := []string{}
		functionBlockListStr := request.URL.Query().Get("function_block_list")
		if functionBlockListStr != "" {
			functionBlockList = strings.Split(functionBlockListStr, ",")
		}

		result, err, code := ctrl.DeviceGroupHelper(token, deviceIds, search, filterMaintainsGroupUsability, functionBlockList)
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
