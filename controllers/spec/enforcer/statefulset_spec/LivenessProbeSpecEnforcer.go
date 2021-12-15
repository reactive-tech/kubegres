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
	"reactive-tech.io/kubegres/controllers/ctx"
	"reactive-tech.io/kubegres/controllers/states"
	"reflect"
)

type LivenessProbeSpecEnforcer struct {
	kubegresContext ctx.KubegresContext
	resourcesStates states.ResourcesStates
}

func CreateLivenessProbeSpecEnforcer(kubegresContext ctx.KubegresContext) LivenessProbeSpecEnforcer {
	return LivenessProbeSpecEnforcer{kubegresContext: kubegresContext}
}

func (r *LivenessProbeSpecEnforcer) GetSpecName() string {
	return "LivenessProbe"
}

func (r *LivenessProbeSpecEnforcer) CheckForSpecDifference(statefulSet *apps.StatefulSet) StatefulSetSpecDifference {
	current := statefulSet.Spec.Template.Spec.Containers[0].LivenessProbe
	expected := r.kubegresContext.Kubegres.Spec.LivenessProbe

	if expected == nil {
		return StatefulSetSpecDifference{}
	}

	if !reflect.DeepEqual(current, expected) {
		return StatefulSetSpecDifference{
			SpecName: r.GetSpecName(),
			Current:  current.String(),
			Expected: expected.String(),
		}
	}

	return StatefulSetSpecDifference{}
}

func (r *LivenessProbeSpecEnforcer) EnforceSpec(statefulSet *apps.StatefulSet) (wasSpecUpdated bool, err error) {
	statefulSet.Spec.Template.Spec.Containers[0].LivenessProbe = r.kubegresContext.Kubegres.Spec.LivenessProbe
	return true, nil
}

func (r *LivenessProbeSpecEnforcer) OnSpecEnforcedSuccessfully(statefulSet *apps.StatefulSet) error {
	return nil
}
