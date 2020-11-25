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

package mock

import (
	"context"
	"device-selection/pkg/model/devicemodel"
	"device-selection/pkg/tests/environment/kafka"
	"log"
)

type ConsumerImpl struct {
	ctx     context.Context
	zkUrl   string
	groupId string
}

func NewConsumer(ctx context.Context, zkUrl string, groupId string) *ConsumerImpl {
	return &ConsumerImpl{ctx: ctx, zkUrl: zkUrl, groupId: groupId}
}

func (this *ConsumerImpl) Subscribe(topic string, callback func(msg []byte)) {
	err := kafka.NewConsumer(this.ctx, this.zkUrl, this.groupId, topic, func(delivery []byte) error {
		log.Println("TEST-DEBUG: consume", this.groupId, topic, string(delivery))
		callback(delivery)
		return nil
	})
	if err != nil {
		log.Println("FATAL: Subscribe()", err)
		panic(err)
	}
}

var DtTopic = "device-types"
var ProtocolTopic = "protocols"
var DeviceTopic = "devices"
var DeviceGroupTopic = "device-groups"
var HubTopic = "hubs"
var ConceptTopic = "concepts"
var CharacteristicTopic = "characteristics"
var AspectTopic = "aspects"
var FunctionTopic = "functions"
var DeviceClassTopic = "device-classes"

type DeviceCommand struct {
	Command string             `json:"command"`
	Id      string             `json:"id"`
	Owner   string             `json:"owner"`
	Device  devicemodel.Device `json:"device"`
}

type AspectCommand struct {
	Command string             `json:"command"`
	Id      string             `json:"id"`
	Owner   string             `json:"owner"`
	Aspect  devicemodel.Aspect `json:"aspect"`
}

type CharacteristicCommand struct {
	Command        string                     `json:"command"`
	ConceptId      string                     `json:"concept_id"`
	Id             string                     `json:"id"`
	Owner          string                     `json:"owner"`
	Characteristic devicemodel.Characteristic `json:"characteristic"`
}

type ConceptCommand struct {
	Command string              `json:"command"`
	Id      string              `json:"id"`
	Owner   string              `json:"owner"`
	Concept devicemodel.Concept `json:"concept"`
}

type DeviceClassCommand struct {
	Command     string                  `json:"command"`
	Id          string                  `json:"id"`
	Owner       string                  `json:"owner"`
	DeviceClass devicemodel.DeviceClass `json:"device_class"`
}

type DeviceGroupCommand struct {
	Command     string                  `json:"command"`
	Id          string                  `json:"id"`
	Owner       string                  `json:"owner"`
	DeviceGroup devicemodel.DeviceGroup `json:"device_group"`
}

type DeviceTypeCommand struct {
	Command    string                 `json:"command"`
	Id         string                 `json:"id"`
	Owner      string                 `json:"owner"`
	DeviceType devicemodel.DeviceType `json:"device_type"`
}

type FunctionCommand struct {
	Command  string               `json:"command"`
	Id       string               `json:"id"`
	Owner    string               `json:"owner"`
	Function devicemodel.Function `json:"function"`
}

type HubCommand struct {
	Command string          `json:"command"`
	Id      string          `json:"id"`
	Owner   string          `json:"owner"`
	Hub     devicemodel.Hub `json:"hub"`
}

type ProtocolCommand struct {
	Command  string               `json:"command"`
	Id       string               `json:"id"`
	Owner    string               `json:"owner"`
	Protocol devicemodel.Protocol `json:"protocol"`
}
