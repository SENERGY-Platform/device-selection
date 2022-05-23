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

package controller

import (
	"bytes"
	"device-selection/pkg/model/devicemodel"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
)

func (this *Controller) GetFunction(id string, token string) (f devicemodel.Function, err error) {
	functions, err := this.GetFunctions(token)
	if err != nil {
		return
	}
	for _, function := range functions {
		if function.Id == id {
			return function, nil
		}
	}
	return f, errors.New("not found")
}

func (this *Controller) GetFunctions(token string) (functions []devicemodel.Function, err error) {
	err = this.cache.Use("functions", func() (interface{}, error) {
		req, err := http.NewRequest("GET", this.config.PermSearchUrl+"/v3/resources/functions", nil)
		if err != nil {
			debug.PrintStack()
			return nil, err
		}
		req.Header.Set("Authorization", token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			debug.PrintStack()
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			debug.PrintStack()
			return nil, fmt.Errorf("unable to find functions: %v", buf.String())
		}
		var fu []devicemodel.Function
		err = json.NewDecoder(resp.Body).Decode(&fu)
		return fu, err
	}, &functions)
	return
}
