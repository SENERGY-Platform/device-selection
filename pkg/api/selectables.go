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
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
)

func init() {
	endpoints = append(endpoints, &SelectablesEndpoints{})
}

type SelectablesEndpoints struct{}

// Selectables godoc
// @Summary      deprecated selectables
// @Description  deprecated; finds devices, device-groups and/or imports that match all provided filter-criteria
// @Tags         selectables, deprecated
// @Produce      json
// @Security Bearer
// @Param        include_groups query bool false "result should include matching device-groups"
// @Param        include_imports query bool false "result should include matching imports"
// @Param        local_devices query string false "comma seperated list of local device ids; result devices must be in this list (if one is given)"
// @Param        complete_services query bool false "adds full import-type and import path options to the result. device services are already complete, the name is a legacy artefact"
// @Param        filter_protocols query string false "comma seperated list of protocol ids, that should be ignored"
// @Param        filter_interaction query string false "interaction that is not allowed in the result"
// @Param        json query string false "json encoded criteria list (model.FilterCriteriaAndSet like [{&quot;function_id&quot;:&quot;&quot;,&quot;aspect_id&quot;:&quot;&quot;,&quot;device_class_id&quot;:&quot;&quot;}])"
// @Param        base64 query string false "alternative to json; base64 encoded json of criteria list"
// @Param        function_id query string false "alternative to json and base64 if only one filter criteria is needed"
// @Param        device_class_id query string false "alternative to json and base64 if only one filter criteria is needed"
// @Param        aspect_id query string false "alternative to json and base64 if only one filter criteria is needed"
// @Success      200 {array}  []model.Selectable
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /selectables [GET]
func (this *DeviceGroupsHelper) Selectables(router *http.ServeMux, config configuration.Config, ctrl *controller.Controller) {
	router.HandleFunc("GET /selectables", func(writer http.ResponseWriter, request *http.Request) {
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
}

// SelectablesV2 godoc
// @Summary      selectables
// @Description  finds devices, device-groups and/or imports that match all provided filter-criteria
// @Tags         selectables
// @Produce      json
// @Security Bearer
// @Param        include_devices query bool false "result should include matching devices"
// @Param        include_groups query bool false "result should include matching device-groups"
// @Param        include_imports query bool false "result should include matching imports"
// @Param        include_id_modified query bool false "result should include all valid device id modifications"
// @Param        import_path_trim_first_element query bool false "trim first element of import paths"
// @Param        devices query string false "comma seperated list of device ids; result devices must be in this list (if one is given)"
// @Param        local_devices query string false "comma seperated list of local device ids; result devices must be in this list (if one is given)"
// @Param        local_device_owner query string false "used in combination with local_devices to identify devices, default is the requesting user"
// @Param        json query string false "json encoded criteria list (model.FilterCriteriaAndSet like [{&quot;interaction&quot;:&quot;&quot;,&quot;function_id&quot;:&quot;&quot;,&quot;aspect_id&quot;:&quot;&quot;,&quot;device_class_id&quot;:&quot;&quot;}])"
// @Param        base64 query string false "alternative to json; base64 encoded json of criteria list"
// @Param        interaction query string false "alternative to json and base64 if only one filter criteria is needed"
// @Param        function_id query string false "alternative to json and base64 if only one filter criteria is needed"
// @Param        device_class_id query string false "alternative to json and base64 if only one filter criteria is needed"
// @Param        aspect_id query string false "alternative to json and base64 if only one filter criteria is needed"
// @Param        filter_devices_by_attr_keys query string false "comma seperated list of attribute keys; result devices have these attributes (if one is given)"
// @Success      200 {array}  []model.Selectable
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/selectables [GET]
func (this *DeviceGroupsHelper) SelectablesV2(router *http.ServeMux, config configuration.Config, ctrl *controller.Controller) {
	router.HandleFunc("GET /v2/selectables", func(writer http.ResponseWriter, request *http.Request) {
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
		importPathTrimFirstElement, _ := strconv.ParseBool(request.URL.Query().Get("import_path_trim_first_element"))

		var withLocalDeviceIds []string
		localDevicesQueryParam := request.URL.Query().Get("local_devices")
		if request.URL.Query().Has("local_devices") {
			withLocalDeviceIds = []string{}
			for _, localId := range strings.Split(localDevicesQueryParam, ",") {
				withLocalDeviceIds = append(withLocalDeviceIds, strings.TrimSpace(localId))
			}
		}

		localDeviceOwner := request.URL.Query().Get("local_device_owner")

		var withDeviceIds []string
		devicesQueryParam := request.URL.Query().Get("devices")
		if request.URL.Query().Has("devices") {
			withDeviceIds = []string{}
			for _, Id := range strings.Split(devicesQueryParam, ",") {
				withDeviceIds = append(withDeviceIds, strings.TrimSpace(Id))
			}
		}

		var filterDevicesByAttributeKeys []string
		filterDevicesByAttributeKeysParam := request.URL.Query().Get("filter_devices_by_attr_keys")
		if filterDevicesByAttributeKeysParam != "" {
			for _, key := range strings.Split(filterDevicesByAttributeKeysParam, ",") {
				filterDevicesByAttributeKeys = append(filterDevicesByAttributeKeys, strings.TrimSpace(key))
			}
		}

		result, err, code := ctrl.GetFilteredDevicesV2(token, model.GetFilteredDevicesV2Options{
			FilterCriteria:              criteria,
			IncludeDevices:              includeDevices,
			IncludeGroups:               includeGroups,
			IncludeImports:              includeImports,
			IncludeIdModified:           includeIdModified,
			WithDeviceIds:               withDeviceIds,
			WithLocalDeviceIds:          withLocalDeviceIds,
			LocalDeviceOwner:            localDeviceOwner,
			FilterByDeviceAttributeKeys: filterDevicesByAttributeKeys,
			ImportPathTrimFirstElement:  importPathTrimFirstElement,
		})
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

// QuerySelectables godoc
// @Summary      selectables
// @Description  finds devices, device-groups and/or imports that match all provided filter-criteria
// @Tags         selectables
// @Produce      json
// @Security Bearer
// @Param        include_devices query bool false "result should include matching devices"
// @Param        include_groups query bool false "result should include matching device-groups"
// @Param        include_imports query bool false "result should include matching imports"
// @Param        include_id_modified query bool false "result should include all valid device id modifications"
// @Param        import_path_trim_first_element query bool false "trim first element of import paths"
// @Param        devices query string false "comma seperated list of device ids; result devices must be in this list (if one is given)"
// @Param        local_devices query string false "comma seperated list of local device ids; result devices must be in this list (if one is given)"
// @Param        local_device_owner query string false "used in combination with local_devices to identify devices, default is the requesting user"
// @Param        filter_devices_by_attr_keys query string false "comma seperated list of attribute keys; result devices have these attributes (if one is given)"
// @Param        message body model.FilterCriteriaAndSet true "criteria list"
// @Success      200 {array}  []model.Selectable
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/query/selectables [POST]
func (this *DeviceGroupsHelper) QuerySelectables(router *http.ServeMux, config configuration.Config, ctrl *controller.Controller) {
	router.HandleFunc("POST /v2/query/selectables", func(writer http.ResponseWriter, request *http.Request) {
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
		importPathTrimFirstElement, _ := strconv.ParseBool(request.URL.Query().Get("import_path_trim_first_element"))

		var withLocalDeviceIds []string = nil
		localDevicesQueryParam := request.URL.Query().Get("local_devices")
		if localDevicesQueryParam != "" {
			for _, localId := range strings.Split(localDevicesQueryParam, ",") {
				withLocalDeviceIds = append(withLocalDeviceIds, strings.TrimSpace(localId))
			}
		}

		localDeviceOwner := request.URL.Query().Get("local_device_owner")

		var withDeviceIds []string = nil
		devicesQueryParam := request.URL.Query().Get("devices")
		if devicesQueryParam != "" {
			for _, id := range strings.Split(devicesQueryParam, ",") {
				withDeviceIds = append(withDeviceIds, strings.TrimSpace(id))
			}
		}

		var filterDevicesByAttributeKeys []string
		filterDevicesByAttributeKeysParam := request.URL.Query().Get("filter_devices_by_attr_keys")
		if filterDevicesByAttributeKeysParam != "" {
			for _, key := range strings.Split(filterDevicesByAttributeKeysParam, ",") {
				filterDevicesByAttributeKeys = append(filterDevicesByAttributeKeys, strings.TrimSpace(key))
			}
		}

		result, err, code := ctrl.GetFilteredDevicesV2(token, model.GetFilteredDevicesV2Options{
			FilterCriteria:              criteria,
			IncludeDevices:              includeDevices,
			IncludeGroups:               includeGroups,
			IncludeImports:              includeImports,
			IncludeIdModified:           includeIdModified,
			WithDeviceIds:               withDeviceIds,
			WithLocalDeviceIds:          withLocalDeviceIds,
			LocalDeviceOwner:            localDeviceOwner,
			FilterByDeviceAttributeKeys: filterDevicesByAttributeKeys,
			ImportPathTrimFirstElement:  importPathTrimFirstElement,
		})
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
