/*
 * Copyright 2024 InfAI (CC SES)
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

package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/SENERGY-Platform/device-selection/pkg/model"
	"github.com/SENERGY-Platform/models/go/models"
)

type Client interface {
	GetSelectables(token string, criteria []models.DeviceGroupFilterCriteria, options *GetSelectablesOptions) ([]model.Selectable, int, error)
}

type ClientImpl struct {
	baseUrl string
}

// Client only implements GetSelectables. Client should be extended if more API functionality is required
func NewClient(baseUrl string) Client {
	return &ClientImpl{baseUrl: baseUrl}
}

func do[T any](req *http.Request) (result T, code int, err error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, http.StatusInternalServerError, err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		temp, _ := io.ReadAll(resp.Body) //read error response end ensure that resp.Body is read to EOF
		return result, resp.StatusCode, fmt.Errorf("unexpected statuscode %v: %v", resp.StatusCode, string(temp))
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		_, _ = io.ReadAll(resp.Body) //ensure resp.Body is read to EOF
		return result, http.StatusInternalServerError, err
	}
	return
}
