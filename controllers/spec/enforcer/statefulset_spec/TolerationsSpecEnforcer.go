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
	v1 "k8s.io/api/core/v1"
	"reactive-tech.io/kubegres/controllers/ctx"
	"reflect"
)

type TolerationsSpecEnforcer struct {
	kubegresContext ctx.KubegresContext
}

func CreateTolerationsSpecEnforcer(kubegresContext ctx.KubegresContext) TolerationsSpecEnforcer {
	return TolerationsSpecEnforcer{kubegresContext: kubegresContext}
}

func (r *TolerationsSpecEnforcer) GetSpecName() string {
	return "Tolerations"
}

func (r *TolerationsSpecEnforcer) CheckForSpecDifference(statefulSet *apps.StatefulSet) StatefulSetSpecDifference {
	current := statefulSet.Spec.Template.Spec.Tolerations

	var expected []v1.Toleration
	instance := r.kubegresContext.GetInstanceFromStatefulSet(*statefulSet)
	nodeSet := r.kubegresContext.GetNodeSetSpecFromInstance(instance)

	if nodeSet.Tolerations != nil {
		expected = nodeSet.Tolerations
	} else {
		expected = r.kubegresContext.Kubegres.Spec.Scheduler.Tolerations
	}

	if !r.compare(current, expected) {
		return StatefulSetSpecDifference{
			SpecName: r.GetSpecName(),
			Current:  r.toString(current),
			Expected: r.toString(expected),
		}
	}

	return StatefulSetSpecDifference{}
}

func (r *TolerationsSpecEnforcer) EnforceSpec(statefulSet *apps.StatefulSet) (wasSpecUpdated bool, err error) {
	instance := r.kubegresContext.GetInstanceFromStatefulSet(*statefulSet)
	nodeSet := r.kubegresContext.GetNodeSetSpecFromInstance(instance)

	if nodeSet.Tolerations != nil {
		statefulSet.Spec.Template.Spec.Tolerations = nodeSet.Tolerations
	} else {
		statefulSet.Spec.Template.Spec.Tolerations = r.kubegresContext.Kubegres.Spec.Scheduler.Tolerations
	}

	return true, nil
}

func (r *TolerationsSpecEnforcer) OnSpecEnforcedSuccessfully(statefulSet *apps.StatefulSet) error {
	return nil
}

func (r *TolerationsSpecEnforcer) compare(current []v1.Toleration, expected []v1.Toleration) bool {

	if len(current) != len(expected) {
		return false
	}

	index := 0
	for _, expectedItem := range expected {
		if !reflect.DeepEqual(expectedItem, current[index]) {
			return false
		}
		index++
	}

	return true
}

func (r *TolerationsSpecEnforcer) toString(tolerations []v1.Toleration) string {

	toString := ""
	for _, toleration := range tolerations {
		toString += toleration.String() + " - "
	}
	return toString
}
