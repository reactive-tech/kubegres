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
	"k8s.io/apimachinery/pkg/api/resource"
	"reactive-tech.io/kubegres/controllers/ctx"
)

type ResourcesSpecEnforcer struct {
	kubegresContext ctx.KubegresContext
}

func CreateResourcesSpecEnforcer(kubegresContext ctx.KubegresContext) ResourcesSpecEnforcer {
	return ResourcesSpecEnforcer{kubegresContext: kubegresContext}
}

func (r *ResourcesSpecEnforcer) GetSpecName() string {
	return "Resources"
}

func (r *ResourcesSpecEnforcer) CheckForSpecDifference(statefulSet *apps.StatefulSet) StatefulSetSpecDifference {

	current := statefulSet.Spec.Template.Spec.Containers[0].Resources
	expected := r.kubegresContext.Kubegres.Spec.Resources

	if !r.compareResources(current, expected) {
		return StatefulSetSpecDifference{
			SpecName: r.GetSpecName(),
			Current:  current.String(),
			Expected: expected.String(),
		}
	}

	return StatefulSetSpecDifference{}
}

func (r *ResourcesSpecEnforcer) EnforceSpec(statefulSet *apps.StatefulSet) (wasSpecUpdated bool, err error) {
	statefulSet.Spec.Template.Spec.Containers[0].Resources = r.kubegresContext.Kubegres.Spec.Resources
	return true, nil
}

func (r *ResourcesSpecEnforcer) OnSpecEnforcedSuccessfully(statefulSet *apps.StatefulSet) error {
	return nil
}

func (r *ResourcesSpecEnforcer) compareResources(currentRequirement v1.ResourceRequirements,
	expectedRequirement v1.ResourceRequirements) bool {

	if !r.compareResourceLists(currentRequirement.Limits, expectedRequirement.Limits) {
		return false
	}

	if !r.compareResourceLists(currentRequirement.Requests, expectedRequirement.Requests) {
		return false
	}

	return true
}

func (r *ResourcesSpecEnforcer) compareResourceLists(current v1.ResourceList, expected v1.ResourceList) bool {
	if len(current) != len(expected) {
		return false
	}

	for key, expectedElement := range expected {
		expectedElementParsed := resource.MustParse(expectedElement.String())
		currentElement := current[key]
		if expectedElementParsed != currentElement {
			return false
		}
	}
	return true
}
