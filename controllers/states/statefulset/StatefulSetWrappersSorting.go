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

package statefulset

import (
	"strconv"
)

type SortByInstanceIndex []StatefulSetWrapper
type ReverseSortByInstanceIndex []StatefulSetWrapper

func (f SortByInstanceIndex) Len() int {
	return len(f)
}

func (f SortByInstanceIndex) Less(i, j int) bool {
	return getInstanceIndex(f[i]) < getInstanceIndex(f[j])
}

func (f SortByInstanceIndex) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f ReverseSortByInstanceIndex) Len() int {
	return len(f)
}

func (f ReverseSortByInstanceIndex) Less(i, j int) bool {
	return getInstanceIndex(f[i]) > getInstanceIndex(f[j])
}

func (f ReverseSortByInstanceIndex) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func getInstanceIndex(statefulSet StatefulSetWrapper) int32 {
	instanceIndexStr := statefulSet.StatefulSet.Spec.Template.Labels["index"]
	instanceIndex, _ := strconv.ParseInt(instanceIndexStr, 10, 32)
	return int32(instanceIndex)
}
