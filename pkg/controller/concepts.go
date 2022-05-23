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
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

func (this *Controller) GetConcept(id string, token string) (c devicemodel.Concept, err error) {
	err = this.cache.Use(id, func() (interface{}, error) {
		req, err := http.NewRequest("GET", this.config.DeviceRepoUrl+"/concepts/"+id, nil)
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
			err := fmt.Errorf("unable to find concept: %v %v", id, buf.String())
			log.Println("ERROR:", err)
			debug.PrintStack()
			return nil, err
		}
		var cInner devicemodel.Concept
		err = json.NewDecoder(resp.Body).Decode(&cInner)
		return cInner, err
	}, &c)
	return
}
