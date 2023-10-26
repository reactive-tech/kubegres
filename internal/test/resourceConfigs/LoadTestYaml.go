/*
Copyright 2023 Reactive Tech Limited.
"Reactive Tech Limited" is a company located in England, United Kingdom.
https://www.reactive-tech.io

Lead Developer: Alex Arica

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resourceConfigs

import (
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"log"
	kubegresv1 "reactive-tech.io/kubegres/api/v1"
)

func LoadCustomConfigMapYaml(yamlFileName string) v1.ConfigMap {
	fileContents := getFileContents(yamlFileName)
	obj := decodeYaml(fileContents)
	return *obj.(*v1.ConfigMap)
}

func LoadBackUpPvcYaml() *v1.PersistentVolumeClaim {
	fileContents := getFileContents(BackUpPvcYamlFile)
	obj := decodeYaml(fileContents)
	return obj.(*v1.PersistentVolumeClaim)
}

func LoadKubegresYaml() *kubegresv1.Kubegres {
	fileContents := getFileContents(KubegresYamlFile)
	obj := decodeYaml(fileContents)
	return obj.(*kubegresv1.Kubegres)
}

func LoadSecretYaml() v1.Secret {
	fileContents := getFileContents(SecretYamlFile)
	obj := decodeYaml(fileContents)
	return *obj.(*v1.Secret)
}

func LoadYamlServiceToSqlQueryPrimaryDb() v1.Service {
	fileContents := getFileContents(ServiceToSqlQueryPrimaryDbYamlFile)
	obj := decodeYaml(fileContents)
	return *obj.(*v1.Service)
}

func LoadYamlServiceToSqlQueryReplicaDb() v1.Service {
	fileContents := getFileContents(ServiceToSqlQueryReplicaDbServiceYamlFile)
	obj := decodeYaml(fileContents)
	return *obj.(*v1.Service)
}

func getFileContents(filePath string) string {
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal("Unable to find file '"+filePath+"'. Given error: ", err)
	}
	return string(contents)
}

func decodeYaml(yamlContents string) runtime.Object {

	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, _, err := decode([]byte(yamlContents), nil, nil)

	if err != nil {
		log.Fatal("Error in decode:", obj, err)
	}

	return obj
}
