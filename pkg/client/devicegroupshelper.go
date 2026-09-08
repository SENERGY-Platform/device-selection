/*
 * Copyright 2026 InfAI (CC SES)
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
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/SENERGY-Platform/device-selection/pkg/model"
)

type DeviceGroupHelperOptions struct {
	Search                  string
	Limit                   int64 //ignored if 0; the service defaults to 100
	Offset                  int64 //ignored if 0
	MaintainsGroupUsability bool
	FunctionBlockList       []string
}

func (c *ClientImpl) DeviceGroupHelper(token string, deviceIds []string, options *DeviceGroupHelperOptions) (result model.DeviceGroupHelperResult, code int, err error) {
	query := url.Values{}
	if options != nil {
		if options.Search != "" {
			query.Set("search", options.Search)
		}
		if options.Limit > 0 {
			query.Set("limit", strconv.FormatInt(options.Limit, 10))
		}
		if options.Offset > 0 {
			query.Set("offset", strconv.FormatInt(options.Offset, 10))
		}
		query.Set("maintains_group_usability", strconv.FormatBool(options.MaintainsGroupUsability))
		if len(options.FunctionBlockList) > 0 {
			query.Set("function_block_list", strings.Join(options.FunctionBlockList, ","))
		}
	}
	if deviceIds == nil {
		deviceIds = []string{}
	}
	b, err := json.Marshal(deviceIds)
	if err != nil {
		return result, http.StatusInternalServerError, err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseUrl+"/device-group-helper?"+query.Encode(), bytes.NewBuffer(b))
	if err != nil {
		return result, http.StatusInternalServerError, err
	}
	req.Header.Set("Authorization", token)
	return do[model.DeviceGroupHelperResult](req)
}
