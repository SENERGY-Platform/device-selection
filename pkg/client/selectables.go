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
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/SENERGY-Platform/device-selection/v2/pkg/model"
	"github.com/SENERGY-Platform/models/go/models"
)

type GetSelectablesOptions struct {
	IncludeGroups               bool
	IncludeImports              bool
	IncludeDevices              bool
	IncludeIdModified           bool
	WithDeviceIds               []string
	WithLocalDeviceIds          []string
	LocalDeviceOwner            string
	FilterByDeviceAttributeKeys []string
}

func (c *ClientImpl) GetSelectables(token string, criteria []models.DeviceGroupFilterCriteria, options *GetSelectablesOptions) ([]model.Selectable, int, error) {
	query := url.Values{}
	if options != nil {
		query.Set("include_groups", strconv.FormatBool(options.IncludeGroups))
		query.Set("include_imports", strconv.FormatBool(options.IncludeImports))
		query.Set("include_devices", strconv.FormatBool(options.IncludeDevices))
		query.Set("include_id_modified", strconv.FormatBool(options.IncludeIdModified))
		if options.WithLocalDeviceIds != nil {
			query.Set("local_devices", strings.Join(options.WithLocalDeviceIds, ","))
		}
		if options.LocalDeviceOwner != "" {
			query.Set("local_device_owner", options.LocalDeviceOwner)
		}
		if options.WithDeviceIds != nil {
			query.Set("devices", strings.Join(options.WithDeviceIds, ","))
		}
		if len(options.FilterByDeviceAttributeKeys) > 0 {
			query.Set("filter_devices_by_attr_keys", strings.Join(options.FilterByDeviceAttributeKeys, ","))
		}
	}
	b, err := json.Marshal(criteria)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseUrl+"/v2/query/selectables?"+query.Encode(), bytes.NewBuffer(b))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	req.Header.Set("Authorization", token)
	return do[[]model.Selectable](req)
}

type GetBulkSelectablesOptions struct {
	// CompleteServices adds the full import type and the import path options to the answer.
	// Device services are complete either way; the name is a legacy artefact of the endpoint.
	CompleteServices bool
}

// GetBulkSelectablesV2 asks for the selectables of several criteria sets in one request. Each
// element of the answer carries the id its request element had, so a caller that named them
// after something of its own - a bpmn element, say - can match the result back to it without
// relying on the order.
func (c *ClientImpl) GetBulkSelectablesV2(token string, bulk model.BulkRequestV2, options *GetBulkSelectablesOptions) (model.BulkResult, int, error) {
	query := url.Values{}
	if options != nil {
		query.Set("complete_services", strconv.FormatBool(options.CompleteServices))
	}
	b, err := json.Marshal(bulk)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseUrl+"/v2/bulk/selectables?"+query.Encode(), bytes.NewBuffer(b))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	req.Header.Set("Authorization", token)
	return do[model.BulkResult](req)
}
