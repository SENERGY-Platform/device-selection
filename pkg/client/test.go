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
	value []model.Selectable
	code  int
	err   error
}

func NewTestClient() *TestClient {
	return &TestClient{value: []model.Selectable{}, code: http.StatusOK, err: nil}
}

func (c *TestClient) GetSelectables(token string, criteria []models.DeviceGroupFilterCriteria, options *GetSelectablesOptions) ([]model.Selectable, int, error) {
	return c.value, c.code, c.err
}

func (c *TestClient) SetResponse(value []model.Selectable, code int, err error) {
	c.value = value
	c.code = code
	c.err = err
}
