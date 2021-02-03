/*
 * Copyright 2021 InfAI (CC SES)
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
	"context"
	"device-selection/pkg/model"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"github.com/segmentio/kafka-go"
	"net"
	"net/http"
	"net/http/httptest"
)

type ImportTypeCommand struct {
	Command    string             `json:"command"`
	Id         string             `json:"id"`
	Owner      string             `json:"owner"`
	ImportType ImportTypeExtended `json:"import_type"`
}

type ImportTypeExtended struct {
	Id              string                      `json:"id"`
	Name            string                      `json:"name"`
	Description     string                      `json:"description"`
	Image           string                      `json:"image"`
	DefaultRestart  bool                        `json:"default_restart"`
	Configs         []model.ImportTypeConfig    `json:"configs"`
	AspectIds       []string                    `json:"aspect_ids"`
	Output          model.ImportContentVariable `json:"output"`
	FunctionIds     []string                    `json:"function_ids"`
	AspectFunctions []string                    `json:"aspect_functions"`
	Owner           string                      `json:"owner"`
}

type ImportRepo struct {
	importTypes map[string]model.ImportType
	writer      *kafka.Writer
	ts          *httptest.Server
}

func NewImportRepo(writer *kafka.Writer) *ImportRepo {
	repo := &ImportRepo{
		importTypes: map[string]model.ImportType{},
		writer:      writer,
	}

	router := httprouter.New()

	router.GET("/import-types/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		dt, ok := repo.importTypes[id]
		if ok {
			err := json.NewEncoder(writer).Encode(dt)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			http.Error(writer, "404", 404)
		}
	})

	router.PUT("/import-types/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		var importType model.ImportType
		err := json.NewDecoder(request.Body).Decode(&importType)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		msg := ImportTypeCommand{
			Command:    "PUT",
			Id:         id,
			Owner:      "anyone",
			ImportType: extendImportType(importType),
		}
		raw, _ := json.Marshal(&msg)
		err = repo.writer.WriteMessages(context.Background(), kafka.Message{
			Topic: "import-types",
			Value: raw,
			Key:   []byte(id),
		})
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		repo.importTypes[id] = importType
	})

	repo.ts = &httptest.Server{
		Config: &http.Server{Handler: router},
	}
	repo.ts.Listener, _ = net.Listen("tcp", ":")
	repo.ts.Start()

	return repo
}

func (this *ImportRepo) Stop() {
	this.ts.Close()
}

func (this *ImportRepo) Url() string {
	return this.ts.URL
}

func extendImportType(importType model.ImportType) ImportTypeExtended {
	aspectIds := []string{}
	for _, a := range importType.AspectIds.([]interface{}) {
		aspectIds = append(aspectIds, a.(string))
	}
	functionIds := []string{}
	for _, a := range importType.FunctionIds.([]interface{}) {
		functionIds = append(functionIds, a.(string))
	}
	ex := ImportTypeExtended{
		Id:              importType.Id,
		Name:            importType.Name,
		Description:     importType.Description,
		Image:           importType.Image,
		DefaultRestart:  importType.DefaultRestart,
		Configs:         importType.Configs,
		AspectIds:       aspectIds,
		Output:          importType.Output,
		FunctionIds:     functionIds,
		Owner:           importType.Owner,
		AspectFunctions: []string{},
	}
	for _, aspect := range aspectIds {
		for _, function := range functionIds {
			ex.AspectFunctions = append(ex.AspectFunctions, aspect+"_"+function)
		}
	}
	return ex
}
