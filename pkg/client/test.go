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
	"net/http"

	"github.com/SENERGY-Platform/device-selection/pkg/model"
	"github.com/SENERGY-Platform/models/go/models"
)

type TestClient struct {
	value                   []model.Selectable
	code                    int
	err                     error
	deviceGroupHelperResult model.DeviceGroupHelperResult
	bulkResult              model.BulkResult
}

func NewTestClient() *TestClient {
	return &TestClient{value: []model.Selectable{}, code: http.StatusOK, err: nil}
}

func (c *TestClient) GetSelectables(token string, criteria []models.DeviceGroupFilterCriteria, options *GetSelectablesOptions) ([]model.Selectable, int, error) {
	return c.value, c.code, c.err
}

// GetBulkSelectablesV2 answers every request element with the same selectables, unless a bulk
// result was set explicitly. Answering per element is what a caller that keys the result by id
// needs, so the fallback keeps the ids of the request rather than returning an empty list.
func (c *TestClient) GetBulkSelectablesV2(token string, bulk model.BulkRequestV2, options *GetBulkSelectablesOptions) (model.BulkResult, int, error) {
	if c.bulkResult != nil {
		return c.bulkResult, c.code, c.err
	}
	result := model.BulkResult{}
	for _, element := range bulk {
		result = append(result, model.BulkResultElement{Id: element.Id, Selectables: c.value})
	}
	return result, c.code, c.err
}

func (c *TestClient) DeviceGroupHelper(token string, deviceIds []string, options *DeviceGroupHelperOptions) (model.DeviceGroupHelperResult, int, error) {
	return c.deviceGroupHelperResult, c.code, c.err
}

func (c *TestClient) SetResponse(value []model.Selectable, code int, err error) {
	c.value = value
	c.code = code
	c.err = err
}

func (c *TestClient) SetDeviceGroupHelperResponse(value model.DeviceGroupHelperResult, code int, err error) {
	c.deviceGroupHelperResult = value
	c.code = code
	c.err = err
}

func (c *TestClient) SetBulkResponse(value model.BulkResult, code int, err error) {
	c.bulkResult = value
	c.code = code
	c.err = err
}
