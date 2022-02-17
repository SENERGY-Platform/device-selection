/*
 * Copyright 2019 InfAI (CC SES)
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

package mock

import (
	"device-selection/pkg/model"
	"device-selection/pkg/model/devicemodel"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
)

type DeviceRepo struct {
	db       map[string]interface{}
	ts       *httptest.Server
	localIds map[string]bool
}

func NewDeviceRepo(consumer Consumer) *DeviceRepo {
	repo := &DeviceRepo{
		db:       map[string]interface{}{},
		localIds: map[string]bool{},
	}

	consumer.Subscribe(ConceptTopic, func(msg []byte) {
		cmd := ConceptCommand{}
		json.Unmarshal(msg, &cmd)
		if cmd.Command == "PUT" {
			repo.db[cmd.Id] = cmd.Concept
		} else if cmd.Command == "DELETE" {
			delete(repo.db, cmd.Id)
		}
	})
	consumer.Subscribe(CharacteristicTopic, func(msg []byte) {
		cmd := CharacteristicCommand{}
		json.Unmarshal(msg, &cmd)
		if cmd.Command == "PUT" {
			repo.db[cmd.Id] = cmd.Characteristic
		} else if cmd.Command == "DELETE" {
			delete(repo.db, cmd.Id)
		}
	})

	consumer.Subscribe(AspectTopic, func(msg []byte) {
		cmd := AspectCommand{}
		json.Unmarshal(msg, &cmd)
		if cmd.Command == "PUT" {
			repo.db[cmd.Id] = cmd.Aspect
		} else if cmd.Command == "DELETE" {
			delete(repo.db, cmd.Id)
		}
	})

	consumer.Subscribe(FunctionTopic, func(msg []byte) {
		cmd := FunctionCommand{}
		json.Unmarshal(msg, &cmd)
		if cmd.Command == "PUT" {
			repo.db[cmd.Id] = cmd.Function
		} else if cmd.Command == "DELETE" {
			delete(repo.db, cmd.Id)
		}
	})

	consumer.Subscribe(DeviceClassTopic, func(msg []byte) {
		cmd := DeviceClassCommand{}
		json.Unmarshal(msg, &cmd)
		if cmd.Command == "PUT" {
			repo.db[cmd.Id] = cmd.DeviceClass
		} else if cmd.Command == "DELETE" {
			delete(repo.db, cmd.Id)
		}
	})

	consumer.Subscribe(DtTopic, func(msg []byte) {
		cmd := DeviceTypeCommand{}
		json.Unmarshal(msg, &cmd)
		if cmd.Command == "PUT" {
			repo.db[cmd.Id] = cmd.DeviceType
		} else if cmd.Command == "DELETE" {
			delete(repo.db, cmd.Id)
		}
	})
	consumer.Subscribe(DeviceTopic, func(msg []byte) {
		cmd := DeviceCommand{}
		json.Unmarshal(msg, &cmd)
		if cmd.Command == "PUT" {
			repo.localIds[cmd.Device.LocalId] = true
			repo.db[cmd.Id] = cmd.Device
		} else if cmd.Command == "DELETE" {
			delete(repo.db, cmd.Id)
		}
	})

	consumer.Subscribe(DeviceGroupTopic, func(msg []byte) {
		cmd := DeviceGroupCommand{}
		json.Unmarshal(msg, &cmd)
		if cmd.Command == "PUT" {
			repo.db[cmd.Id] = cmd.DeviceGroup
		} else if cmd.Command == "DELETE" {
			delete(repo.db, cmd.Id)
		}
	})

	consumer.Subscribe(HubTopic, func(msg []byte) {
		cmd := HubCommand{}
		json.Unmarshal(msg, &cmd)
		if cmd.Command == "PUT" {
			repo.db[cmd.Id] = cmd.Hub
		} else if cmd.Command == "DELETE" {
			delete(repo.db, cmd.Id)
		}
	})
	consumer.Subscribe(ProtocolTopic, func(msg []byte) {
		cmd := ProtocolCommand{}
		json.Unmarshal(msg, &cmd)
		if cmd.Command == "PUT" {
			repo.db[cmd.Id] = cmd.Protocol
		} else if cmd.Command == "DELETE" {
			delete(repo.db, cmd.Id)
		}
	})

	router := httprouter.New()

	router.GET("/device-groups/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		dt, ok := repo.db[id]
		if ok {
			json.NewEncoder(writer).Encode(dt)
		} else {
			http.Error(writer, "404", 404)
		}
	})

	router.PUT("/device-groups", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		group := devicemodel.DeviceGroup{}
		err = json.NewDecoder(request.Body).Decode(&group)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if group.Id == "" {
			http.Error(writer, "missing device id", http.StatusBadRequest)
			return
		}

		writer.WriteHeader(http.StatusOK)
	})

	router.GET("/device-types/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		dt, ok := repo.db[id]
		if ok {
			json.NewEncoder(writer).Encode(dt)
		} else {
			http.Error(writer, "404", 404)
		}
	})

	router.GET("/device-types", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		var descriptions model.FilterCriteriaAndSet
		filter := request.URL.Query().Get("filter")
		err := json.Unmarshal([]byte(filter), &descriptions)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		result := []devicemodel.DeviceType{}
		for _, dtInterface := range repo.db {
			dt, ok := dtInterface.(devicemodel.DeviceType)
			if ok && dtMatchesAllCriteria(dt, descriptions) {
				result = append(result, dt)
			}
		}
		log.Println("TEST-DEBUG: /device-types", result)
		json.NewEncoder(writer).Encode(result)
	})

	router.PUT("/device-types", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		dt := devicemodel.DeviceType{}
		err = json.NewDecoder(request.Body).Decode(&dt)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if dt.Id == "" {
			http.Error(writer, "missing device id", http.StatusBadRequest)
			return
		}
		if len(dt.Services) == 0 {
			http.Error(writer, "expect at least one service", http.StatusBadRequest)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})

	router.GET("/devices/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		dt, ok := repo.db[id]
		if ok {
			json.NewEncoder(writer).Encode(dt)
		} else {
			http.Error(writer, "404", 404)
		}
	})

	router.PUT("/devices", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		device := devicemodel.Device{}
		err = json.NewDecoder(request.Body).Decode(&device)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if device.Id == "" {
			http.Error(writer, "missing device id", http.StatusBadRequest)
			return
		}
		if device.LocalId == "" {
			http.Error(writer, "missing local id", http.StatusBadRequest)
			return
		}
		if _, ok := repo.localIds[device.LocalId]; ok {
			http.Error(writer, "expect local id to be globally unique", http.StatusBadRequest)
			return
		}
		if device.DeviceTypeId == "" {
			http.Error(writer, "missing device-type id", http.StatusBadRequest)
			return
		}

		writer.WriteHeader(http.StatusOK)
	})

	router.GET("/hubs/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		dt, ok := repo.db[id]
		if ok {
			json.NewEncoder(writer).Encode(dt)
		} else {
			http.Error(writer, "404", 404)
		}
	})

	router.PUT("/hubs", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		hub := devicemodel.Hub{}
		err = json.NewDecoder(request.Body).Decode(&hub)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if hub.Id == "" {
			http.Error(writer, "missing hub id", http.StatusBadRequest)
			return
		}
		if hub.Name == "" {
			http.Error(writer, "missing hub name", http.StatusBadRequest)
			return
		}

		if len(hub.DeviceLocalIds) == 1 {
			if _, ok := repo.localIds[hub.DeviceLocalIds[0]]; !ok {
				http.Error(writer, "unknown device local id", http.StatusBadRequest)
				return
			}
		}
		writer.WriteHeader(http.StatusOK)
	})

	router.PUT("/protocols", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		protocol := devicemodel.Protocol{}
		err = json.NewDecoder(request.Body).Decode(&protocol)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if protocol.Id == "" {
			http.Error(writer, "missing device id", http.StatusBadRequest)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})

	router.GET("/aspects/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		aspect, ok := repo.db[id]
		if ok {
			json.NewEncoder(writer).Encode(aspect)
		} else {
			http.Error(writer, "404", 404)
		}
	})

	router.PUT("/aspects", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		aspect := devicemodel.Aspect{}
		err = json.NewDecoder(request.Body).Decode(&aspect)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if aspect.Id == "" {
			http.Error(writer, "missing device id", http.StatusBadRequest)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})

	router.GET("/functions/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		function, ok := repo.db[id]
		if ok {
			json.NewEncoder(writer).Encode(function)
		} else {
			http.Error(writer, "404", 404)
		}
	})

	router.PUT("/functions", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		function := devicemodel.Function{}
		err = json.NewDecoder(request.Body).Decode(&function)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if function.Id == "" {
			http.Error(writer, "missing device id", http.StatusBadRequest)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})

	router.GET("/device-classes/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		deviceclass, ok := repo.db[id]
		if ok {
			json.NewEncoder(writer).Encode(deviceclass)
		} else {
			http.Error(writer, "404", 404)
		}
	})

	router.PUT("/device-classes", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		deviceclass := devicemodel.DeviceClass{}
		err = json.NewDecoder(request.Body).Decode(&deviceclass)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if deviceclass.Id == "" {
			http.Error(writer, "missing device id", http.StatusBadRequest)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})

	router.GET("/concepts/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		concept, ok := repo.db[id]
		if ok {
			json.NewEncoder(writer).Encode(concept)
		} else {
			http.Error(writer, "404", 404)
		}
	})

	router.PUT("/concepts", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		concept := devicemodel.Concept{}
		err = json.NewDecoder(request.Body).Decode(&concept)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if concept.Id == "" {
			http.Error(writer, "missing concept id", http.StatusBadRequest)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})

	router.PUT("/characteristics", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		characteristic := devicemodel.Characteristic{}
		err = json.NewDecoder(request.Body).Decode(&characteristic)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if characteristic.Id == "" {
			http.Error(writer, "missing characteristic id", http.StatusBadRequest)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})

	router.GET("/characteristics/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		concept, ok := repo.db[id]
		if ok {
			json.NewEncoder(writer).Encode(concept)
		} else {
			http.Error(writer, "404", 404)
		}
	})

	repo.ts = &httptest.Server{
		Config: &http.Server{Handler: router},
	}
	repo.ts.Listener, _ = net.Listen("tcp", ":")
	repo.ts.Start()

	return repo
}

func (this *DeviceRepo) Stop() {
	this.ts.Close()
}

func (this *DeviceRepo) Url() string {
	return this.ts.URL
}

func dtMatchesAllCriteria(dt devicemodel.DeviceType, descriptions model.FilterCriteriaAndSet) bool {
	for _, criteria := range descriptions {
		if !dtMatchesCriteria(dt, criteria) {
			return false
		}
	}
	return true
}

func dtMatchesCriteria(dt devicemodel.DeviceType, criteria devicemodel.FilterCriteria) bool {
	if criteria.DeviceClassId != "" && criteria.DeviceClassId != dt.DeviceClassId {
		return false
	}
	for _, service := range dt.Services {
		for _, content := range service.Inputs {
			if contentVariableContainsCriteria(content.ContentVariable, criteria) {
				return true
			}
		}
		for _, content := range service.Outputs {
			if contentVariableContainsCriteria(content.ContentVariable, criteria) {
				return true
			}
		}
	}
	return false
}

//simple without aspect hierarchy
func contentVariableContainsCriteria(variable devicemodel.ContentVariable, criteria devicemodel.FilterCriteria) bool {
	if variable.FunctionId == criteria.FunctionId && (criteria.AspectId == "" || variable.AspectId == criteria.AspectId) {
		return true
	}
	for _, sub := range variable.SubContentVariables {
		if contentVariableContainsCriteria(sub, criteria) {
			return true
		}
	}
	return false
}
