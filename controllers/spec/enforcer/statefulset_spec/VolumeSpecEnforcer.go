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

type VolumeSpecEnforcer struct {
	kubegresContext ctx.KubegresContext
}

func CreateVolumeSpecEnforcer(kubegresContext ctx.KubegresContext) VolumeSpecEnforcer {
	return VolumeSpecEnforcer{kubegresContext: kubegresContext}
}

func (r *VolumeSpecEnforcer) GetSpecName() string {
	return "Volume"
}

func (r *VolumeSpecEnforcer) CheckForSpecDifference(statefulSet *apps.StatefulSet) StatefulSetSpecDifference {

	currentCustomVolumes := r.getCurrentCustomVolumes(statefulSet)
	expectedCustomVolumes := r.kubegresContext.Kubegres.Spec.Volume.Volumes

	if !r.compareVolumes(currentCustomVolumes, expectedCustomVolumes) {
		return StatefulSetSpecDifference{
			SpecName: "Volume.Volumes",
			Current:  r.volumesToString(currentCustomVolumes),
			Expected: r.volumesToString(expectedCustomVolumes),
		}
	}

	currentCustomVolumeMounts := r.getCurrentCustomVolumeMounts(statefulSet)
	expectedCustomVolumeMounts := r.kubegresContext.Kubegres.Spec.Volume.VolumeMounts

	if !r.compareVolumeMounts(currentCustomVolumeMounts, expectedCustomVolumeMounts) {
		return StatefulSetSpecDifference{
			SpecName: "Volume.VolumeMounts",
			Current:  r.volumeMountsToString(currentCustomVolumeMounts),
			Expected: r.volumeMountsToString(expectedCustomVolumeMounts),
		}
	}

	return StatefulSetSpecDifference{}
}

func (r *VolumeSpecEnforcer) EnforceSpec(statefulSet *apps.StatefulSet) (wasSpecUpdated bool, err error) {

	r.removeCustomVolumes(statefulSet)
	r.removeCustomVolumeMounts(statefulSet)

	if r.kubegresContext.Kubegres.Spec.Volume.Volumes != nil {
		statefulSet.Spec.Template.Spec.Volumes = append(statefulSet.Spec.Template.Spec.Volumes, r.kubegresContext.Kubegres.Spec.Volume.Volumes...)
	}

	if statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts != nil {
		statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts = append(statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts, r.kubegresContext.Kubegres.Spec.Volume.VolumeMounts...)
	}

	return true, nil
}

func (r *VolumeSpecEnforcer) OnSpecEnforcedSuccessfully(statefulSet *apps.StatefulSet) error {
	return nil
}

func (r *VolumeSpecEnforcer) compareVolumes(currentCustomVolumes []v1.Volume, expectedCustomVolumes []v1.Volume) bool {

	if len(expectedCustomVolumes) != len(currentCustomVolumes) {
		return false
	}

	for _, expectedCustomVolume := range expectedCustomVolumes {
		if !r.doesExpectedVolumeExist(expectedCustomVolume, currentCustomVolumes) {
			return false
		}
	}
	return true
}

func (r *VolumeSpecEnforcer) doesExpectedVolumeExist(expectedCustomVolume v1.Volume, currentCustomVolumes []v1.Volume) bool {
	for _, currentCustomVolume := range currentCustomVolumes {
		if reflect.DeepEqual(expectedCustomVolume, currentCustomVolume) {
			return true
		}
	}
	return false
}

func (r *VolumeSpecEnforcer) compareVolumeMounts(currentCustomVolumeMounts []v1.VolumeMount, expectedCustomVolumeMounts []v1.VolumeMount) bool {

	if len(expectedCustomVolumeMounts) != len(currentCustomVolumeMounts) {
		return false
	}

	for _, expectedCustomVolumeMount := range expectedCustomVolumeMounts {
		if !r.doesExpectedVolumeMountExist(expectedCustomVolumeMount, currentCustomVolumeMounts) {
			return false
		}
	}
	return true
}

func (r *VolumeSpecEnforcer) doesExpectedVolumeMountExist(expectedCustomVolumeMount v1.VolumeMount, currentCustomVolumeMounts []v1.VolumeMount) bool {
	for _, currentCustomVolumeMount := range currentCustomVolumeMounts {
		if reflect.DeepEqual(expectedCustomVolumeMount, currentCustomVolumeMount) {
			return true
		}
	}
	return false
}

func (r *VolumeSpecEnforcer) getCurrentCustomVolumes(statefulSet *apps.StatefulSet) []v1.Volume {

	statefulSetSpec := &statefulSet.Spec.Template.Spec
	var customVolumes []v1.Volume

	for _, volume := range statefulSetSpec.Volumes {
		if !r.kubegresContext.IsReservedVolumeName(volume.Name) {
			customVolumes = append(customVolumes, volume)
		}
	}
	return customVolumes
}

func (r *VolumeSpecEnforcer) getCurrentCustomVolumeMounts(statefulSet *apps.StatefulSet) []v1.VolumeMount {

	container := &statefulSet.Spec.Template.Spec.Containers[0]
	var customVolumeMounts []v1.VolumeMount

	for _, volumeMount := range container.VolumeMounts {
		if !r.kubegresContext.IsReservedVolumeName(volumeMount.Name) {
			customVolumeMounts = append(customVolumeMounts, volumeMount)
		}
	}
	return customVolumeMounts
}

func (r *VolumeSpecEnforcer) removeCustomVolumes(statefulSet *apps.StatefulSet) {

	currentCustomVolumes := r.getCurrentCustomVolumes(statefulSet)
	if len(currentCustomVolumes) == 0 {
		return
	}

	currentCustomVolumesCopy := make([]v1.Volume, len(currentCustomVolumes))
	copy(currentCustomVolumesCopy, currentCustomVolumes)

	statefulSetSpec := &statefulSet.Spec.Template.Spec

	for _, customVolume := range currentCustomVolumesCopy {
		index := r.getIndexOfVolume(customVolume, statefulSetSpec.Volumes)
		if index >= 0 {
			statefulSetSpec.Volumes = append(statefulSetSpec.Volumes[:index], statefulSetSpec.Volumes[index+1:]...)
		}
	}
}

func (r *VolumeSpecEnforcer) getIndexOfVolume(volumeToSearch v1.Volume, volumes []v1.Volume) int {
	index := 0
	for _, volume := range volumes {
		if reflect.DeepEqual(volume, volumeToSearch) {
			return index
		}
		index++
	}
	return -1
}

func (r *VolumeSpecEnforcer) removeCustomVolumeMounts(statefulSet *apps.StatefulSet) {

	currentCustomVolumeMounts := r.getCurrentCustomVolumeMounts(statefulSet)
	if len(currentCustomVolumeMounts) == 0 {
		return
	}

	currentCustomVolumeMountsCopy := make([]v1.VolumeMount, len(currentCustomVolumeMounts))
	copy(currentCustomVolumeMountsCopy, currentCustomVolumeMounts)

	container := &statefulSet.Spec.Template.Spec.Containers[0]

	for _, customVolumeMount := range currentCustomVolumeMountsCopy {
		index := r.getIndexOfVolumeMount(customVolumeMount, container.VolumeMounts)
		if index >= 0 {
			container.VolumeMounts = append(container.VolumeMounts[:index], container.VolumeMounts[index+1:]...)
		}
	}
}

func (r *VolumeSpecEnforcer) getIndexOfVolumeMount(volumeMountToSearch v1.VolumeMount, volumeMounts []v1.VolumeMount) int {
	index := 0
	for _, volumeMount := range volumeMounts {
		if reflect.DeepEqual(volumeMount, volumeMountToSearch) {
			return index
		}
		index++
	}
	return -1
}

func (r *VolumeSpecEnforcer) volumesToString(volumes []v1.Volume) string {

	toString := ""
	for _, volume := range volumes {
		toString += volume.String() + " - "
	}
	return toString
}

func (r *VolumeSpecEnforcer) volumeMountsToString(volumeMounts []v1.VolumeMount) string {

	toString := ""
	for _, volumeMount := range volumeMounts {
		toString += volumeMount.String() + " - "
	}
	return toString
}
