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

package controller

import (
	"bytes"
	"device-selection/pkg/model/devicemodel"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
)

func (this *Controller) GetAspectNode(id string, token string) (result devicemodel.AspectNode, err error) {
	err = this.cache.Use("aspect-nodes."+id, func() (interface{}, error) {
		req, err := http.NewRequest("GET", this.config.DeviceRepoUrl+"/aspect-nodes/"+url.PathEscape(id), nil)
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
			return nil, fmt.Errorf("unable to find aspect: %v %v", id, buf.String())
		}
		var aspect devicemodel.AspectNode
		err = json.NewDecoder(resp.Body).Decode(&aspect)
		return aspect, err
	}, &result)
	return
}
