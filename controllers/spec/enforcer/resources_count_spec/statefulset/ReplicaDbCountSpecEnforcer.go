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
	v1 "k8s.io/api/apps/v1"
	postgresV1 "reactive-tech.io/kubegres/api/v1"
	"reactive-tech.io/kubegres/controllers/ctx"
	"reactive-tech.io/kubegres/controllers/operation"
	"reactive-tech.io/kubegres/controllers/spec/template"
	"reactive-tech.io/kubegres/controllers/states"
	"reactive-tech.io/kubegres/controllers/states/statefulset"
	"strconv"
)

type ReplicaDbCountSpecEnforcer struct {
	kubegresContext   ctx.KubegresContext
	resourcesStates   states.ResourcesStates
	resourcesCreator  template.ResourcesCreatorFromTemplate
	blockingOperation *operation.BlockingOperation
}

func CreateReplicaDbCountSpecEnforcer(
	kubegresContext ctx.KubegresContext,
	resourcesStates states.ResourcesStates,
	resourcesCreator template.ResourcesCreatorFromTemplate,
	blockingOperation *operation.BlockingOperation) ReplicaDbCountSpecEnforcer {

	return ReplicaDbCountSpecEnforcer{
		kubegresContext:   kubegresContext,
		resourcesStates:   resourcesStates,
		resourcesCreator:  resourcesCreator,
		blockingOperation: blockingOperation,
	}
}

func (r *ReplicaDbCountSpecEnforcer) CreateOperationConfigForReplicaDbDeploying() operation.BlockingOperationConfig {

	return operation.BlockingOperationConfig{
		OperationId:       operation.OperationIdReplicaDbCountSpecEnforcement,
		StepId:            operation.OperationStepIdReplicaDbDeploying,
		TimeOutInSeconds:  300,
		CompletionChecker: r.isReplicaDbReady,
	}
}

func (r *ReplicaDbCountSpecEnforcer) CreateOperationConfigForReplicaDbUndeploying() operation.BlockingOperationConfig {

	return operation.BlockingOperationConfig{
		OperationId:       operation.OperationIdReplicaDbCountSpecEnforcement,
		StepId:            operation.OperationStepIdReplicaDbUndeploying,
		TimeOutInSeconds:  60,
		CompletionChecker: r.isReplicaDbUndeployed,
	}
}

func (r *ReplicaDbCountSpecEnforcer) Enforce() error {
	if r.blockingOperation.IsActiveOperationIdDifferentOf(operation.OperationIdReplicaDbCountSpecEnforcement) {
		return nil
	}

	if r.hasLastAttemptTimedOut() {
		if r.isPreviouslyFailedAttemptOnReplicaDbFixed() {
			r.blockingOperation.RemoveActiveOperation()
			r.logKubegresFeaturesAreReEnabled()

		} else {
			r.logTimedOut()
			return nil
		}
	}

	if !r.isPrimaryDbReady() {
		return nil
	}

	isManualFailoverRequested := r.isManualFailoverRequested()
	if isManualFailoverRequested {
		r.resetInSpecManualFailover()
	}

	if r.isReplicaOperationInProgress() {
		return nil
	}

	// Check if the nodeSets deployed are different from the spec
	nextInstanceToDeploy := r.getNextInstanceToDeploy()
	if nextInstanceToDeploy != "" {
		if r.isAutomaticFailoverDisabled() && !isManualFailoverRequested &&
			!r.doesSpecRequireTheDeploymentOfAdditionalReplicas() {
			r.logAutomaticFailoverIsDisabled()
			return nil
		}

		return r.deployReplicaStatefulSet(nextInstanceToDeploy)
	}

	nextInstanceToUndeploy := r.getNextInstanceToUndeploy()
	if nextInstanceToUndeploy != "" {
		replicaToUndeploy := r.getReplicaToUndeploy(nextInstanceToUndeploy)
		return r.undeployReplicaStatefulSets(replicaToUndeploy)
	}

	for _, replicaStatefulSet := range r.getDeployedReplicas() {
		if !replicaStatefulSet.IsReady {
			return r.undeployReplicaStatefulSets(replicaStatefulSet)
		}
	}

	return nil
}

func (r *ReplicaDbCountSpecEnforcer) isReplicaOperationInProgress() bool {
	return r.blockingOperation.GetActiveOperation().OperationId == operation.OperationIdReplicaDbCountSpecEnforcement
}

func (r *ReplicaDbCountSpecEnforcer) getNextInstanceToDeploy() string {
	existingStatefulSets := r.resourcesStates.StatefulSets.Replicas.All.GetAllSortedByInstance()
	for _, nodeSet := range r.kubegresContext.GetNodeSetsFromSpec() {
		if nodeSet.Name == r.resourcesStates.StatefulSets.Primary.Instance() {
			continue
		}

		found := false
		for _, instance := range existingStatefulSets {
			if nodeSet.Name == instance.Instance() {
				found = true
				break
			}
		}
		if !found {
			return nodeSet.Name
		}
	}
	return ""
}

func (r *ReplicaDbCountSpecEnforcer) getNextInstanceToUndeploy() string {
	expectedNodeSets := r.kubegresContext.GetNodeSetsFromSpec()
	for _, statefulSet := range r.resourcesStates.StatefulSets.Replicas.All.GetAllSortedByInstance() {
		found := false
		for _, nodeSet := range expectedNodeSets {
			if nodeSet.Name == statefulSet.Instance() {
				found = true
				break
			}
		}
		if !found {
			return statefulSet.Instance()
		}
	}
	return ""
}

func (r *ReplicaDbCountSpecEnforcer) getDeployedReplicas() []statefulset.StatefulSetWrapper {
	return r.resourcesStates.StatefulSets.Replicas.All.GetAllSortedByInstance()
}

func (r *ReplicaDbCountSpecEnforcer) getNumberDeployedReplicas() int32 {
	return r.resourcesStates.StatefulSets.Replicas.NumberDeployed
}

func (r *ReplicaDbCountSpecEnforcer) getExpectedNumberReplicasToDeploy() int32 {
	expectedNumberToDeploy := r.resourcesStates.StatefulSets.SpecExpectedNumberToDeploy

	if expectedNumberToDeploy <= 1 {
		return 0
	}
	return expectedNumberToDeploy - 1
}

func (r *ReplicaDbCountSpecEnforcer) hasLastAttemptTimedOut() bool {
	return r.blockingOperation.HasActiveOperationIdTimedOut(operation.OperationIdReplicaDbCountSpecEnforcement)
}

func (r *ReplicaDbCountSpecEnforcer) isPreviouslyFailedAttemptOnReplicaDbFixed() bool {
	activeOperation := r.blockingOperation.GetActiveOperation()
	instance := activeOperation.StatefulSetOperation.Instance
	replica, err := r.resourcesStates.StatefulSets.Replicas.All.GetByInstance(instance)

	return err != nil || replica.IsReady
}

func (r *ReplicaDbCountSpecEnforcer) logKubegresFeaturesAreReEnabled() {
	r.kubegresContext.Log.InfoEvent("KubegresReEnabled", "Replica DB which caused operation to time-out "+
		"is either set to ready again or it was removed. We can safely re-enable all features of Kubegres.")
}

func (r *ReplicaDbCountSpecEnforcer) logTimedOut() {

	activeOperation := r.blockingOperation.GetActiveOperation()
	operationTimeOutStr := strconv.FormatInt(r.CreateOperationConfigForReplicaDbDeploying().TimeOutInSeconds, 10)
	replicaStatefulSetName := activeOperation.StatefulSetOperation.Name

	if activeOperation.StepId == operation.OperationStepIdReplicaDbDeploying {

		err := errors.New("replica DB StatefulSet deployment timed-out")
		r.kubegresContext.Log.ErrorEvent("ReplicaStatefulSetDeploymentTimedOutErr", err,
			"Last deployment attempt of a Replica DB StatefulSet has timed-out after "+operationTimeOutStr+" seconds. "+
				"The new Replica DB is still NOT ready. It must be fixed manually. "+
				"Until the ReplicaDB is ready, most of the features of Kubegres are disabled for safety reason. ",
			"Replica DB StatefulSet to fix", replicaStatefulSetName)

	} else {
		err := errors.New("replica DB StatefulSet un-deployment timed-out")
		r.kubegresContext.Log.ErrorEvent("ReplicaStatefulSetDeploymentTimedOutErr", err,
			"Last un-deployment attempt of a Replica DB StatefulSet has timed-out after "+operationTimeOutStr+" seconds. "+
				"The new Replica DB is still NOT removed. It must be removed manually. "+
				"Until the ReplicaDB is removed, most of the features of Kubegres are disabled for safety reason. ",
			"Replica DB StatefulSet to remove", replicaStatefulSetName)
	}
}

func (r *ReplicaDbCountSpecEnforcer) isAutomaticFailoverDisabled() bool {
	return r.kubegresContext.Kubegres.Spec.Failover.IsDisabled
}

func (r *ReplicaDbCountSpecEnforcer) isManualFailoverRequested() bool {
	return r.kubegresContext.Kubegres.Spec.Failover.PromotePod != ""
}

func (r *ReplicaDbCountSpecEnforcer) doesSpecRequireTheDeploymentOfAdditionalReplicas() bool {
	return r.kubegresContext.ReplicasCount() > r.kubegresContext.Kubegres.Status.EnforcedReplicas
}

func (r *ReplicaDbCountSpecEnforcer) resetInSpecManualFailover() error {
	r.kubegresContext.Log.Info("Resetting the field 'failover.promotePod' in spec.")
	r.kubegresContext.Kubegres.Spec.Failover.PromotePod = ""
	return r.kubegresContext.Client.Update(r.kubegresContext.Ctx, r.kubegresContext.Kubegres)
}

func (r *ReplicaDbCountSpecEnforcer) isPrimaryDbReady() bool {
	return r.resourcesStates.StatefulSets.Primary.IsReady
}

func (r *ReplicaDbCountSpecEnforcer) isReplicaDbReady(operation postgresV1.KubegresBlockingOperation) bool {
	instance := operation.StatefulSetOperation.Instance
	statefulSetWrapper, err := r.resourcesStates.StatefulSets.Replicas.All.GetByInstance(instance)
	if err != nil {
		r.kubegresContext.Log.InfoEvent("A replica StatefulSet's instance does not exist. As a result "+
			"we will return false inside a blocking operation completion checker 'isReplicaDbReady()'",
			"Instance", instance)
		return false
	}

	return statefulSetWrapper.IsReady
}

func (r *ReplicaDbCountSpecEnforcer) isReplicaDbUndeployed(operation postgresV1.KubegresBlockingOperation) bool {
	instance := operation.StatefulSetOperation.Instance
	_, err := r.resourcesStates.StatefulSets.Replicas.All.GetByInstance(instance)
	return err != nil
}

func (r *ReplicaDbCountSpecEnforcer) deployReplicaStatefulSet(instance string) error {
	err := r.activateBlockingOperationForDeployment(instance)
	if err != nil {
		r.kubegresContext.Log.ErrorEvent("ReplicaStatefulSetOperationActivationErr", err, "Error while activating blocking operation for the deployment of a Replica StatefulSet.", "Instance", instance)
		return err
	}

	replicaStatefulSet, err := r.resourcesCreator.CreateReplicaStatefulSet(instance)
	if err != nil {
		r.kubegresContext.Log.ErrorEvent("ReplicaStatefulSetTemplateErr", err, "Error while creating a Replica StatefulSet object from template.", "Instance", instance)
		r.blockingOperation.RemoveActiveOperation()
		return err
	}

	r.kubegresContext.Log.Info("Deploying Replica statefulSet '" + replicaStatefulSet.Name + "'")
	err = r.kubegresContext.Client.Create(r.kubegresContext.Ctx, &replicaStatefulSet)
	if err != nil {
		r.kubegresContext.Log.ErrorEvent("ReplicaStatefulSetDeploymentErr", err, "Unable to deploy Replica StatefulSet.", "Replica name", replicaStatefulSet.Name)
		r.blockingOperation.RemoveActiveOperation()
		return err
	}

	r.kubegresContext.Status.SetEnforcedReplicas(r.kubegresContext.Kubegres.Status.EnforcedReplicas + 1)

	r.kubegresContext.Status.SetLastCreatedInstance(instance)
	r.kubegresContext.Log.InfoEvent("ReplicaStatefulSetDeployment", "Deployed Replica StatefulSet.", "Replica name", replicaStatefulSet.Name)
	return nil
}

func (r *ReplicaDbCountSpecEnforcer) activateBlockingOperationForDeployment(instance string) error {
	return r.blockingOperation.ActivateOperationOnStatefulSet(operation.OperationIdReplicaDbCountSpecEnforcement,
		operation.OperationStepIdReplicaDbDeploying,
		instance)
}

func (r *ReplicaDbCountSpecEnforcer) activateBlockingOperationForUndeployment(instance string) error {
	return r.blockingOperation.ActivateOperationOnStatefulSet(operation.OperationIdReplicaDbCountSpecEnforcement,
		operation.OperationStepIdReplicaDbUndeploying,
		instance)
}

func (r *ReplicaDbCountSpecEnforcer) undeployReplicaStatefulSets(replicaToUndeploy statefulset.StatefulSetWrapper) error {
	if replicaToUndeploy.StatefulSet.Name == "" {
		return nil
	}

	r.kubegresContext.Log.Info("We are going to undeploy a Replica statefulSet.", "Instance", replicaToUndeploy.Instance)

	err := r.activateBlockingOperationForUndeployment(replicaToUndeploy.Instance())
	if err != nil {
		r.kubegresContext.Log.ErrorEvent("ReplicaStatefulSetOperationActivationErr", err, "Error while activating blocking operation for the undeployment of a Replica StatefulSet.", "Instance", replicaToUndeploy.Instance)
		return err
	}

	err = r.deleteStatefulSet(replicaToUndeploy.StatefulSet)
	if err != nil {
		r.blockingOperation.RemoveActiveOperation()
		return err
	}

	r.kubegresContext.Status.SetEnforcedReplicas(r.kubegresContext.Kubegres.Status.EnforcedReplicas - 1)

	return nil
}

func (r *ReplicaDbCountSpecEnforcer) getReplicaToDeploy() statefulset.StatefulSetWrapper {
	replicasToUndeploy := r.getReplicasReverseSortedByInstance()

	if len(replicasToUndeploy) == 0 {
		return statefulset.StatefulSetWrapper{}
	}

	return replicasToUndeploy[0]
}

func (r *ReplicaDbCountSpecEnforcer) getReplicaToUndeploy(instance string) statefulset.StatefulSetWrapper {
	replicasToUndeploy, err := r.resourcesStates.StatefulSets.Replicas.All.GetByInstance(instance)
	if err != nil {
		r.kubegresContext.Log.ErrorEvent("ReplicaStatefulSetDeletionErr", err, "Unable to find StatefulSet to delete.", "Instance", instance)
		return statefulset.StatefulSetWrapper{}
	}

	return replicasToUndeploy
}

func (r *ReplicaDbCountSpecEnforcer) getReplicasReverseSortedByInstance() []statefulset.StatefulSetWrapper {
	return r.resourcesStates.StatefulSets.Replicas.All.GetAllReverseSortedByInstance()
}

func (r *ReplicaDbCountSpecEnforcer) deleteStatefulSet(statefulSetToDelete v1.StatefulSet) error {

	r.kubegresContext.Log.Info("Deleting Replica statefulSet", "name", statefulSetToDelete.Name)
	err := r.kubegresContext.Client.Delete(r.kubegresContext.Ctx, &statefulSetToDelete)

	if err != nil {
		r.kubegresContext.Log.ErrorEvent("ReplicaStatefulSetDeletionErr", err, "Unable to delete Replica StatefulSet.", "Replica name", statefulSetToDelete.Name)
		return err
	}

	r.kubegresContext.Log.InfoEvent("ReplicaStatefulSetDeletion", "Deleted Replica StatefulSet.", "Replica name", statefulSetToDelete.Name)
	return nil
}

func (r *ReplicaDbCountSpecEnforcer) logAutomaticFailoverIsDisabled() {
	r.kubegresContext.Log.InfoEvent("AutomaticFailoverIsDisabled",
		"We need to deploy additional Replica(s) because the number of ReplicasCount deployed is less "+
			"than the number of required ReplicasCount in the Spec. "+
			"However, a Replica failover cannot happen because the automatic failover feature is disabled in the YAML. "+
			"To re-enable automatic failover, either set the field 'failover.isDisabled' to false "+
			"or remove that field from the YAML.")
}
