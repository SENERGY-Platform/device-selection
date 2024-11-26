/*
 * Copyright 2021 InfAI (CC SES)
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

import (
	"errors"
	"github.com/SENERGY-Platform/device-selection/pkg/model/devicemodel"
	"sync"
)

func (this *Controller) GetFunction(id string, token string) (f devicemodel.Function, err error) {
	functions, err := this.GetFunctions(token)
	if err != nil {
		return
	}
	for _, function := range functions {
		if function.Id == id {
			return function, nil
		}
	}
	return f, errors.New("not found")
}

func (this *Controller) GetFunctions(token string) (functions []devicemodel.Function, err error) {
	err = this.cache.Use("functions", func() (interface{}, error) {
		mux := sync.Mutex{}
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			controllfunctions, localErr, _ := this.devicerepo.GetFunctionsByType(devicemodel.SES_ONTOLOGY_CONTROLLING_FUNCTION)
			mux.Lock()
			defer mux.Unlock()
			if err != nil {
				err = errors.Join(err, localErr)
			} else {
				functions = append(functions, controllfunctions...)
			}
		}()
		go func() {
			defer wg.Done()
			measuringfunctions, localErr, _ := this.devicerepo.GetFunctionsByType(devicemodel.SES_ONTOLOGY_MEASURING_FUNCTION)
			mux.Lock()
			defer mux.Unlock()
			if err != nil {
				err = errors.Join(err, localErr)
			} else {
				functions = append(functions, measuringfunctions...)
			}
		}()
		wg.Wait()
		return functions, err
	}, &functions)
	return
}
