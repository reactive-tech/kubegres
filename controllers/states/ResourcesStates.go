/*
Copyright 2021 Reactive Tech Limited.
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

package states

import (
	"reactive-tech.io/kubegres/controllers/ctx"
	"reactive-tech.io/kubegres/controllers/states/statefulset"
)

type ResourcesStates struct {
	DbStorageClass DbStorageClassStates
	StatefulSets   statefulset.StatefulSetsStates
	Services       ServicesStates
	Config         ConfigStates
	BackUp         BackUpStates

	kubegresContext ctx.KubegresContext
}

func LoadResourcesStates(kubegresContext ctx.KubegresContext) (ResourcesStates, error) {
	resourcesStates := ResourcesStates{kubegresContext: kubegresContext}
	err := resourcesStates.loadStates()
	return resourcesStates, err
}

func (r *ResourcesStates) loadStates() (err error) {
	err = r.loadDbStorageClassStates()
	if err != nil {
		return err
	}

	err = r.loadConfigStates()
	if err != nil {
		return err
	}

	err = r.loadStatefulSetsStates()
	if err != nil {
		return err
	}

	err = r.loadServicesStates()
	if err != nil {
		return err
	}

	err = r.loadBackUpStates()
	if err != nil {
		return err
	}

	return nil
}

func (r *ResourcesStates) loadDbStorageClassStates() (err error) {
	r.DbStorageClass, err = loadDbStorageClass(r.kubegresContext)
	return err
}

func (r *ResourcesStates) loadConfigStates() (err error) {
	r.Config, err = loadConfigStates(r.kubegresContext)
	return err
}

func (r *ResourcesStates) loadStatefulSetsStates() (err error) {
	r.StatefulSets, err = statefulset.LoadStatefulSetsStates(r.kubegresContext)
	return err
}

func (r *ResourcesStates) loadServicesStates() (err error) {
	r.Services, err = loadServicesStates(r.kubegresContext)
	return err
}

func (r *ResourcesStates) loadBackUpStates() (err error) {
	r.BackUp, err = loadBackUpStates(r.kubegresContext)
	return err
}
