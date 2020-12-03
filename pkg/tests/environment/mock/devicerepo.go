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
	"device-selection/pkg/model/devicemodel"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
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
	repo := &DeviceRepo{db: map[string]interface{}{}, localIds: map[string]bool{}}
	consumer.Subscribe(DtTopic, func(msg []byte) {
		cmd := DeviceTypeCommand{}
		json.Unmarshal(msg, &cmd)
		if cmd.Command == "PUT" {
			for i, service := range cmd.DeviceType.Services {
				service.AspectIds = nil
				service.FunctionIds = nil
				cmd.DeviceType.Services[i] = service
			}
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
