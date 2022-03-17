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

package failover

import (
	"errors"
	core "k8s.io/api/core/v1"
	v1 "reactive-tech.io/kubegres/api/v1"
	"reactive-tech.io/kubegres/controllers/ctx"
	"reactive-tech.io/kubegres/controllers/operation"
	"reactive-tech.io/kubegres/controllers/states"
	"reactive-tech.io/kubegres/controllers/states/statefulset"
	"strconv"
)

type PrimaryToReplicaFailOver struct {
	kubegresContext   ctx.KubegresContext
	resourcesStates   states.ResourcesStates
	blockingOperation *operation.BlockingOperation
}

func CreatePrimaryToReplicaFailOver(kubegresContext ctx.KubegresContext,
	resourcesStates states.ResourcesStates,
	blockingOperation *operation.BlockingOperation) PrimaryToReplicaFailOver {

	return PrimaryToReplicaFailOver{
		kubegresContext:   kubegresContext,
		resourcesStates:   resourcesStates,
		blockingOperation: blockingOperation,
	}
}

func (r *PrimaryToReplicaFailOver) CreateOperationConfigWaitingBeforeForFailingOver() operation.BlockingOperationConfig {
	return operation.BlockingOperationConfig{
		OperationId:                         operation.OperationIdPrimaryDbCountSpecEnforcement,
		StepId:                              operation.OperationStepIdPrimaryDbWaitingBeforeFailingOver,
		TimeOutInSeconds:                    10,
		AfterCompletionMoveToTransitionStep: true,
	}
}

func (r *PrimaryToReplicaFailOver) CreateOperationConfigForFailingOver() operation.BlockingOperationConfig {
	return operation.BlockingOperationConfig{
		OperationId:       operation.OperationIdPrimaryDbCountSpecEnforcement,
		StepId:            operation.OperationStepIdPrimaryDbFailingOver,
		TimeOutInSeconds:  300,
		CompletionChecker: r.isFailOverCompleted,
	}
}

func (r *PrimaryToReplicaFailOver) ShouldWeFailOver() bool {
	if !r.hasPrimaryEverBeenDeployed() {
		return false

	} else if !r.isThereReadyReplica() {
		r.logFailoverCannotHappenAsNoReplicaDeployed()
		return false
	} else if r.isManualFailoverRequested() {
		return true

	} else if r.isNewPrimaryRequired() {
		if r.isAutomaticFailoverDisabled() {
			r.logFailoverCannotHappenAsAutomaticFailoverIsDisabled()
			return false
		}
		return true
	}

	return false
}

func (r *PrimaryToReplicaFailOver) FailOver() error {
	if r.blockingOperation.IsActiveOperationIdDifferentOf(operation.OperationIdPrimaryDbCountSpecEnforcement) {
		return nil
	}

	if r.hasLastFailOverAttemptTimedOut() {
		r.logFailoverTimedOut()
		return nil
	}

	var newPrimary, err = r.selectReplicaToPromote()
	if err != nil {
		return err
	}

	if !r.isWaitingBeforeStartingFailOver() {
		return r.waitBeforePromotingReplicaToPrimary(newPrimary)
	} else {
		return r.promoteReplicaToPrimary(newPrimary)
	}
}

func (r *PrimaryToReplicaFailOver) isFailOverCompleted(operation v1.KubegresBlockingOperation) bool {

	if r.blockingOperation.GetNbreSecondsSinceOperationHasStarted() < 40 {

		if r.isPrimaryDbReady() {
			r.kubegresContext.Log.Info("The new Primary Pod is ready. " +
				"We are waiting on the connections between pods to be ready before completing the failover process.")
		}

		return false
	}

	return r.isPrimaryDbReady()
}

func (r *PrimaryToReplicaFailOver) isNewPrimaryRequired() bool {
	return !r.isPrimaryDbDeployed() || !r.isPrimaryDbReady()
}

func (r *PrimaryToReplicaFailOver) isPrimaryDbReady() bool {
	return r.resourcesStates.StatefulSets.Primary.IsReady
}

func (r *PrimaryToReplicaFailOver) isThereReadyReplica() bool {
	return r.resourcesStates.StatefulSets.Replicas.NumberReady > 0
}

func (r *PrimaryToReplicaFailOver) isAutomaticFailoverDisabled() bool {
	return r.kubegresContext.Kubegres.Spec.Failover.IsDisabled
}

func (r *PrimaryToReplicaFailOver) isPrimaryDbDeployed() bool {
	return r.resourcesStates.StatefulSets.Primary.IsDeployed
}

func (r *PrimaryToReplicaFailOver) hasLastFailOverAttemptTimedOut() bool {
	return r.blockingOperation.HasActiveOperationIdTimedOut(operation.OperationIdPrimaryDbCountSpecEnforcement)
}

func (r *PrimaryToReplicaFailOver) logFailoverTimedOut() {

	activeOperation := r.blockingOperation.GetActiveOperation()
	operationTimeOutStr := strconv.FormatInt(r.CreateOperationConfigForFailingOver().TimeOutInSeconds, 10)
	primaryStatefulSetName := activeOperation.StatefulSetOperation.Name

	err := errors.New("FailOver timed-out")
	r.kubegresContext.Log.ErrorEvent("FailOverTimedOutErr", err,
		"Last FailOver attempt has timed-out after "+operationTimeOutStr+" seconds. "+
			"The new Primary DB is still NOT ready. It must be fixed manually. "+
			"Until the PrimaryDB is ready, most of the features of Kubegres are disabled for safety reason. ",
		"Primary DB StatefulSet to fix", primaryStatefulSetName)
}

func (r *PrimaryToReplicaFailOver) getPodToManuallyPromote() string {
	return r.kubegresContext.Kubegres.Spec.Failover.PromotePod
}

func (r *PrimaryToReplicaFailOver) getInstanceToManuallyPromote() string {
	podToPromote := r.getPodToManuallyPromote()
	if podToPromote == "" || podToPromote == r.resourcesStates.StatefulSets.Primary.Pod.Pod.Name {
		return ""
	}

	for _, statefulSetWrapper := range r.resourcesStates.StatefulSets.Replicas.All.GetAllSortedByInstance() {
		if podToPromote == statefulSetWrapper.Pod.Pod.Name {
			return statefulSetWrapper.Instance()
		}
	}

	r.logManualFailoverCannotHappenAsConfigErr()
	return ""
}

func (r *PrimaryToReplicaFailOver) getPrimaryInstance() string {
	return r.resourcesStates.StatefulSets.Primary.Instance()
}

func (r *PrimaryToReplicaFailOver) isManualFailoverRequested() bool {
	return r.getInstanceToManuallyPromote() != "" &&
		r.getInstanceToManuallyPromote() != r.getPrimaryInstance()
}

func (r *PrimaryToReplicaFailOver) isWaitingBeforeStartingFailOver() bool {
	if !r.blockingOperation.IsActiveOperationInTransition(operation.OperationIdPrimaryDbCountSpecEnforcement) {
		return false
	}

	previouslyActiveOperation := r.blockingOperation.GetPreviouslyActiveOperation()
	return previouslyActiveOperation.StepId == operation.OperationStepIdPrimaryDbWaitingBeforeFailingOver
}

func (r *PrimaryToReplicaFailOver) selectReplicaToPromote() (statefulset.StatefulSetWrapper, error) {
	if r.isManualFailoverRequested() {
		return r.manuallySelectReplicaToPromote()
	}

	for _, statefulSetWrapper := range r.resourcesStates.StatefulSets.Replicas.All.GetAllSortedByInstance() {
		if statefulSetWrapper.IsReady {
			return statefulSetWrapper, nil
		}
	}

	errorMsg := r.logFailoverCannotHappenAsNoHealthyReplica()
	return statefulset.StatefulSetWrapper{}, errors.New(errorMsg)
}

func (r *PrimaryToReplicaFailOver) manuallySelectReplicaToPromote() (statefulset.StatefulSetWrapper, error) {

	replicaInstanceToPromote := r.getInstanceToManuallyPromote()
	r.logManualFailoverIsRequested()

	for _, statefulSetWrapper := range r.resourcesStates.StatefulSets.Replicas.All.GetAllSortedByInstance() {
		if statefulSetWrapper.IsReady && statefulSetWrapper.Instance() == replicaInstanceToPromote {
			return statefulSetWrapper, nil
		}
	}

	errorMsg := r.logManualFailoverCannotHappenAsConfigErr()
	return statefulset.StatefulSetWrapper{}, errors.New(errorMsg)
}

func (r *PrimaryToReplicaFailOver) promoteReplicaToPrimary(newPrimary statefulset.StatefulSetWrapper) error {
	newPrimary.StatefulSet.Labels[ctx.ReplicationRoleLabelKey] = ctx.PrimaryRoleName
	newPrimary.StatefulSet.Spec.Template.Labels[ctx.ReplicationRoleLabelKey] = ctx.PrimaryRoleName
	volumeMount := core.VolumeMount{
		Name:      "base-config",
		MountPath: "/tmp/promote_replica_to_primary.sh",
		SubPath:   "promote_replica_to_primary.sh",
	}

	initContainer := &newPrimary.StatefulSet.Spec.Template.Spec.InitContainers[0]
	initContainer.VolumeMounts = append(initContainer.VolumeMounts, volumeMount)
	initContainer.Command = []string{"sh", "-c", "/tmp/promote_replica_to_primary.sh"}

	err := r.activateOperationFailingOver(newPrimary)
	if err != nil {
		r.kubegresContext.Log.ErrorEvent("FailOverOperationActivationErr", err,
			"Error while activating a blocking operation for the FailOver of a Primary DB.",
			"Instance", newPrimary.Instance)
		return err
	}

	r.kubegresContext.Log.InfoEvent("FailOver", "FailOver: Promoting Replica to Primary.",
		"Replica to promote", newPrimary.StatefulSet.Name)

	err2 := r.kubegresContext.Client.Update(r.kubegresContext.Ctx, &newPrimary.StatefulSet)
	if err2 != nil {
		r.kubegresContext.Log.ErrorEvent("FailOverErr", err2,
			"FailOver: Unable to promote Replica to Primary.",
			"Replica to promote", newPrimary.StatefulSet.Name)
		r.blockingOperation.RemoveActiveOperation()
		return err
	}

	return nil
}

func (r *PrimaryToReplicaFailOver) waitBeforePromotingReplicaToPrimary(newPrimary statefulset.StatefulSetWrapper) error {

	r.deletePrimaryStatefulSet()

	err := r.activateOperationWaitingBeforeFailingOver(newPrimary)
	if err != nil {
		r.kubegresContext.Log.ErrorEvent("FailOverOperationActivationErr", err,
			"Error while activating a blocking operation to wait before starting the FailOver of a Primary DB.",
			"Instance", newPrimary.Instance)
		return err
	}

	r.kubegresContext.Log.Info("FailOver: Waiting before promoting a Replica to a Primary...",
		"Replica to promote", newPrimary.StatefulSet.Name)
	return nil
}

func (r *PrimaryToReplicaFailOver) activateOperationWaitingBeforeFailingOver(newPrimary statefulset.StatefulSetWrapper) error {
	return r.blockingOperation.ActivateOperationOnStatefulSet(operation.OperationIdPrimaryDbCountSpecEnforcement,
		operation.OperationStepIdPrimaryDbWaitingBeforeFailingOver,
		newPrimary.Instance())
}

func (r *PrimaryToReplicaFailOver) activateOperationFailingOver(newPrimary statefulset.StatefulSetWrapper) error {
	return r.blockingOperation.ActivateOperationOnStatefulSet(operation.OperationIdPrimaryDbCountSpecEnforcement,
		operation.OperationStepIdPrimaryDbFailingOver,
		newPrimary.Instance())
}

func (r *PrimaryToReplicaFailOver) deletePrimaryStatefulSet() {

	statefulSetToDelete := r.resourcesStates.StatefulSets.Primary.StatefulSet
	r.kubegresContext.Log.Info("FailOver: Deleting the failing Primary StatefulSet.",
		"Primary name", statefulSetToDelete.Name)

	err := r.kubegresContext.Client.Delete(r.kubegresContext.Ctx, &statefulSetToDelete)
	if err == nil {
		r.kubegresContext.Log.InfoEvent("FailOverPrimaryDeleted",
			"Deleted the failing Primary StatefulSet.",
			"Primary name", statefulSetToDelete.Name)
	}
}

func (r *PrimaryToReplicaFailOver) logFailoverCannotHappenAsNoReplicaDeployed() {
	message := ""
	if r.isManualFailoverRequested() {
		message = "A manual failover to promote a Replica as a Primary was requested."
	} else if r.isNewPrimaryRequired() {
		message = "A failover is required for a Primary Pod as it is not healthy."
	}

	if message != "" {
		r.kubegresContext.Log.InfoEvent("FailoverCannotHappenAsNoReplicaDeployed",
			message+" However, a failover cannot happen because there is not any Replica deployed.")
	}
}

func (r *PrimaryToReplicaFailOver) hasPrimaryEverBeenDeployed() bool {
	return r.kubegresContext.Kubegres.Status.EnforcedReplicas > 0
}

func (r *PrimaryToReplicaFailOver) logFailoverCannotHappenAsAutomaticFailoverIsDisabled() {
	r.kubegresContext.Log.InfoEvent("AutomaticFailoverIsDisabled",
		"A failover is required for a Primary Pod as it is not healthy. "+
			"However, a failover cannot happen because the automatic failover feature is disabled in the YAML. "+
			"To re-enable automatic failover, either set the field 'failover.isDisabled' to false "+
			"or remove that field from the YAML.")
}

func (r *PrimaryToReplicaFailOver) logManualFailoverIsRequested() {
	r.kubegresContext.Log.InfoEvent("ManualFailover",
		"A manual failover to promote a Replica as a Primary was requested.")
}

func (r *PrimaryToReplicaFailOver) logFailoverCannotHappenAsNoHealthyReplica() string {
	errorReason := "FailoverCannotHappenAsNotFoundHealthyReplicaErr"
	errorMsg := "We cannot Failover to a Replica because there are not any ReplicasCount which are ready to serve requests. " +
		"Primary has to be fixed manually."
	r.kubegresContext.Log.ErrorEvent(errorReason, errors.New(""), errorMsg)
	return errorMsg
}

func (r *PrimaryToReplicaFailOver) logManualFailoverCannotHappenAsConfigErr() string {
	errorReason := "ManualFailoverCannotHappenAsConfigErr"
	errorMsg := "The value of the field 'failover.promotePod' is set to '" + r.getPodToManuallyPromote() + "'. " +
		"That value is either the name of a Primary Pod OR a Replica Pod which is not ready OR a Pod which does not exist. " +
		"Please set the name of a Replica Pod that you would like to promote as a Primary Pod."
	r.kubegresContext.Log.WarningEvent(errorReason, errorMsg)
	return errorMsg
}
