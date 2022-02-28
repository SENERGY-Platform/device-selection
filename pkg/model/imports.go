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

package model

import "device-selection/pkg/model/basecontentvariable"

type ImportType struct {
	Id             string                `json:"id"`
	Name           string                `json:"name"`
	Description    string                `json:"description"`
	Image          string                `json:"image"`
	DefaultRestart bool                  `json:"default_restart"`
	Configs        []ImportTypeConfig    `json:"configs"`
	Output         ImportContentVariable `json:"output"`
	Owner          string                `json:"owner"`
}

type Type string

type ImportContentVariable struct {
	Name                string                  `json:"name"`
	Type                Type                    `json:"type"`
	CharacteristicId    string                  `json:"characteristic_id"`
	SubContentVariables []ImportContentVariable `json:"sub_content_variables"`
	UseAsTag            bool                    `json:"use_as_tag"`
	FunctionId          string                  `json:"function_id,omitempty"`
	AspectId            string                  `json:"aspect_id,omitempty"`
	IsVoid              bool                    `json:"is_void"`
}

type ImportTypeConfig struct {
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	Type         Type        `json:"type"`
	DefaultValue interface{} `json:"default_value"`
}

type Import struct {
	Id           string         `json:"id"`
	Name         string         `json:"name"`
	ImportTypeId string         `json:"import_type_id"`
	Image        string         `json:"image"`
	KafkaTopic   string         `json:"kafka_topic"`
	Configs      []ImportConfig `json:"configs"`
	Restart      *bool          `json:"restart"`
}

type ImportConfig struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

type ImportTypeFilterCriteria struct {
	FunctionId string `json:"function_id"`
	AspectId   string `json:"aspect_id"`
}

func (this ImportTypeFilterCriteria) Short() string {
	return this.AspectId + "_" + this.FunctionId
}

func (this *ImportContentVariable) GetName() string {
	return this.Name
}

func (this *ImportContentVariable) GetCharacteristicId() string {
	return this.CharacteristicId
}

func (this *ImportContentVariable) GetSubContentVariables() []basecontentvariable.Descriptor {
	ls := make([]basecontentvariable.Descriptor, len(this.SubContentVariables))
	for idx := range this.SubContentVariables {
		ls[idx] = &this.SubContentVariables[idx]
	}
	return ls
}

func (this *ImportContentVariable) GetFunctionId() string {
	return this.FunctionId
}

func (this *ImportContentVariable) GetAspectId() string {
	return this.AspectId
}

func (this *ImportContentVariable) GetIsVoid() bool {
	return this.IsVoid
}
