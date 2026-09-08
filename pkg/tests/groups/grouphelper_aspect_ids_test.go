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

package groups

import (
	"context"
	"encoding/json"
	"maps"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/SENERGY-Platform/device-selection/pkg/client"
	"github.com/SENERGY-Platform/device-selection/pkg/model/devicemodel"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/helper"
)

// A content variable may carry several aspects, and the criteria of a device record that: one
// over the whole aspect list of the variable, plus one per single aspect. POST
// /device-group-helper intersects the criteria of its devices by Short(), so the aspects a
// group keeps for one function are the intersection of the aspects of its devices - a device
// with [a b] and one with [b c] leave the group with b alone, because b is the only aspect
// that stands as a criteria of its own on both sides.
func TestGroupHelperAspectIds(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	getTemperature := devicemodel.MEASURING_FUNCTION_PREFIX + "getTemperature"

	//flat aspects without a hierarchy, so that no ancestor criteria join the result
	aspectA := "urn:infai:ses:aspect:a"
	aspectB := "urn:infai:ses:aspect:b"
	aspectC := "urn:infai:ses:aspect:c"
	aspectD := "urn:infai:ses:aspect:d"

	//and a separate hierarchy for the cases below: q is a descendant of p, r stands beside it.
	//These are their own aspects, so that the flat cases above are unaffected by the hierarchy.
	aspectP := "urn:infai:ses:aspect:p"
	aspectQ := "urn:infai:ses:aspect:q"
	aspectR := "urn:infai:ses:aspect:r"

	aspects := []devicemodel.Aspect{
		{Id: aspectA, Name: "a"},
		{Id: aspectB, Name: "b"},
		{Id: aspectC, Name: "c"},
		{Id: aspectD, Name: "d"},
		{Id: aspectP, Name: "p", SubAspects: []devicemodel.Aspect{{Id: aspectQ, Name: "q"}}},
		{Id: aspectR, Name: "r"},
	}
	functions := []devicemodel.Function{{Id: getTemperature}}

	//one device-type per aspect set, each with a single service measuring the same function
	deviceTypeAspects := map[string][]string{
		"dt_ab":       {aspectA, aspectB},
		"dt_bc":       {aspectB, aspectC},
		"dt_ab_again": {aspectA, aspectB},
		"dt_d":        {aspectD},
		"dt_pr":       {aspectP, aspectR},
		"dt_qr":       {aspectQ, aspectR},
	}
	deviceTypes := []devicemodel.DeviceType{}
	devices := []devicemodel.Device{}
	for _, deviceTypeId := range slices.Sorted(maps.Keys(deviceTypeAspects)) {
		deviceTypes = append(deviceTypes, devicemodel.DeviceType{
			Id:            deviceTypeId,
			Name:          deviceTypeId,
			DeviceClassId: "urn:infai:ses:device-class:thermometer",
			Services: []devicemodel.Service{{
				Id:          deviceTypeId + "_service",
				Name:        deviceTypeId + "_service",
				Interaction: devicemodel.REQUEST,
				Outputs: []devicemodel.Content{{
					Id: deviceTypeId + "_output",
					ContentVariable: devicemodel.ContentVariable{
						Id:               deviceTypeId + "_variable",
						Name:             "temperature",
						FunctionId:       getTemperature,
						AspectIds:        deviceTypeAspects[deviceTypeId],
						CharacteristicId: "urn:infai:ses:characteristic:temperature",
					},
				}},
			}},
		})
		deviceId := strings.Replace(deviceTypeId, "dt_", "device_", 1)
		devices = append(devices, devicemodel.Device{
			Id:           deviceId,
			LocalId:      deviceId,
			Name:         deviceId,
			DeviceTypeId: deviceTypeId,
			OwnerId:      helper.JwtSubject,
		})
	}

	_, _, _, selectionurl, err := helper.EnvWithMetadata(ctx, wg, deviceTypes, devices, aspects, functions)
	if err != nil {
		t.Error(err)
		return
	}
	c := client.NewClient(selectionurl)

	measuring := func(aspectIds ...string) devicemodel.DeviceGroupFilterCriteria {
		return devicemodel.DeviceGroupFilterCriteria{
			FunctionId:  getTemperature,
			AspectId:    aspectIds[0], //the alphabetically first, which is the deprecated alias
			AspectIds:   aspectIds,
			Interaction: devicemodel.REQUEST,
		}
	}

	//a single device keeps the criteria over its whole aspect list next to the single aspects
	t.Run("one device with two aspects", testGroupHelperAspectIds(c, []string{"device_ab"}, []devicemodel.DeviceGroupFilterCriteria{
		measuring(aspectA, aspectB),
		measuring(aspectA),
		measuring(aspectB),
	}))

	t.Run("one device with one aspect", testGroupHelperAspectIds(c, []string{"device_d"}, []devicemodel.DeviceGroupFilterCriteria{
		measuring(aspectD),
	}))

	//the case this test exists for: [a b] and [b c] leave b
	t.Run("intersection keeps only the common aspect", testGroupHelperAspectIds(c, []string{"device_ab", "device_bc"}, []devicemodel.DeviceGroupFilterCriteria{
		measuring(aspectB),
	}))

	t.Run("device order does not matter", testGroupHelperAspectIds(c, []string{"device_bc", "device_ab"}, []devicemodel.DeviceGroupFilterCriteria{
		measuring(aspectB),
	}))

	//two devices carrying the same pair keep the pair, because the list criteria survives too
	t.Run("intersection of the same aspect pair keeps the pair", testGroupHelperAspectIds(c, []string{"device_ab", "device_ab_again"}, []devicemodel.DeviceGroupFilterCriteria{
		measuring(aspectA, aspectB),
		measuring(aspectA),
		measuring(aspectB),
	}))

	//the aspect list of a variable is not an alternative: [a b] does not cover a group member
	//that only measures d, so the group is left without a criteria for the shared function
	t.Run("no common aspect", testGroupHelperAspectIds(c, []string{"device_ab", "device_d"}, []devicemodel.DeviceGroupFilterCriteria{}))

	t.Run("three devices narrow to the aspect all of them carry", testGroupHelperAspectIds(c, []string{"device_ab", "device_bc", "device_ab_again"}, []devicemodel.DeviceGroupFilterCriteria{
		measuring(aspectB),
	}))

	//An aspect criteria covers the subtree of its node, so a variable carrying [q r] is found
	//by a query over [p r] as well. The criteria of that device therefore hold [p r] next to
	//[q r], and the ancestor of a single aspect stands on its own the way it always did.
	t.Run("one device with a descendant aspect", testGroupHelperAspectIds(c, []string{"device_qr"}, []devicemodel.DeviceGroupFilterCriteria{
		measuring(aspectQ, aspectR),
		measuring(aspectP, aspectR),
		measuring(aspectQ),
		measuring(aspectP),
		measuring(aspectR),
	}))

	t.Run("one device with the ancestor aspect", testGroupHelperAspectIds(c, []string{"device_pr"}, []devicemodel.DeviceGroupFilterCriteria{
		measuring(aspectP, aspectR),
		measuring(aspectP),
		measuring(aspectR),
	}))

	//the case a hierarchy adds: a device carrying [p r] and one carrying [q r] keep [p r],
	//because p covers q. Without the ancestor of the list, the group would fall back to the
	//single p and r and lose that one variable carries both.
	t.Run("intersection over a descendant keeps the ancestor pair", testGroupHelperAspectIds(c, []string{"device_pr", "device_qr"}, []devicemodel.DeviceGroupFilterCriteria{
		measuring(aspectP, aspectR),
		measuring(aspectP),
		measuring(aspectR),
	}))

	t.Run("intersection over a descendant, device order reversed", testGroupHelperAspectIds(c, []string{"device_qr", "device_pr"}, []devicemodel.DeviceGroupFilterCriteria{
		measuring(aspectP, aspectR),
		measuring(aspectP),
		measuring(aspectR),
	}))
}

func testGroupHelperAspectIds(c client.Client, deviceIds []string, expectedCriteria []devicemodel.DeviceGroupFilterCriteria) func(t *testing.T) {
	return func(t *testing.T) {
		result, code, err := c.DeviceGroupHelper(helper.AdminJwt, deviceIds, nil)
		if err != nil {
			t.Error(err, code)
			return
		}
		actual := sortCriteriaByShort(result.Criteria)
		expected := sortCriteriaByShort(expectedCriteria)
		if !reflect.DeepEqual(actual, expected) {
			actualJson, _ := json.Marshal(actual)
			expectedJson, _ := json.Marshal(expected)
			t.Error("\na=", string(actualJson), "\ne=", string(expectedJson))
		}
	}
}

// sortCriteriaByShort orders by the key the intersection uses, which is total: sorting by
// AspectId alone leaves the criteria over an aspect list next to the one over its first
// aspect in the order the answer happened to carry them.
func sortCriteriaByShort(criteria []devicemodel.DeviceGroupFilterCriteria) []devicemodel.DeviceGroupFilterCriteria {
	result := slices.Clone(criteria)
	slices.SortFunc(result, func(a, b devicemodel.DeviceGroupFilterCriteria) int {
		return strings.Compare(a.Short(), b.Short())
	})
	return result
}
