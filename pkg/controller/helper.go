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

package controller

import "encoding/json"

func Clone[T any](orig T) (result T) {
	origJSON, err := json.Marshal(orig)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(origJSON, &result)
	if err != nil {
		panic(err)
	}
	return
}

func RemoveDuplicates[T comparable](slice []T) []T {
	keys := make(map[T]bool)
	result := []T{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			result = append(result, entry)
		}
	}
	return result
}

func RemoveDuplicatesF[T any, C comparable](slice []T, f func(T) C) []T {
	keys := make(map[C]bool)
	result := []T{}
	for _, entry := range slice {
		key := f(entry)
		if _, value := keys[key]; !value {
			keys[key] = true
			result = append(result, entry)
		}
	}
	return result
}
