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

package resources_count_spec

type ResourceCountSpecEnforcer interface {
	EnforceSpec() error
}

type ResourcesCountSpecEnforcer struct {
	registry []ResourceCountSpecEnforcer
}

func (r *ResourcesCountSpecEnforcer) AddSpecEnforcer(specEnforcer ResourceCountSpecEnforcer) {
	r.registry = append(r.registry, specEnforcer)
}

func (r *ResourcesCountSpecEnforcer) EnforceSpec() error {
	for _, specEnforcer := range r.registry {
		if err := specEnforcer.EnforceSpec(); err != nil {
			return err
		}
	}
	return nil
}
