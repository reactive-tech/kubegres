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
	"errors"
	"fmt"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"reactive-tech.io/kubegres/controllers/ctx"
	"reactive-tech.io/kubegres/controllers/states"
	"reactive-tech.io/kubegres/controllers/states/statefulset"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

type StorageClassSizeSpecEnforcer struct {
	kubegresContext ctx.KubegresContext
	resourcesStates states.ResourcesStates
}

func CreateStorageClassSizeSpecEnforcer(kubegresContext ctx.KubegresContext, resourcesStates states.ResourcesStates) StorageClassSizeSpecEnforcer {
	return StorageClassSizeSpecEnforcer{kubegresContext: kubegresContext, resourcesStates: resourcesStates}
}

func (r *StorageClassSizeSpecEnforcer) GetSpecName() string {
	return "StorageClassSize"
}

func (r *StorageClassSizeSpecEnforcer) CheckForSpecDifference(statefulSet *apps.StatefulSet) StatefulSetSpecDifference {

	// TODO: codes to re-enable when Kubernetes allows updating storage size in StatefulSet (see https://github.com/kubernetes/enhancements/pull/2842)
	/*
		current := statefulSet.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[core.ResourceStorage]
		expected := resource.MustParse(r.kubegresContext.Kubegres.Spec.Database.Size)

		if current != expected {
			return StatefulSetSpecDifference{
				SpecName: r.GetSpecName(),
				Current:  current.String(),
				Expected: expected.String(),
			}
		}*/

	return StatefulSetSpecDifference{}
}

func (r *StorageClassSizeSpecEnforcer) EnforceSpec(statefulSet *apps.StatefulSet) (wasSpecUpdated bool, err error) {

	// The reason why we have to update PVC rather than updating StatefulSet is because
	// StatefulSet does not allow updating the following field:
	// statefulSet.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[core.ResourceStorage] = resource.MustParse(newSize)
	// If we update the field above we get the following error:
	// "Forbidden: updates to Statefulset spec for fields other than 'replicas', 'template', and 'updateStrategy' are forbidden"

	persistentVolumeClaimName, err := r.getPersistentVolumeClaimName(statefulSet)
	if err != nil {
		r.kubegresContext.Log.Error(err, "Unable to find StatefulSet's Persistence Volume Claim name", "StatefulSet name", statefulSet.Name)
		return false, err
	}

	persistentVolumeClaim, err := r.getPersistentVolumeClaim(persistentVolumeClaimName)
	if err != nil {
		r.kubegresContext.Log.Error(err, "Unable to find Persistence Volume Claim with the given name.", "PVC name", persistentVolumeClaimName)
		return false, err
	}

	newSize := r.kubegresContext.Kubegres.Spec.Database.Size
	r.kubegresContext.Log.Info("Updating Persistence Volume Claim to new size", "PVC name", persistentVolumeClaimName, "New size", newSize)
	persistentVolumeClaim.Spec.Resources.Requests = core.ResourceList{core.ResourceStorage: resource.MustParse(newSize)}

	err = r.kubegresContext.Client.Update(r.kubegresContext.Ctx, persistentVolumeClaim)
	if err != nil {
		r.kubegresContext.Log.Error(err, "Unable to update Persistence Volume Claim to new size", "PVC name", persistentVolumeClaimName, "New size", newSize)
		return false, err
	}

	r.kubegresContext.Log.Info("Updated Persistence Volume Claim Spec to new size", "PVC name", persistentVolumeClaimName, "New size", newSize)

	r.updateStatefulSetToForceToRestart(statefulSet)

	return true, nil
}

func (r *StorageClassSizeSpecEnforcer) OnSpecEnforcedSuccessfully(statefulSet *apps.StatefulSet) error {
	return nil
}

func (r *StorageClassSizeSpecEnforcer) getPersistentVolumeClaimName(statefulSet *apps.StatefulSet) (string, error) {

	statefulSetWrapper, err := r.getStatefulSetWrapper(statefulSet)
	if err != nil {
		return "", err
	}

	podVolumes := statefulSetWrapper.Pod.Pod.Spec.Volumes
	if len(podVolumes) == 0 {
		return "", errors.New("The pod does not have a volume. Pod name: " + statefulSetWrapper.Pod.Pod.Name)
	}

	return podVolumes[0].PersistentVolumeClaim.ClaimName, nil
}

func (r *StorageClassSizeSpecEnforcer) getPersistentVolumeClaim(persistentVolumeClaimName string) (*core.PersistentVolumeClaim, error) {

	namespace := r.kubegresContext.Kubegres.Namespace
	persistentVolumeClaimKey := client.ObjectKey{Namespace: namespace, Name: persistentVolumeClaimName}
	persistentVolumeClaim := &core.PersistentVolumeClaim{}
	err := r.kubegresContext.Client.Get(r.kubegresContext.Ctx, persistentVolumeClaimKey, persistentVolumeClaim)

	if err != nil {
		return &core.PersistentVolumeClaim{}, err
	}

	return persistentVolumeClaim, nil
}

func (r *StorageClassSizeSpecEnforcer) getStatefulSetWrapper(statefulSet *apps.StatefulSet) (statefulset.StatefulSetWrapper, error) {

	statefulSetWrapper, err := r.resourcesStates.StatefulSets.All.GetByName(statefulSet.Name)
	if err == nil {
		return statefulSetWrapper, nil
	}

	return statefulset.StatefulSetWrapper{}, errors.New("Cannot find statefulSet inside ResourcesStates. StatefulSet name: " + statefulSet.Name)
}

func (r *StorageClassSizeSpecEnforcer) updateStatefulSetToForceToRestart(statefulSet *apps.StatefulSet) {
	index := 0
	sizeChangedCounter := statefulSet.Spec.Template.ObjectMeta.Labels["sizeChangedCounter"]

	if sizeChangedCounter != "" {
		index, _ = strconv.Atoi(sizeChangedCounter)
	}

	statefulSet.Spec.Template.ObjectMeta.Labels["sizeChangedCounter"] = fmt.Sprint(index + 1)
}
