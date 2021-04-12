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

package comparator

import (
	core "k8s.io/api/core/v1"
	postgresV1 "reactive-tech.io/kubegres/api/v1"
)

type PodSpecComparator struct {
	Pod          core.Pod
	PostgresSpec postgresV1.KubegresSpec
}

func (r *PodSpecComparator) IsSpecUpToDate() bool {
	return r.isImageUpToDate() &&
		r.isPortUpToDate()
}

func (r *PodSpecComparator) isImageUpToDate() bool {
	current := r.Pod.Spec.Containers[0].Image
	expected := r.PostgresSpec.Image
	return current == expected
}

func (r *PodSpecComparator) isPortUpToDate() bool {
	current := r.Pod.Spec.Containers[0].Ports[0].ContainerPort
	expected := r.PostgresSpec.Port
	return current == expected
}
