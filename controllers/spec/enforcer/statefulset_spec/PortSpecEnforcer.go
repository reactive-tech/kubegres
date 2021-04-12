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
	core "k8s.io/api/core/v1"
	"reactive-tech.io/kubegres/controllers/ctx"
	"reactive-tech.io/kubegres/controllers/states"
	"strconv"
)

type PortSpecEnforcer struct {
	kubegresContext ctx.KubegresContext
	resourcesStates states.ResourcesStates
}

func CreatePortSpecEnforcer(kubegresContext ctx.KubegresContext, resourcesStates states.ResourcesStates) PortSpecEnforcer {
	return PortSpecEnforcer{kubegresContext: kubegresContext, resourcesStates: resourcesStates}
}

func (r *PortSpecEnforcer) GetSpecName() string {
	return "Port"
}

func (r *PortSpecEnforcer) CheckForSpecDifference(statefulSet *apps.StatefulSet) StatefulSetSpecDifference {

	current := int(statefulSet.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort)
	expected := int(r.kubegresContext.Kubegres.Spec.Port)

	if current != expected {
		return StatefulSetSpecDifference{
			SpecName: r.GetSpecName(),
			Current:  strconv.Itoa(current),
			Expected: strconv.Itoa(expected),
		}
	}

	return StatefulSetSpecDifference{}
}

func (r *PortSpecEnforcer) EnforceSpec(statefulSet *apps.StatefulSet) (wasSpecUpdated bool, err error) {
	statefulSet.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = r.kubegresContext.Kubegres.Spec.Port
	return true, nil
}

func (r *PortSpecEnforcer) OnSpecEnforcedSuccessfully(statefulSet *apps.StatefulSet) error {

	if r.isPrimaryStatefulSet(statefulSet) && r.isPrimaryServicePortNotUpToDate() {
		return r.deletePrimaryService()

	} else if !r.isPrimaryStatefulSet(statefulSet) && r.isReplicaServicePortNotUpToDate() {
		return r.deleteReplicaService()
	}

	return nil
}

func (r *PortSpecEnforcer) isPrimaryStatefulSet(statefulSet *apps.StatefulSet) bool {
	return statefulSet.Name == r.resourcesStates.StatefulSets.Primary.StatefulSet.Name
}

func (r *PortSpecEnforcer) isPrimaryServicePortNotUpToDate() bool {
	if !r.resourcesStates.Services.Primary.IsDeployed {
		return false
	}

	servicePort := r.resourcesStates.Services.Primary.Service.Spec.Ports[0].Port
	postgresPort := r.kubegresContext.Kubegres.Spec.Port
	return postgresPort != servicePort
}

func (r *PortSpecEnforcer) isReplicaServicePortNotUpToDate() bool {
	if !r.resourcesStates.Services.Replica.IsDeployed {
		return false
	}

	servicePort := r.resourcesStates.Services.Replica.Service.Spec.Ports[0].Port
	postgresPort := r.kubegresContext.Kubegres.Spec.Port
	return postgresPort != servicePort
}

func (r *PortSpecEnforcer) deletePrimaryService() error {
	return r.deleteService(r.resourcesStates.Services.Primary.Service)
}

func (r *PortSpecEnforcer) deleteReplicaService() error {
	return r.deleteService(r.resourcesStates.Services.Replica.Service)
}

func (r *PortSpecEnforcer) deleteService(serviceToDelete core.Service) error {

	r.kubegresContext.Log.Info("Deleting Service because Spec port changed.", "Service name", serviceToDelete.Name)
	err := r.kubegresContext.Client.Delete(r.kubegresContext.Ctx, &serviceToDelete)

	if err != nil {
		r.kubegresContext.Log.ErrorEvent("ServiceDeletionErr", err, "Unable to delete a service to apply the new Spec port change.", "Service name", serviceToDelete.Name)
		r.kubegresContext.Log.Error(err, "Unable to delete Service", "name", serviceToDelete.Name)
		return err
	}

	r.kubegresContext.Log.InfoEvent("ServiceDeletion", "Deleted a service to apply the new Spec port change.", "Service name", serviceToDelete.Name)
	return nil
}
