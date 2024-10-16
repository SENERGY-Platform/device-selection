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

package environment

import (
	"context"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/environment/docker"
	"github.com/SENERGY-Platform/device-selection/pkg/tests/environment/mock"
	"log"
	"runtime/debug"
	"sync"
)

func NewWithImport(ctx context.Context, wg *sync.WaitGroup) (kafkaBroker string, deviceManagerUrl string, deviceRepoUrl string, permSearchUrl string, permv2Url string, importRepoUrl string, importDeployUrl string, err error) {
	kafkaBroker, deviceManagerUrl, deviceRepoUrl, permSearchUrl, permv2Url, err = docker.DeviceManagerWithDependenciesAndKafka(ctx, wg)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return "", "", "", "", "", "", "", err
	}
	_, mongoIp, err := docker.MongoDB(ctx, wg)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return "", "", "", "", "", "", "", err
	}

	importMongo := "mongodb://" + mongoIp + ":27017"
	_, importRepoIp, err := docker.ImportRepo(ctx, wg, kafkaBroker, importMongo, permSearchUrl, deviceRepoUrl, permv2Url)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return "", "", "", "", "", "", "", err
	}

	importRepoUrl = "http://" + importRepoIp + ":8080"
	importDeploy := mock.NewImportDeploy()

	go func() {
		<-ctx.Done()
		importDeploy.Stop()
	}()

	importDeployUrl = importDeploy.Url()

	return
}

const ConceptTopic = "concepts"
const FunctionTopic = "functions"
const ImportTypeTopic = "import-types"
