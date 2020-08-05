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

package devices

import (
	"device-selection/pkg/devicemodel"
	"encoding/json"
	"errors"
	"net/http"
	"runtime/debug"
	"time"
)

func (this *Devices) GetProtocols(token string) (result []devicemodel.Protocol, err error, code int) {
	err, code = this.getUncachedList(token, "protocols", &result)
	return
}

func (this *Devices) getUncachedList(token string, resource string, result interface{}) (error, int) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest(
		"GET",
		this.config.DeviceRepoUrl+"/"+resource,
		nil,
	)
	if err != nil {
		debug.PrintStack()
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)

	resp, err := client.Do(req)
	if err != nil {
		debug.PrintStack()
		return err, http.StatusInternalServerError
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		debug.PrintStack()
		return errors.New("unexpected statuscode"), resp.StatusCode
	}

	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		debug.PrintStack()
		return err, http.StatusInternalServerError
	}
	return nil, 200
}
