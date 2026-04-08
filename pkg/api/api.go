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
	"context"
	"errors"
	"log"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/SENERGY-Platform/device-selection/pkg/api/util"
	"github.com/SENERGY-Platform/device-selection/pkg/configuration"
	"github.com/SENERGY-Platform/device-selection/pkg/controller"
	"github.com/SENERGY-Platform/service-commons/pkg/accesslog"
)

//go:generate go install github.com/swaggo/swag/cmd/swag@latest
//go:generate swag init --instanceName devicemanager -o ../../docs --parseDependency -d . -g api.go

type EndpointMethod = func(router *http.ServeMux, config configuration.Config, control *controller.Controller)

var endpoints = []interface{}{} //list of objects with EndpointMethod

// starts http server; if wg is not nil it will be set as done when the server is stopped
func Start(ctx context.Context, wg *sync.WaitGroup, config configuration.Config, ctrl *controller.Controller) (err error) {
	config.GetLogger().Info("start api on " + config.ApiPort)
	router := Router(config, ctrl)
	server := &http.Server{Addr: ":" + config.ApiPort, Handler: router, WriteTimeout: 10 * time.Second, ReadTimeout: 2 * time.Second, ReadHeaderTimeout: 2 * time.Second}
	wg.Add(1)
	go func() {
		config.GetLogger().Info("listening", "address", server.Addr)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			config.GetLogger().Error("FATAL: api server error", "error", err)
			log.Fatal(err)
		}
	}()
	go func() {
		<-ctx.Done()
		config.GetLogger().Info("api shutdown", "result", server.Shutdown(context.Background()))
		wg.Done()
	}()
	return nil
}

// Router doc
// @title         Device-Selection API
// @version       0.1
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath  /
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func Router(config configuration.Config, ctrl *controller.Controller) http.Handler {
	handler := GetRouterWithoutMiddleware(config, ctrl)
	config.GetLogger().Info("add cors")
	corsHandler := util.NewCors(handler)
	config.GetLogger().Info("add logging")
	logger := accesslog.New(corsHandler)
	return logger
}

func GetRouterWithoutMiddleware(config configuration.Config, command *controller.Controller) http.Handler {
	router := http.NewServeMux()
	config.GetLogger().Info("add heart beat endpoint")
	router.HandleFunc("GET /", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})
	for _, e := range endpoints {
		for name, call := range getEndpointMethods(e) {
			config.GetLogger().Info("add endpoint " + name)
			call(router, config, command)
		}
	}
	return router
}

func getEndpointMethods(e interface{}) map[string]EndpointMethod {
	result := map[string]EndpointMethod{}
	objRef := reflect.ValueOf(e)
	methodCount := objRef.NumMethod()
	for i := 0; i < methodCount; i++ {
		m := objRef.Method(i)
		f, ok := m.Interface().(EndpointMethod)
		if ok {
			name := getTypeName(objRef.Type()) + "::" + objRef.Type().Method(i).Name
			result[name] = f
		}
	}
	return result
}

func getTypeName(t reflect.Type) (res string) {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}
