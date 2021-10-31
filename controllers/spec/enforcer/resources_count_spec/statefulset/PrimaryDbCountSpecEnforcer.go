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
	v1 "k8s.io/api/core/v1"
	postgresV1 "reactive-tech.io/kubegres/api/v1"
	"reactive-tech.io/kubegres/controllers/ctx"
	"reactive-tech.io/kubegres/controllers/operation"
	"reactive-tech.io/kubegres/controllers/spec/enforcer/resources_count_spec/statefulset/failover"
	"reactive-tech.io/kubegres/controllers/spec/template"
	"reactive-tech.io/kubegres/controllers/states"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

type PrimaryDbCountSpecEnforcer struct {
	kubegresContext          ctx.KubegresContext
	resourcesStates          states.ResourcesStates
	resourcesCreator         template.ResourcesCreatorFromTemplate
	primaryToReplicaFailOver failover.PrimaryToReplicaFailOver
	blockingOperation        *operation.BlockingOperation
}

func CreatePrimaryDbCountSpecEnforcer(
	kubegresContext ctx.KubegresContext,
	resourcesStates states.ResourcesStates,
	resourcesCreator template.ResourcesCreatorFromTemplate,
	blockingOperation *operation.BlockingOperation,
	primaryToReplicaFailOver failover.PrimaryToReplicaFailOver) PrimaryDbCountSpecEnforcer {

	return PrimaryDbCountSpecEnforcer{
		kubegresContext:          kubegresContext,
		resourcesStates:          resourcesStates,
		resourcesCreator:         resourcesCreator,
		blockingOperation:        blockingOperation,
		primaryToReplicaFailOver: primaryToReplicaFailOver,
	}
}

func (r *PrimaryDbCountSpecEnforcer) CreateOperationConfigForPrimaryDbDeploying() operation.BlockingOperationConfig {

	return operation.BlockingOperationConfig{
		OperationId:       operation.OperationIdPrimaryDbCountSpecEnforcement,
		StepId:            operation.OperationStepIdPrimaryDbDeploying,
		TimeOutInSeconds:  300,
		CompletionChecker: func(operation postgresV1.KubegresBlockingOperation) bool { return r.isPrimaryDbReady() },
	}
}

func (r *PrimaryDbCountSpecEnforcer) Enforce() error {

	// Backward compatibility logic where we initialize the field 'EnforcedReplicas'
	// added in Kubegres' status from version 1.8
	r.initialiseStatusEnforcedReplicas()

	if r.blockingOperation.IsActiveOperationIdDifferentOf(operation.OperationIdPrimaryDbCountSpecEnforcement) {
		return nil
	}

	if r.isPrimaryDbReady() && r.hasLastPrimaryCountSpecEnforcementAttemptTimedOut() {
		r.blockingOperation.RemoveActiveOperation()
		r.logKubegresFeaturesAreReEnabled()
	}

	if r.primaryToReplicaFailOver.ShouldWeFailOver() {
		return r.primaryToReplicaFailOver.FailOver()

	} else if r.shouldWeDeployNewPrimaryDb() {
		return r.deployNewPrimaryStatefulSet()
	}

	return nil
}

func (r *PrimaryDbCountSpecEnforcer) initialiseStatusEnforcedReplicas() {
	if r.kubegresContext.Kubegres.Status.EnforcedReplicas > 0 {
		return
	}

	specReplicas := *r.kubegresContext.Kubegres.Spec.Replicas

	if specReplicas >= 1 && specReplicas == r.resourcesStates.StatefulSets.NbreDeployed {
		r.kubegresContext.Status.SetEnforcedReplicas(specReplicas)
	}
}

func (r *PrimaryDbCountSpecEnforcer) hasLastPrimaryCountSpecEnforcementAttemptTimedOut() bool {
	return r.blockingOperation.HasActiveOperationIdTimedOut(operation.OperationIdPrimaryDbCountSpecEnforcement)
}

func (r *PrimaryDbCountSpecEnforcer) isPrimaryDbReady() bool {
	return r.resourcesStates.StatefulSets.Primary.IsReady
}

func (r *PrimaryDbCountSpecEnforcer) logKubegresFeaturesAreReEnabled() {
	r.kubegresContext.Log.InfoEvent("KubegresReEnabled", "PrimaryDB is set to ready again. "+
		"We can safely re-enable all features of Kubegres.")
}

func (r *PrimaryDbCountSpecEnforcer) shouldWeDeployNewPrimaryDb() bool {

	shouldWeDeployNewPrimary := !r.resourcesStates.StatefulSets.Primary.IsDeployed &&
		r.resourcesStates.StatefulSets.Replicas.NbreDeployed == 0

	if shouldWeDeployNewPrimary {
		if *r.kubegresContext.Kubegres.Spec.Replicas == 1 || !r.hasPrimaryEverBeenDeployed() {
			return true
		}
	}

	return false
}

func (r *PrimaryDbCountSpecEnforcer) hasPrimaryEverBeenDeployed() bool {
	return r.kubegresContext.Kubegres.Status.EnforcedReplicas > 0
}

func (r *PrimaryDbCountSpecEnforcer) deployNewPrimaryStatefulSet() error {

	if r.hasLastPrimaryCountSpecEnforcementAttemptTimedOut() {
		r.logDeploymentTimedOut()
		return nil
	}

	instanceIndex := r.getInstanceIndexIfPrimaryNeedsToBeRecreatedAndThereIsNoReplicaSetUp()

	if instanceIndex == 0 {
		instanceIndex = r.kubegresContext.Status.GetLastCreatedInstanceIndex() + 1
	}

	if err := r.activateBlockingOperation(instanceIndex); err != nil {
		r.kubegresContext.Log.ErrorEvent("PrimaryStatefulSetOperationActivationErr", err, "Error while activating blocking operation for the deployment of a Primary StatefulSet.", "InstanceIndex", instanceIndex)
		return err
	}

	primaryStatefulSet, err := r.resourcesCreator.CreatePrimaryStatefulSet(instanceIndex)
	if err != nil {
		r.kubegresContext.Log.ErrorEvent("PrimaryStatefulSetTemplateErr", err, "Error while creating a Primary StatefulSet object from template.", "InstanceIndex", instanceIndex)
		r.blockingOperation.RemoveActiveOperation()
		return err
	}

	if err = r.kubegresContext.Client.Create(r.kubegresContext.Ctx, &primaryStatefulSet); err != nil {
		r.kubegresContext.Log.ErrorEvent("PrimaryStatefulSetDeploymentErr", err, "Unable to deploy Primary StatefulSet.", "Primary name", primaryStatefulSet.Name)
		r.blockingOperation.RemoveActiveOperation()
		return err
	}

	r.kubegresContext.Status.SetEnforcedReplicas(r.kubegresContext.Kubegres.Status.EnforcedReplicas + 1)

	if r.kubegresContext.Status.GetLastCreatedInstanceIndex() == 0 {
		r.kubegresContext.Status.SetLastCreatedInstanceIndex(1)
	}

	r.kubegresContext.Log.InfoEvent("PrimaryStatefulSetDeployment", "Deployed Primary StatefulSet.", "Primary name", primaryStatefulSet.Name)
	return nil
}

func (r *PrimaryDbCountSpecEnforcer) logDeploymentTimedOut() {

	activeOperation := r.blockingOperation.GetActiveOperation()
	operationTimeOutStr := strconv.FormatInt(r.CreateOperationConfigForPrimaryDbDeploying().TimeOutInSeconds, 10)
	primaryStatefulSetName := activeOperation.StatefulSetOperation.Name

	err := errors.New("Primary DB StatefulSet deployment timed-out")
	r.kubegresContext.Log.ErrorEvent("PrimaryStatefulSetDeploymentTimedOutErr", err,
		"Last attempt to deploy a Primary DB StatefulSet has timed-out after "+operationTimeOutStr+" seconds. "+
			"The new Primary DB is still NOT ready. It must be fixed manually. "+
			"Until the PrimaryDB is ready, most of the features of Kubegres are disabled for safety reason. ",
		"Primary DB StatefulSet to fix", primaryStatefulSetName)
}

func (r *PrimaryDbCountSpecEnforcer) getInstanceIndexIfPrimaryNeedsToBeRecreatedAndThereIsNoReplicaSetUp() int32 {
	if r.isThereReplicaToFailOverToInSpec() || r.wasPrimaryNeverDeployed() {
		return 0
	}

	primaryPvc := r.getLastDeployedPrimaryPvc()

	if primaryPvc.Name != "" {
		r.kubegresContext.Log.InfoEvent("RecreatingPrimaryStatefulSet", "We will recreate a new Primary StatefulSet as it is not available. We will reuse existing PVC: '"+primaryPvc.Name+"'")
		return r.kubegresContext.Status.GetLastCreatedInstanceIndex()
	}

	r.kubegresContext.Log.InfoEvent("RecreatingPrimaryStatefulSet", "We will recreate a new Primary StatefulSet as it is not available. We will create a new PVC as the existing one is not available '"+primaryPvc.Name+"'")
	return 0
}

func (r *PrimaryDbCountSpecEnforcer) isThereReplicaToFailOverToInSpec() bool {
	return *r.kubegresContext.Kubegres.Spec.Replicas > 1
}

func (r *PrimaryDbCountSpecEnforcer) wasPrimaryNeverDeployed() bool {
	return r.kubegresContext.Status.GetLastCreatedInstanceIndex() == 0
}

func (r *PrimaryDbCountSpecEnforcer) activateBlockingOperation(statefulSetInstanceIndex int32) error {
	return r.blockingOperation.ActivateOperationOnStatefulSet(operation.OperationIdPrimaryDbCountSpecEnforcement,
		operation.OperationStepIdPrimaryDbDeploying,
		statefulSetInstanceIndex)
}

func (r *PrimaryDbCountSpecEnforcer) getLastDeployedPrimaryPvc() *v1.PersistentVolumeClaim {

	lastCreatedInstanceIndex := r.kubegresContext.Status.GetLastCreatedInstanceIndex()
	kubegresName := r.kubegresContext.Kubegres.Name
	resourceName := "postgres-db-" + kubegresName + "-" + strconv.Itoa(int(lastCreatedInstanceIndex)) + "-0"

	namespace := r.kubegresContext.Kubegres.Namespace
	resourceKey := client.ObjectKey{Namespace: namespace, Name: resourceName}
	pvc := &v1.PersistentVolumeClaim{}
	_ = r.kubegresContext.Client.Get(r.kubegresContext.Ctx, resourceKey, pvc)
	return pvc
}
