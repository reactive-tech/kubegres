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

package template

import (
	"k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"reactive-tech.io/kubegres/controllers/ctx"
	"reactive-tech.io/kubegres/controllers/states"
)

const (
	defaultMode int32 = 0777
)

type CustomConfigSpecHelper struct {
	kubegresContext ctx.KubegresContext
	resourcesStates states.ResourcesStates
}

func CreateCustomConfigSpecHelper(kubegresContext ctx.KubegresContext, resourcesStates states.ResourcesStates) CustomConfigSpecHelper {
	return CustomConfigSpecHelper{kubegresContext: kubegresContext, resourcesStates: resourcesStates}
}

func (r *CustomConfigSpecHelper) ConfigureStatefulSet(statefulSet *v1.StatefulSet) (hasStatefulSetChanged bool, differenceDetails string) {

	configMap := r.resourcesStates.Config

	if r.updateVolumeMountNameIfChanged(configMap.ConfigLocations.PostgreConf, states.ConfigMapDataKeyPostgresConf, statefulSet) {
		differenceDetails += r.createDescriptionMsg(configMap.ConfigLocations.PostgreConf, states.ConfigMapDataKeyPostgresConf)
		hasStatefulSetChanged = true
	}

	if r.updateVolumeMountNameIfChanged(configMap.ConfigLocations.PrimaryInitScript, states.ConfigMapDataKeyPrimaryInitScript, statefulSet) {
		differenceDetails += r.createDescriptionMsg(configMap.ConfigLocations.PrimaryInitScript, states.ConfigMapDataKeyPrimaryInitScript)
		hasStatefulSetChanged = true
	}

	if r.updateVolumeMountNameIfChanged(configMap.ConfigLocations.PgHbaConf, states.ConfigMapDataKeyPgHbaConf, statefulSet) {
		differenceDetails += r.createDescriptionMsg(configMap.ConfigLocations.PgHbaConf, states.ConfigMapDataKeyPgHbaConf)
		hasStatefulSetChanged = true
	}

	statefulSetTemplateSpec := &statefulSet.Spec.Template.Spec

	customConfigMapVolume := r.getCustomConfigMapVolume(statefulSetTemplateSpec.Volumes)

	if configMap.IsCustomConfigDeployed {

		if customConfigMapVolume == nil ||
			customConfigMapVolume.ConfigMap.Name != r.getSpecCustomConfig() {

			if customConfigMapVolume != nil &&
				customConfigMapVolume.ConfigMap.Name != r.getSpecCustomConfig() {
				r.deleteCustomConfigMapVolumeIfExist(statefulSetTemplateSpec)
			}

			r.addNewConfigMapVolumeWithSpecValue(statefulSetTemplateSpec)
			hasStatefulSetChanged = true
			differenceDetails += r.getSpecCustomConfig()
		}

	} else if customConfigMapVolume != nil {
		r.deleteCustomConfigMapVolumeIfExist(statefulSetTemplateSpec)
		hasStatefulSetChanged = true
		differenceDetails += "Deleted from StatefulSet Spec, the volume configuration for customConfig as it is not used anymore"
	}

	return hasStatefulSetChanged, differenceDetails
}

func (r *CustomConfigSpecHelper) updateVolumeMountNameIfChanged(volumeMountName, configMapDataKey string, statefulSet *v1.StatefulSet) (updated bool) {

	volumeMounts := statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts

	for i := 0; i < len(volumeMounts); i++ {
		volumeMount := volumeMounts[i]
		if volumeMount.SubPath == configMapDataKey && volumeMount.Name != volumeMountName {
			statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts[i].Name = volumeMountName
			updated = true
		}
	}
	return updated
}

func (r *CustomConfigSpecHelper) createDescriptionMsg(volumeMountName, configMapDataKey string) string {
	return "VolumeMount with subPath: '" + configMapDataKey + "' was updated to name: '" + volumeMountName + "' - "
}

func (r *CustomConfigSpecHelper) getVolumeMountIndex(configMapDataKey string, statefulSet *v1.StatefulSet) int {
	volumeMounts := statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts
	for i := 0; i < len(volumeMounts); i++ {
		if volumeMounts[i].SubPath == configMapDataKey {
			return i
		}
	}
	return -1
}

func (r *CustomConfigSpecHelper) getCustomConfigMapVolume(volumes []core.Volume) *core.Volume {
	for _, volume := range volumes {
		if volume.Name == states.CustomConfigType {
			return &volume
		}
	}
	return nil
}

func (r *CustomConfigSpecHelper) isCustomConfigMapNameDifferentThanKubegresSpec(existingCustomConfigMapVolume *core.Volume) bool {
	return existingCustomConfigMapVolume != nil &&
		existingCustomConfigMapVolume.ConfigMap.Name != r.getSpecCustomConfig()
}

func (r *CustomConfigSpecHelper) updateCustomConfigMapNameWithKubegresSpec(existingCustomConfigMapVolume *core.Volume) {
	existingCustomConfigMapVolume.ConfigMap.Name = r.getSpecCustomConfig()
}

func (r *CustomConfigSpecHelper) addNewConfigMapVolumeWithSpecValue(statefulSetTemplateSpec *core.PodSpec) {
	statefulSetTemplateSpec.Volumes = append(statefulSetTemplateSpec.Volumes, r.createConfigMapVolume())
}

func (r *CustomConfigSpecHelper) deleteCustomConfigMapVolumeIfExist(statefulSetTemplateSpec *core.PodSpec) {

	newVolumes := make([]core.Volume, 0)

	for _, volume := range statefulSetTemplateSpec.Volumes {
		if volume.Name != states.CustomConfigType {
			newVolumes = append(newVolumes, volume)
		}
	}

	statefulSetTemplateSpec.Volumes = newVolumes
}

func (r *CustomConfigSpecHelper) createConfigMapVolume() core.Volume {
	defMode := defaultMode
	return core.Volume{
		Name: states.CustomConfigType,
		VolumeSource: core.VolumeSource{
			ConfigMap: &core.ConfigMapVolumeSource{
				DefaultMode: &defMode,
				LocalObjectReference: core.LocalObjectReference{
					Name: r.getSpecCustomConfig(),
				},
			},
		},
	}
}

func (r *CustomConfigSpecHelper) doesCustomConfigExist() bool {
	return r.getSpecCustomConfig() != "" &&
		r.getSpecCustomConfig() != ctx.BaseConfigMapName
}

func (r *CustomConfigSpecHelper) getSpecCustomConfig() string {
	return r.kubegresContext.Kubegres.Spec.CustomConfig
}
