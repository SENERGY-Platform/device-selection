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

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/device-selection/pkg/model"
	"github.com/SENERGY-Platform/device-selection/pkg/model/devicemodel"
	importrepo "github.com/SENERGY-Platform/import-repository/lib/client"
	importrepomodel "github.com/SENERGY-Platform/import-repository/lib/model"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
)

func (this *Controller) getFilteredImports(token string, descriptions model.FilterCriteriaAndSet) (result []model.Selectable, err error, code int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}

	criteria := []importrepo.ImportTypeFilterCriteria{}

	for _, c := range descriptions {
		importTypeCriteria := importrepo.ImportTypeFilterCriteria{
			FunctionId: c.FunctionId,
			AspectIds:  []string{c.AspectId},
		}
		aspect, err := this.GetAspectNode(c.AspectId, token)
		if err != nil {
			return result, err, http.StatusInternalServerError
		}
		for _, aid := range aspect.DescendentIds {
			importTypeCriteria.AspectIds = append(importTypeCriteria.AspectIds, aid)
		}
		criteria = append(criteria, importTypeCriteria)
	}

	importTypes, _, err, code := this.importrepo.ListImportTypes(jwtToken, importrepo.ImportTypeListOptions{
		Limit:    1000,
		Offset:   0,
		SortBy:   "name.asc",
		Criteria: criteria,
	})
	if err != nil {
		return result, err, code
	}
	importTypeIds := []string{}
	for _, importType := range importTypes {
		importTypeIds = append(importTypeIds, importType.Id)
	}

	if this.config.Debug {
		log.Println("DEBUG: getFilteredImports()::Found " + strconv.Itoa(len(importTypeIds)) + " matching import types")
	}

	instances, err, code := this.getImportsByTypes(token, importTypeIds)
	if err != nil {
		return result, err, code
	}
	if this.config.Debug {
		log.Println("DEBUG: getFilteredImports()::Found " + strconv.Itoa(len(instances)) + " matching import instances")
	}

	for _, instance := range instances {
		temp := instance //prevent that every result element becomes the last element of groups
		for _, importType := range importTypes {
			if importType.Id == temp.ImportTypeId {
				tempType := castImportType(importType)
				result = append(result, model.Selectable{Import: &temp, ImportType: &tempType})
			}
		}
	}
	return result, err, code
}

func (this *Controller) getFilteredImportsV2(token string, descriptions model.FilterCriteriaAndSet, importPathTrimFirstElement bool) (result []model.Selectable, err error, code int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	criteria := []importrepo.ImportTypeFilterCriteria{}
	for _, c := range descriptions {
		criteriaFilter := importrepo.ImportTypeFilterCriteria{
			FunctionId: c.FunctionId,
		}
		if c.AspectId != "" {
			criteriaFilter.AspectIds = []string{c.AspectId}
			aspect, err := this.GetAspectNode(c.AspectId, token)
			if err != nil {
				return result, err, http.StatusInternalServerError
			}
			for _, aid := range aspect.DescendentIds {
				criteriaFilter.AspectIds = append(criteriaFilter.AspectIds, aid)
			}
		}
		criteria = append(criteria, criteriaFilter)
	}
	importTypes, _, err, code := this.importrepo.ListImportTypes(jwtToken, importrepo.ImportTypeListOptions{
		Limit:    1000,
		Offset:   0,
		SortBy:   "name.asc",
		Criteria: criteria,
	})
	if err != nil {
		return result, err, code
	}
	importTypeIds := []string{}
	for _, importType := range importTypes {
		importTypeIds = append(importTypeIds, importType.Id)
	}

	if this.config.Debug {
		log.Println("DEBUG: getFilteredImports()::Found " + strconv.Itoa(len(importTypeIds)) + " matching import types")
	}

	instances, err, code := this.getImportsByTypes(token, importTypeIds)
	if err != nil {
		return result, err, code
	}
	if this.config.Debug {
		log.Println("DEBUG: getFilteredImports()::Found " + strconv.Itoa(len(instances)) + " matching import instances")
	}

	aspectCache := &map[string]devicemodel.AspectNode{}

	for _, instance := range instances {
		temp := instance //prevent that every result element becomes the last element of groups
		if temp.ImportTypeId != "" {
			fullType, err := this.getFullImportType(token, temp.ImportTypeId)
			if err != nil {
				return result, err, code
			}
			var pathOptions []model.PathOption
			if importPathTrimFirstElement {
				for _, sub := range fullType.Output.SubContentVariables {
					subOptions, err := this.getImportPathOptions(token, sub, descriptions, nil, aspectCache)
					if err != nil {
						return result, err, code
					}
					pathOptions = append(pathOptions, subOptions...)
				}
			} else {
				pathOptions, err = this.getImportPathOptions(token, fullType.Output, descriptions, nil, aspectCache)
				if err != nil {
					return result, err, code
				}
			}

			var pathOptionsMap map[string][]model.PathOption
			if pathOptions != nil && len(pathOptions) > 0 {
				pathOptionsMap = map[string][]model.PathOption{fullType.Id: pathOptions}
				result = append(result, model.Selectable{Import: &temp, ImportType: &fullType, ServicePathOptions: pathOptionsMap})
			}
		}
	}
	return result, nil, http.StatusOK
}

func (this *Controller) getImportPathOptions(token string, variable model.ImportContentVariable, criteria model.FilterCriteriaAndSet, currentPath []string, aspectCache *map[string]devicemodel.AspectNode) (result []model.PathOption, err error) {
	currentPath = append(currentPath, variable.Name)
	match, err := this.contentVariableContainsAnyCriteria(&variable, criteria, token, aspectCache)
	//match, err := this.importVariableMatchesAllCriteria(token, variable, criteria, aspectCache)
	if err != nil {
		return result, err
	}
	if match {
		result = append(result, model.PathOption{
			Path:             strings.Join(currentPath, "."),
			CharacteristicId: variable.CharacteristicId,
			AspectNode: devicemodel.AspectNode{
				Id: variable.AspectId,
			},
			FunctionId:  variable.FunctionId,
			IsVoid:      false,
			Type:        variable.Type,
			Interaction: devicemodel.EVENT,
		})
	}
	for _, sub := range variable.SubContentVariables {
		temp, err := this.getImportPathOptions(token, sub, criteria, currentPath, aspectCache)
		if err != nil {
			return result, err
		}
		result = append(result, temp...)
	}
	return result, nil
}

func (this *Controller) importVariableMatchesAllCriteria(token string, variable model.ImportContentVariable, criteria []devicemodel.FilterCriteria, cache *map[string]devicemodel.AspectNode) (match bool, err error) {
	for _, c := range criteria {
		match, err = this.contentVariableContainsCriteria(&variable, c, token, cache)
		if err != nil {
			return false, err
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}

func (this *Controller) getImportsByTypes(token string, typeIds []string) (result []model.Import, err error, code int) {
	req, err := http.NewRequest("GET", this.config.ImportDeployUrl+"/instances?&limit=10000&offset=0&sort=name.asc", nil)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		debug.PrintStack()
		return result, errors.New(buf.String()), resp.StatusCode
	}
	all := []model.Import{}
	err = json.NewDecoder(resp.Body).Decode(&all)
	if err != nil {
		debug.PrintStack()
		return result, err, http.StatusInternalServerError
	}

	for _, instance := range all {
		for _, typeId := range typeIds {
			if typeId == instance.ImportTypeId {
				result = append(result, instance)
				break
			}
		}
	}

	return result, nil, http.StatusOK
}

func (this *Controller) getFullImportType(token string, id string) (fullType model.ImportType, err error) {
	err = this.cache.Use(id, func() (interface{}, error) {
		var result model.ImportType
		req, err := http.NewRequest("GET", this.config.ImportRepoUrl+"/import-types/"+id, nil)
		if err != nil {
			debug.PrintStack()
			return result, err
		}
		req.Header.Set("Authorization", token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			debug.PrintStack()
			return result, err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			debug.PrintStack()
			return result, errors.New(buf.String())
		}

		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			debug.PrintStack()
			return result, err
		}

		return result, nil
	}, &fullType)

	return
}

func castImportType(importType importrepomodel.ImportType) model.ImportType {
	return model.ImportType{
		Id:             importType.Id,
		Name:           importType.Name,
		Description:    importType.Description,
		Image:          importType.Image,
		DefaultRestart: importType.DefaultRestart,
		Configs:        castImportTypeConfigs(importType.Configs),
		Output:         castImportTypeContentVariable(importType.Output),
		Owner:          importType.Owner,
	}
}

func castImportTypeContentVariable(cv importrepomodel.ContentVariable) model.ImportContentVariable {
	sub := []model.ImportContentVariable{}
	for _, s := range cv.SubContentVariables {
		sub = append(sub, castImportTypeContentVariable(s))
	}
	return model.ImportContentVariable{
		Name:                cv.Name,
		Type:                model.Type(cv.Type),
		CharacteristicId:    cv.CharacteristicId,
		SubContentVariables: sub,
		UseAsTag:            cv.UseAsTag,
		FunctionId:          cv.FunctionId,
		AspectId:            cv.AspectId,
	}
}

func castImportTypeConfigs(configs []importrepomodel.ImportConfig) (result []model.ImportTypeConfig) {
	if configs != nil {
		result = []model.ImportTypeConfig{}
	}
	for _, config := range configs {
		result = append(result, model.ImportTypeConfig{
			Name:         config.Name,
			Description:  config.Description,
			Type:         model.Type(config.Type),
			DefaultValue: config.DefaultValue,
		})
	}
	return result
}
