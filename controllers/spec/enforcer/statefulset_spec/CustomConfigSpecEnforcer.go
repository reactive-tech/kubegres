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

package statefulset_spec

import (
	apps "k8s.io/api/apps/v1"
	"reactive-tech.io/kubegres/controllers/spec/template"
)

type CustomConfigSpecEnforcer struct {
	customConfigSpecHelper template.CustomConfigSpecHelper
}

func CreateCustomConfigSpecEnforcer(customConfigSpecHelper template.CustomConfigSpecHelper) CustomConfigSpecEnforcer {
	return CustomConfigSpecEnforcer{customConfigSpecHelper: customConfigSpecHelper}
}

func (r *CustomConfigSpecEnforcer) GetSpecName() string {
	return "CustomConfig"
}

func (r *CustomConfigSpecEnforcer) CheckForSpecDifference(statefulSet *apps.StatefulSet) StatefulSetSpecDifference {

	statefulSetCopy := *statefulSet
	hasStatefulSetChanged, changesDetails := r.customConfigSpecHelper.ConfigureStatefulSet(&statefulSetCopy)

	if hasStatefulSetChanged {
		return StatefulSetSpecDifference{
			SpecName: r.GetSpecName(),
			Current:  " ",
			Expected: changesDetails,
		}
	}

	return StatefulSetSpecDifference{}
}

func (r *CustomConfigSpecEnforcer) EnforceSpec(statefulSet *apps.StatefulSet) (wasSpecUpdated bool, err error) {
	wasSpecUpdated, _ = r.customConfigSpecHelper.ConfigureStatefulSet(statefulSet)
	return wasSpecUpdated, nil
}

func (r *CustomConfigSpecEnforcer) OnSpecEnforcedSuccessfully(statefulSet *apps.StatefulSet) error {
	return nil
}
