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

package statefulset_spec

import (
	apps "k8s.io/api/apps/v1"
	"reactive-tech.io/kubegres/internal/controller/ctx"
)

type ServiceAccountNameSpecEnforcer struct {
	kubegresContext ctx.KubegresContext
}

func CreateServiceAccountNameSpecEnforcer(kubegresContext ctx.KubegresContext) ServiceAccountNameSpecEnforcer {
	return ServiceAccountNameSpecEnforcer{kubegresContext: kubegresContext}
}

func (r *ServiceAccountNameSpecEnforcer) GetSpecName() string {
	return "ServiceAccountName"
}

func (r *ServiceAccountNameSpecEnforcer) CheckForSpecDifference(statefulSet *apps.StatefulSet) StatefulSetSpecDifference {

	current := statefulSet.Spec.Template.Spec.ServiceAccountName
	expected := r.kubegresContext.Kubegres.Spec.ServiceAccountName

	if current != expected {
		return StatefulSetSpecDifference{
			SpecName: r.GetSpecName(),
			Current:  current,
			Expected: expected,
		}
	}

	return StatefulSetSpecDifference{}
}

func (r *ServiceAccountNameSpecEnforcer) EnforceSpec(statefulSet *apps.StatefulSet) (wasSpecUpdated bool, err error) {
	statefulSet.Spec.Template.Spec.ServiceAccountName = r.kubegresContext.Kubegres.Spec.ServiceAccountName
	return true, nil
}

func (r *ServiceAccountNameSpecEnforcer) OnSpecEnforcedSuccessfully(_ *apps.StatefulSet) error {
	return nil
}
