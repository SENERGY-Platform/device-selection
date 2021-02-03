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

package mock

import (
	"device-selection/pkg/model"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net"
	"net/http"
	"net/http/httptest"
)

type ImportDeploy struct {
	instances []model.Import
	ts        *httptest.Server
}

func NewImportDeploy() *ImportDeploy {
	deploy := &ImportDeploy{
		instances: []model.Import{},
	}

	router := httprouter.New()

	router.GET("/instances", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		err := json.NewEncoder(writer).Encode(deploy.instances)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
	})

	router.PUT("/instances", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		err := json.NewDecoder(request.Body).Decode(&deploy.instances)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
	})

	deploy.ts = &httptest.Server{
		Config: &http.Server{Handler: router},
	}
	deploy.ts.Listener, _ = net.Listen("tcp", ":")
	deploy.ts.Start()

	return deploy
}

func (this *ImportDeploy) Stop() {
	this.ts.Close()
}

func (this *ImportDeploy) Url() string {
	return this.ts.URL
}
