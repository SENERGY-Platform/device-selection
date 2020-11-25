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

type SemanticRepo struct {
	dt              map[string]devicemodel.DeviceType
	concepts        map[string]devicemodel.Concept
	characteristics map[string]devicemodel.Characteristic
	aspects         map[string]devicemodel.Aspect
	functions       map[string]devicemodel.Function
	deviceclasses   map[string]devicemodel.DeviceClass
	ts              *httptest.Server
}

func NewSemanticRepo(consumer Consumer) *SemanticRepo {
	repo := &SemanticRepo{
		dt:              map[string]devicemodel.DeviceType{},
		concepts:        map[string]devicemodel.Concept{},
		characteristics: map[string]devicemodel.Characteristic{},
		aspects:         map[string]devicemodel.Aspect{},
		functions:       map[string]devicemodel.Function{},
		deviceclasses:   map[string]devicemodel.DeviceClass{},
	}
	consumer.Subscribe(DtTopic, func(msg []byte) {
		cmd := DeviceTypeCommand{}
		json.Unmarshal(msg, &cmd)
		if cmd.Command == "PUT" {
			for i, service := range cmd.DeviceType.Services {
				service.ProtocolId = ""
				cmd.DeviceType.Services[i] = service
			}
			repo.dt[cmd.Id] = cmd.DeviceType
		} else if cmd.Command == "DELETE" {
			delete(repo.dt, cmd.Id)
		}
	})
	consumer.Subscribe(ConceptTopic, func(msg []byte) {
		cmd := ConceptCommand{}
		json.Unmarshal(msg, &cmd)
		if cmd.Command == "PUT" {
			repo.concepts[cmd.Id] = cmd.Concept
		} else if cmd.Command == "DELETE" {
			delete(repo.concepts, cmd.Id)
		}
	})
	consumer.Subscribe(CharacteristicTopic, func(msg []byte) {
		cmd := CharacteristicCommand{}
		json.Unmarshal(msg, &cmd)
		if cmd.Command == "PUT" {
			repo.characteristics[cmd.Id] = cmd.Characteristic
		} else if cmd.Command == "DELETE" {
			delete(repo.characteristics, cmd.Id)
		}
	})

	consumer.Subscribe(AspectTopic, func(msg []byte) {
		cmd := AspectCommand{}
		json.Unmarshal(msg, &cmd)
		if cmd.Command == "PUT" {
			repo.aspects[cmd.Id] = cmd.Aspect
		} else if cmd.Command == "DELETE" {
			delete(repo.aspects, cmd.Id)
		}
	})

	consumer.Subscribe(FunctionTopic, func(msg []byte) {
		cmd := FunctionCommand{}
		json.Unmarshal(msg, &cmd)
		if cmd.Command == "PUT" {
			repo.functions[cmd.Id] = cmd.Function
		} else if cmd.Command == "DELETE" {
			delete(repo.functions, cmd.Id)
		}
	})

	consumer.Subscribe(DeviceClassTopic, func(msg []byte) {
		cmd := DeviceClassCommand{}
		json.Unmarshal(msg, &cmd)
		if cmd.Command == "PUT" {
			repo.deviceclasses[cmd.Id] = cmd.DeviceClass
		} else if cmd.Command == "DELETE" {
			delete(repo.deviceclasses, cmd.Id)
		}
	})

	router := httprouter.New()

	router.GET("/device-types", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		var descriptions model.FilterCriteriaAndSet
		filter := request.URL.Query().Get("filter")
		err := json.Unmarshal([]byte(filter), &descriptions)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		result := []devicemodel.DeviceType{}
		for _, dt := range repo.dt {
			if dtMatchesAllCriteria(dt, descriptions) {
				result = append(result, dt)
			}
		}
		log.Println("TEST-DEBUG: /device-types", result)
		json.NewEncoder(writer).Encode(result)
	})

	router.GET("/device-types/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		dt, ok := repo.dt[id]
		if ok {
			json.NewEncoder(writer).Encode(dt)
		} else {
			http.Error(writer, "404", 404)
		}
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
		writer.WriteHeader(http.StatusOK)
	})

	router.GET("/aspects/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		aspect, ok := repo.aspects[id]
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
		function, ok := repo.functions[id]
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
		deviceclass, ok := repo.deviceclasses[id]
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
		concept, ok := repo.concepts[id]
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
		concept, ok := repo.characteristics[id]
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
		for _, function := range service.FunctionIds {
			if function == criteria.FunctionId {
				if criteria.AspectId == "" {
					return true
				}
				for _, aspect := range service.AspectIds {
					if criteria.AspectId == aspect {
						return true
					}
				}
			}
		}
	}
	return false
}

func (this *SemanticRepo) Stop() {
	this.ts.Close()
}

func (this *SemanticRepo) Url() string {
	return this.ts.URL
}
