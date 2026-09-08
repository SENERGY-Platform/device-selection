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

package devicemodel

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestFilterCriteriaGetAspectIds(t *testing.T) {
	tests := []struct {
		name     string
		criteria FilterCriteria
		expected []string
	}{
		{"no aspect", FilterCriteria{FunctionId: "f"}, nil},
		{"deprecated alias only", FilterCriteria{AspectId: "a"}, []string{"a"}},
		{"list only", FilterCriteria{AspectIds: []string{"a", "b"}}, []string{"a", "b"}},
		{"alias already in list", FilterCriteria{AspectId: "a", AspectIds: []string{"a", "b"}}, []string{"a", "b"}},
		{"alias next to list", FilterCriteria{AspectId: "c", AspectIds: []string{"a", "b"}}, []string{"a", "b", "c"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.criteria.GetAspectIds()
			if !reflect.DeepEqual(actual, test.expected) {
				t.Error(actual, test.expected)
			}
		})
	}
}

// the deprecated alias may not be appended to the list the caller passed in, which would
// change a criteria the caller still holds
func TestAspectIdsDoesNotModifyInput(t *testing.T) {
	input := make([]string, 1, 2)
	input[0] = "a"
	AspectIds("b", input)
	if !reflect.DeepEqual(input, []string{"a"}) {
		t.Error(input)
	}
}

func TestFilterCriteriaShort(t *testing.T) {
	//the same set of aspects has to render the same key, independent of the order it was listed in
	ab := FilterCriteria{FunctionId: "f", AspectIds: []string{"a", "b"}}.Short()
	ba := FilterCriteria{FunctionId: "f", AspectIds: []string{"b", "a"}}.Short()
	if ab != ba {
		t.Error(ab, ba)
	}
	//the deprecated alias renders like the single element list it stands for
	if single, list := (FilterCriteria{FunctionId: "f", AspectId: "a"}).Short(), (FilterCriteria{FunctionId: "f", AspectIds: []string{"a"}}).Short(); single != list {
		t.Error(single, list)
	}
	if ab == (FilterCriteria{FunctionId: "f", AspectIds: []string{"a"}}).Short() {
		t.Error("criteria over two aspects renders like the one over a single aspect")
	}
}

func TestFilterCriteriaJson(t *testing.T) {
	b, err := json.Marshal(FilterCriteria{FunctionId: "f", AspectIds: []string{"a", "b"}})
	if err != nil {
		t.Error(err)
		return
	}
	if string(b) != `{"function_id":"f","aspect_id":"","aspect_ids":["a","b"],"device_class_id":""}` {
		t.Error(string(b))
	}

	criteria := FilterCriteria{}
	err = json.Unmarshal([]byte(`{"function_id":"f","aspect_ids":["a","b"]}`), &criteria)
	if err != nil {
		t.Error(err)
		return
	}
	if !reflect.DeepEqual(criteria, FilterCriteria{FunctionId: "f", AspectIds: []string{"a", "b"}}) {
		t.Error(criteria)
	}
}
