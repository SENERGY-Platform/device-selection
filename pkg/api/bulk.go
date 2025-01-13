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
	"encoding/json"
	"github.com/SENERGY-Platform/device-selection/pkg/configuration"
	"github.com/SENERGY-Platform/device-selection/pkg/controller"
	"github.com/SENERGY-Platform/device-selection/pkg/model"
	"log"
	"net/http"
	"runtime/debug"
)

func init() {
	endpoints = append(endpoints, &BulkEndpoints{})
}

type BulkEndpoints struct{}

// SelectablesV2 godoc
// @Summary      bulk selectables v2
// @Description  bulk selectables v2
// @Tags         bulk, selectables
// @Accept       json
// @Produce      json
// @Security Bearer
// @Param        message body model.BulkRequestV2 true "BulkRequestV2"
// @Param        complete_services query bool false "adds full import-type and import path options to the result. device services are already complete, the name is a legacy artefact"
// @Success      200 {array}  model.BulkResult
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/bulk/selectables [POST]
func (this *BulkEndpoints) SelectablesV2(router *http.ServeMux, config configuration.Config, ctrl *controller.Controller) {
	router.HandleFunc("POST /v2/bulk/selectables", func(writer http.ResponseWriter, request *http.Request) {
		token := request.Header.Get("Authorization")

		criteria := model.BulkRequestV2{}
		err := json.NewDecoder(request.Body).Decode(&criteria)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		if config.Debug {
			temp, _ := json.Marshal(criteria)
			log.Println("DEBUG:", string(temp))
		}

		result, err, code := ctrl.BulkGetFilteredDevicesV2(token, criteria)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		if request.URL.Query().Get("complete_services") == "true" {
			result, err = ctrl.CompleteBulkServicesV2(token, result, criteria)
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
}

// Selectables godoc
// @Summary      deprecated bulk selectables
// @Description  deprecated bulk selectables
// @Tags         bulk, selectables, deprecated
// @Accept       json
// @Produce      json
// @Security Bearer
// @Param        message body model.BulkRequest true "BulkRequest"
// @Param        complete_services query bool false "adds full import-type and import path options to the result. device services are already complete, the name is a legacy artefact"
// @Success      200 {array}  model.BulkResult
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /bulk/selectables [POST]
func (this *BulkEndpoints) Selectables(router *http.ServeMux, config configuration.Config, ctrl *controller.Controller) {
	router.HandleFunc("POST /bulk/selectables", func(writer http.ResponseWriter, request *http.Request) {
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
			result, err = ctrl.CompleteBulkServices(token, result, criteria)
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
}

// SelectablesCombinedDevices godoc
// @Summary      bulk selectables combined devices
// @Description  returns a list of devices, that fulfill any element of the bulk-request list; include_groups and include_imports must be false
// @Tags         bulk, selectables, devices
// @Accept       json
// @Produce      json
// @Security Bearer
// @Param        message body model.BulkRequest true "BulkRequest"
// @Success      200 {array}  []model.PermSearchDevice
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /bulk/selectables/combined/devices [POST]
func (this *BulkEndpoints) SelectablesCombinedDevices(router *http.ServeMux, config configuration.Config, ctrl *controller.Controller) {
	router.HandleFunc("POST /bulk/selectables/combined/devices", func(writer http.ResponseWriter, request *http.Request) {
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
			if element.IncludeImports {
				http.Error(writer, "unable to combine devices when imports are expected (fix: set include_imports to false)", http.StatusBadRequest)
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
