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
	"device-selection/pkg/tests/environment/docker"
	"device-selection/pkg/tests/environment/kafka"
	"device-selection/pkg/tests/environment/mock"
	"log"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

func New(ctx context.Context, wg *sync.WaitGroup) (kafkaBroker string, deviceManagerUrl string, semanticUrl string, deviceRepoUrl string, permSearchUrl string, importRepoUrl string, importDeployUrl string, err error) {
	_, zk, err := docker.Zookeeper(ctx, wg)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return "", "", "", "", "", "", "", err
	}
	zkUrl := zk + ":2181"

	kafkaPort, err := docker.Kafka(ctx, wg, zkUrl)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return "", "", "", "", "", "", "", err
	}
	kafkaBroker = "127.0.0.1:" + strconv.Itoa(kafkaPort)

	_, elasticIp, err := docker.ElasticSearch(ctx, wg)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return "", "", "", "", "", "", "", err
	}

	_, permIp, err := docker.PermSearch(ctx, wg, zkUrl, elasticIp)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return "", "", "", "", "", "", "", err
	}
	permSearchUrl = "http://" + permIp + ":8080"

	time.Sleep(2 * time.Second)

	semantic := mock.NewSemanticRepo(mock.NewConsumer(ctx, zkUrl, "semantic"))
	deviceRepo := mock.NewDeviceRepo(mock.NewConsumer(ctx, zkUrl, "devicerepo"))
	importRepoProducer, err := kafka.GetProducer([]string{kafkaBroker}, "import-types")
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return "", "", "", "", "", "", "", err
	}
	importRepo := mock.NewImportRepo(importRepoProducer)
	importDeploy := mock.NewImportDeploy()

	go func() {
		<-ctx.Done()
		semantic.Stop()
		deviceRepo.Stop()
		importRepo.Stop()
		importDeploy.Stop()
	}()

	semanticUrl = semantic.Url()
	deviceRepoUrl = deviceRepo.Url()
	importRepoUrl = importRepo.Url()
	importDeployUrl = importDeploy.Url()

	hostIp, err := docker.GetHostIp()
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return "", "", "", "", "", "", "", err
	}

	//transform local-address to address in docker container
	semanticUrlStruct := strings.Split(semantic.Url(), ":")
	semanticUrl = "http://" + hostIp + ":" + semanticUrlStruct[len(semanticUrlStruct)-1]
	log.Println("DEBUG: semantic url transformation:", semantic.Url(), "-->", semanticUrl)

	//transform local-address to address in docker container
	deviceRepoStruct := strings.Split(deviceRepo.Url(), ":")
	deviceRepoUrl = "http://" + hostIp + ":" + deviceRepoStruct[len(deviceRepoStruct)-1]
	log.Println("DEBUG: device-repo url transformation:", deviceRepo.Url(), "-->", deviceRepoUrl)

	_, managerIp, err := docker.DeviceManager(ctx, wg, zkUrl, semanticUrl, deviceRepoUrl, permSearchUrl)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return "", "", "", "", "", "", "", err
	}

	deviceManagerUrl = "http://" + managerIp + ":8080"

	return
}
