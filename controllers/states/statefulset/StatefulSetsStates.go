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
	"errors"
	apps "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"reactive-tech.io/kubegres/controllers/ctx"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

type StatefulSetsStates struct {
	NbreDeployed     int32
	Primary          StatefulSetWrapper
	Replicas         Replicas
	All              StatefulSetWrappers
	SpecNbreToDeploy int32
	kubegresContext  ctx.KubegresContext
}

type Replicas struct {
	All              StatefulSetWrappers
	NbreDeployed     int32
	SpecNbreToDeploy int32
	NbreToDeploy     int32 // zero => no more to deploy; negative => we should deploy less; positive => we should deploy more
}

func LoadStatefulSetsStates(kubegresContext ctx.KubegresContext) (StatefulSetsStates, error) {
	statefulSetsStates := StatefulSetsStates{kubegresContext: kubegresContext}
	err := statefulSetsStates.loadStates()
	return statefulSetsStates, err
}

func (r *StatefulSetsStates) ShouldMoreReplicaBeDeployed() bool {
	return r.Replicas.NbreToDeploy > 0
}

func (r *StatefulSetsStates) ShouldLessReplicaBeDeployed() bool {
	return r.Replicas.NbreToDeploy < 0
}

func (r *StatefulSetsStates) GetNbreReplicaToDeploy() int32 {
	if r.ShouldMoreReplicaBeDeployed() {
		return r.Replicas.NbreToDeploy
	}
	return 0
}

func (r *StatefulSetsStates) GetNbreReplicaToUndeploy() int32 {
	if r.ShouldLessReplicaBeDeployed() {
		return r.Replicas.NbreToDeploy * -1
	}
	return 0
}

func (r *StatefulSetsStates) loadStates() (err error) {

	deployedStatefulSets, err := r.getDeployedStatefulSets()
	if err != nil {
		return err
	}

	r.NbreDeployed = int32(len(deployedStatefulSets.Items))
	r.SpecNbreToDeploy = *r.kubegresContext.Kubegres.Spec.Replicas

	var podsStates PodStates
	if r.NbreDeployed > 0 {
		podsStates, err = loadPodsStates(r.kubegresContext)
		if err != nil {
			return err
		}
	}

	for _, statefulSet := range deployedStatefulSets.Items {
		err := r.createAndAppendStatefulSetStates(statefulSet, podsStates)
		if err != nil {
			return err
		}
	}

	r.calculateNbreReplicasToDeploy()

	return nil
}

func (r *StatefulSetsStates) createAndAppendStatefulSetStates(statefulSet apps.StatefulSet, podsStates PodStates) error {

	statefulSetWrapper := StatefulSetWrapper{}
	statefulSetWrapper.IsDeployed = true
	statefulSetWrapper.IsReady = statefulSet.Status.ReadyReplicas > 0
	statefulSetWrapper.StatefulSet = statefulSet

	instanceIndex, err := r.getInstanceIndexFromSpec(statefulSet)
	if err != nil {
		r.kubegresContext.Log.Error(err, "Unable to get instance index")
		return err
	}

	statefulSetWrapper.InstanceIndex = instanceIndex
	statefulSetWrapper.Pod = r.getPodByInstanceIndex(instanceIndex, podsStates)
	r.All.Add(statefulSetWrapper)

	if r.isPrimary(statefulSet) {
		err = r.setPrimaryStatefulSetStates(statefulSet, statefulSetWrapper)
		if err != nil {
			return err
		}

	} else {
		r.addReplicaStatefulSetStates(statefulSetWrapper)
	}

	return nil
}

func (r *StatefulSetsStates) setPrimaryStatefulSetStates(statefulSet apps.StatefulSet, statefulSetWrapper StatefulSetWrapper) error {

	// If a statefulSet was already deployed as primary, we cannot have a second one as primary
	if r.Primary.IsDeployed {
		errMsg := "Identified 2 instances of statefulSet with Names: '" + r.Primary.StatefulSet.Name + "' and '" + statefulSet.Name +
			"' which have label 'replicationRole' set to 'primary'. Only one instance should be primary for Kubegres resource '" + r.kubegresContext.Kubegres.Name + "'."
		err := errors.New(errMsg)
		r.kubegresContext.Log.ErrorEvent("StatefulSetLoadingErr", err, errMsg)
		return err
	}

	r.Primary = statefulSetWrapper
	return nil
}

func (r *StatefulSetsStates) addReplicaStatefulSetStates(statefulSetWrapper StatefulSetWrapper) {
	r.Replicas.NbreDeployed++
	r.Replicas.All.Add(statefulSetWrapper)
}

func (r *StatefulSetsStates) getPodByInstanceIndex(instanceIndex int32, podsStates PodStates) PodWrapper {
	for _, pod := range podsStates.pods {
		if pod.InstanceIndex == instanceIndex {
			return pod
		}
	}
	return PodWrapper{}
}

func (r *StatefulSetsStates) isPrimary(statefulSet apps.StatefulSet) bool {
	return statefulSet.Spec.Template.Labels["replicationRole"] == ctx.PrimaryRoleName
}

func (r *StatefulSetsStates) getDeployedStatefulSets() (*apps.StatefulSetList, error) {

	list := &apps.StatefulSetList{}
	opts := []client.ListOption{
		client.InNamespace(r.kubegresContext.Kubegres.Namespace),
		client.MatchingFields{ctx.DeploymentOwnerKey: r.kubegresContext.Kubegres.Name},
	}
	err := r.kubegresContext.Client.List(r.kubegresContext.Ctx, list, opts...)

	if err != nil {
		if apierrors.IsNotFound(err) {
			err = nil
		} else {
			r.kubegresContext.Log.ErrorEvent("StatefulSetLoadingErr", err, "Unable to load any deployed StatefulSets.", "Kubegres name", r.kubegresContext.Kubegres.Name)
		}
	}

	return list, err
}

func (r *StatefulSetsStates) getInstanceIndexFromSpec(statefulSet apps.StatefulSet) (int32, error) {
	instanceIndexStr := statefulSet.Spec.Template.Labels["index"]
	instanceIndex, err := strconv.ParseInt(instanceIndexStr, 10, 32)
	if err != nil {
		r.kubegresContext.Log.ErrorEvent("StatefulSetLoadingErr", err, "Unable to convert StatefulSet's label 'index' with value: "+instanceIndexStr+" into an integer. The name of statefulSet with this label is "+statefulSet.Name+".")
		return 0, err
	}
	return int32(instanceIndex), nil
}

func (r *StatefulSetsStates) calculateNbreReplicasToDeploy() {
	r.Replicas.SpecNbreToDeploy = r.getSpecNbreReplicasToDeploy()
	r.Replicas.NbreToDeploy = r.Replicas.SpecNbreToDeploy - r.Replicas.NbreDeployed
}

func (r *StatefulSetsStates) getSpecNbreReplicasToDeploy() int32 {
	if r.SpecNbreToDeploy <= 1 {
		return 0
	}
	return r.SpecNbreToDeploy - 1
}
