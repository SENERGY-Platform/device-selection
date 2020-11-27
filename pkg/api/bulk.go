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
)

func init() {
	endpoints = append(endpoints, BulkEndpoints)
}

func BulkEndpoints(router *httprouter.Router, config configuration.Config, ctrl *controller.Controller) {

	router.POST("/bulk/selectables", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token := request.Header.Get("Authorization")

		criteria := model.BulkRequest{}
		err := json.NewDecoder(request.Body).Decode(&criteria)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		if config.Debug {
			temp, _ := json.Marshal(criteria)
			log.Println("DEBUG:", string(temp))
		}

		result, err, code := ctrl.BulkGetFilteredDevices(token, criteria)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		if request.URL.Query().Get("complete_services") == "true" {
			result, err = ctrl.CompleteBulkServices(token, result)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if config.Debug {
			temp, _ := json.Marshal(result)
			log.Println("DEBUG:", string(temp))
		}

		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			log.Println("ERROR:", err)
			debug.PrintStack()
		}
	})

	router.POST("/bulk/selectables/combined/devices", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token := request.Header.Get("Authorization")
		criteria := model.BulkRequest{}
		err := json.NewDecoder(request.Body).Decode(&criteria)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		for _, element := range criteria {
			if element.IncludeGroups {
				http.Error(writer, "unable to combine devices when groups are expected (fix: set include_groups to false)", http.StatusBadRequest)
				return
			}
		}
		temp, err, code := ctrl.BulkGetFilteredDevices(token, criteria)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		result := ctrl.CombinedDevices(temp)
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			log.Println("ERROR:", err)
			debug.PrintStack()
		}
	})

}
