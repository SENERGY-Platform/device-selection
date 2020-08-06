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

package devicemodel

type Device struct {
	Id           string `json:"id"`
	LocalId      string `json:"local_id,omitempty"`
	Name         string `json:"name,omitempty"`
	DeviceTypeId string `json:"device_type_id,omitempty"`
}

type DeviceType struct {
	Id          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Image       string      `json:"image"`
	Services    []Service   `json:"services"`
	DeviceClass DeviceClass `json:"device_class"`
	RdfType     string      `json:"rdf_type"`
}

type Service struct {
	Id          string     `json:"id"`
	LocalId     string     `json:"local_id,omitempty"`
	Name        string     `json:"name,omitempty"`
	Description string     `json:"description,omitempty"`
	Aspects     []Aspect   `json:"aspects,omitempty"`
	ProtocolId  string     `json:"protocol_id,omitempty"`
	Inputs      []Content  `json:"inputs,omitempty"`
	Outputs     []Content  `json:"outputs,omitempty"`
	Functions   []Function `json:"functions,omitempty"`
	RdfType     string     `json:"rdf_type,omitempty"`
}

type Interaction string

const (
	EVENT             Interaction = "event"
	REQUEST           Interaction = "request"
	EVENT_AND_REQUEST Interaction = "event+request"
)

type Protocol struct {
	Id               string            `json:"id"`
	Name             string            `json:"name"`
	Handler          string            `json:"handler"`
	Interaction      Interaction       `json:"interaction"`
	ProtocolSegments []ProtocolSegment `json:"protocol_segments"`
}
