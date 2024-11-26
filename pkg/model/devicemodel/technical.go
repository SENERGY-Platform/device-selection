/*
 * Copyright 2019 InfAI (CC SES)
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

import (
	"github.com/SENERGY-Platform/models/go/models"
)

type Hub = models.Hub

type Content = models.Content

type Type = models.Type

const (
	String  Type = "https://schema.org/Text"
	Integer Type = "https://schema.org/Integer"
	Float   Type = "https://schema.org/Float"
	Boolean Type = "https://schema.org/Boolean"

	List      Type = "https://schema.org/ItemList"
	Structure Type = "https://schema.org/StructuredValue"
)

type ContentVariable = models.ContentVariable
