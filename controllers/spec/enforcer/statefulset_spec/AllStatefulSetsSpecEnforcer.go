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
	apps "k8s.io/api/apps/v1"
	postgresV1 "reactive-tech.io/kubegres/api/v1"
	"reactive-tech.io/kubegres/controllers/ctx"
	"reactive-tech.io/kubegres/controllers/operation"
	"reactive-tech.io/kubegres/controllers/spec/enforcer/comparator"
	"reactive-tech.io/kubegres/controllers/states"
	"reactive-tech.io/kubegres/controllers/states/statefulset"
)

// Note about Kubernetes: "Forbidden: updates to Statefulset spec for fields other than 'replicas', 'template',
//and 'updateStrategy' are forbidden"

type AllStatefulSetsSpecEnforcer struct {
	kubegresContext   ctx.KubegresContext
	resourcesStates   states.ResourcesStates
	blockingOperation *operation.BlockingOperation
	specsEnforcers    StatefulSetsSpecsEnforcer
}

func CreateAllStatefulSetsSpecEnforcer(kubegresContext ctx.KubegresContext,
	resourcesStates states.ResourcesStates,
	blockingOperation *operation.BlockingOperation,
	specsEnforcers StatefulSetsSpecsEnforcer) AllStatefulSetsSpecEnforcer {

	return AllStatefulSetsSpecEnforcer{
		kubegresContext:   kubegresContext,
		resourcesStates:   resourcesStates,
		blockingOperation: blockingOperation,
		specsEnforcers:    specsEnforcers,
	}
}

func (r *AllStatefulSetsSpecEnforcer) CreateOperationConfigForStatefulSetSpecUpdating() operation.BlockingOperationConfig {

	return operation.BlockingOperationConfig{
		OperationId:                         operation.OperationIdStatefulSetSpecEnforcing,
		StepId:                              operation.OperationStepIdStatefulSetSpecUpdating,
		TimeOutInSeconds:                    300,
		CompletionChecker:                   r.isStatefulSetSpecUpdated,
		AfterCompletionMoveToTransitionStep: true,
	}
}

func (r *AllStatefulSetsSpecEnforcer) CreateOperationConfigForStatefulSetSpecPodUpdating() operation.BlockingOperationConfig {

	return operation.BlockingOperationConfig{
		OperationId:                         operation.OperationIdStatefulSetSpecEnforcing,
		StepId:                              operation.OperationStepIdStatefulSetPodSpecUpdating,
		TimeOutInSeconds:                    300,
		CompletionChecker:                   r.isStatefulSetPodSpecUpdated,
		AfterCompletionMoveToTransitionStep: true,
	}
}

func (r *AllStatefulSetsSpecEnforcer) CreateOperationConfigForStatefulSetWaitingOnStuckPod() operation.BlockingOperationConfig {

	return operation.BlockingOperationConfig{
		OperationId:                         operation.OperationIdStatefulSetSpecEnforcing,
		StepId:                              operation.OperationStepIdStatefulSetWaitingOnStuckPod,
		TimeOutInSeconds:                    300,
		CompletionChecker:                   r.isStatefulSetPodNotStuck,
		AfterCompletionMoveToTransitionStep: true,
	}
}

func (r *AllStatefulSetsSpecEnforcer) EnforceSpec() error {

	if !r.isPrimaryDbReady() {
		return nil
	}

	if r.blockingOperation.IsActiveOperationIdDifferentOf(operation.OperationIdStatefulSetSpecEnforcing) {
		return nil
	}

	for _, statefulSetWrapper := range r.getAllReverseSortedByInstanceIndex() {

		statefulSet := statefulSetWrapper.StatefulSet
		statefulSetInstanceIndex := statefulSetWrapper.InstanceIndex
		specDifferences := r.specsEnforcers.CheckForSpecDifferences(&statefulSet)

		if r.hasLastSpecUpdateAttemptTimedOut(statefulSetInstanceIndex) {

			if r.areNewSpecChangesSameAsFailingSpecChanges(specDifferences) {
				r.logSpecEnforcementTimedOut()
				return nil

			} else {
				r.blockingOperation.RemoveActiveOperation()
				r.logKubegresFeaturesAreReEnabled()
			}
		}

		if specDifferences.IsThereDifference() {
			return r.enforceSpec(statefulSet, statefulSetInstanceIndex, specDifferences)

		} else if r.isStatefulSetSpecUpdating(statefulSetInstanceIndex) {

			isPodReadyAndSpecUpdated, err := r.verifySpecEnforcementIsAppliedToPod(statefulSetWrapper, specDifferences)
			if err != nil || !isPodReadyAndSpecUpdated {
				return err
			}
		}
	}

	return nil
}

func (r *AllStatefulSetsSpecEnforcer) isPrimaryDbReady() bool {
	return r.resourcesStates.StatefulSets.Primary.IsReady
}

func (r *AllStatefulSetsSpecEnforcer) getAllReverseSortedByInstanceIndex() []statefulset.StatefulSetWrapper {
	replicas := r.resourcesStates.StatefulSets.Replicas.All.GetAllReverseSortedByInstanceIndex()
	return append(replicas, r.resourcesStates.StatefulSets.Primary)
}

func (r *AllStatefulSetsSpecEnforcer) hasLastSpecUpdateAttemptTimedOut(statefulSetInstanceIndex int32) bool {
	if !r.blockingOperation.HasActiveOperationIdTimedOut(operation.OperationIdStatefulSetSpecEnforcing) {
		return false
	}

	activeOperation := r.blockingOperation.GetActiveOperation()
	return activeOperation.StatefulSetOperation.InstanceIndex == statefulSetInstanceIndex
}

func (r *AllStatefulSetsSpecEnforcer) areNewSpecChangesSameAsFailingSpecChanges(newSpecDiff StatefulSetSpecDifferences) bool {
	activeOperation := r.blockingOperation.GetActiveOperation()
	return activeOperation.StatefulSetSpecUpdateOperation.SpecDifferences == newSpecDiff.GetSpecDifferencesAsString()
}

func (r *AllStatefulSetsSpecEnforcer) logKubegresFeaturesAreReEnabled() {
	r.kubegresContext.Log.InfoEvent("KubegresReEnabled", "The new Spec changes "+
		"are different to those which failed during the last spec enforcement."+
		"We can safely re-enable all features of Kubegres and we will try to enforce Spec changes again.")
}

func (r *AllStatefulSetsSpecEnforcer) logSpecEnforcementTimedOut() {

	activeOperation := r.blockingOperation.GetActiveOperation()
	specDifferences := activeOperation.StatefulSetSpecUpdateOperation.SpecDifferences
	StatefulSetName := activeOperation.StatefulSetOperation.Name

	err := errors.New("Spec enforcement timed-out")
	r.kubegresContext.Log.ErrorEvent("StatefulSetSpecEnforcementTimedOutErr", err,
		"Last Spec enforcement attempt has timed-out for a StatefulSet. "+
			"You must apply different spec changes to your Kubegres resource since the previous spec changes did not work. "+
			"Until you apply it, most of the features of Kubegres are disabled for safety reason. ",
		"StatefulSet's name", StatefulSetName,
		"One or many of the following specs failed: ", specDifferences)
}

func (r *AllStatefulSetsSpecEnforcer) enforceSpec(statefulSet apps.StatefulSet,
	instanceIndex int32,
	specDifferences StatefulSetSpecDifferences) error {

	err := r.activateOperationStepStatefulSetSpecUpdating(instanceIndex, specDifferences)
	if err != nil {
		r.kubegresContext.Log.ErrorEvent("StatefulSetSpecEnforcementOperationActivationErr", err, "Error while activating blocking operation for the enforcement of StatefulSet spec.", "StatefulSet name", statefulSet.Name)
		return err
	}

	err = r.specsEnforcers.EnforceSpec(&statefulSet)
	if err != nil {
		r.kubegresContext.Log.ErrorEvent("StatefulSetSpecEnforcementErr", err, "Unable to enforce the Spec of a StatefulSet.", "StatefulSet name", statefulSet.Name)
		r.blockingOperation.RemoveActiveOperation()
		return err
	}

	r.kubegresContext.Log.InfoEvent("StatefulSetOperation", "Enforced the Spec of a StatefulSet.", "StatefulSet name", statefulSet.Name)

	return nil
}

func (r *AllStatefulSetsSpecEnforcer) verifySpecEnforcementIsAppliedToPod(statefulSetWrapper statefulset.StatefulSetWrapper,
	specDifferences StatefulSetSpecDifferences) (isPodReadyAndSpecUpdated bool, err error) {

	podWrapper := statefulSetWrapper.Pod
	podSpecComparator := comparator.PodSpecComparator{Pod: podWrapper.Pod, PostgresSpec: r.kubegresContext.Kubegres.Spec}
	isPodSpecUpToDate := podSpecComparator.IsSpecUpToDate()

	if podWrapper.IsStuck {
		return false, r.handleWhenStuckPod(podWrapper, specDifferences)

	} else if !isPodSpecUpToDate || !podWrapper.IsReady {
		return false, r.handleWhenPodSpecNotUpToDateOrNotReady(podWrapper, isPodSpecUpToDate, specDifferences)
	}

	// SUCCESS: Pod spec was successfully updated
	r.kubegresContext.Log.InfoEvent("PodSpecEnforcement", "Enforced the Spec of a StatefulSet's Pod.", "Pod name", podWrapper.Pod.Name)
	r.blockingOperation.RemoveActiveOperation()
	return true, r.specsEnforcers.OnSpecUpdatedSuccessfully(&statefulSetWrapper.StatefulSet)
}

func (r *AllStatefulSetsSpecEnforcer) handleWhenStuckPod(podWrapper statefulset.PodWrapper,
	specDifferences StatefulSetSpecDifferences) (err error) {

	err = r.activateOperationStepWaitingUntilPodIsNotStuck(podWrapper.InstanceIndex, specDifferences)
	if err != nil {
		r.kubegresContext.Log.ErrorEvent("PodSpecEnforcementOperationActivationErr", err,
			"Error while activating a blocking operation for the enforcement of a Pod's spec. "+
				"We wanted to wait for a Pod which is stuck.",
			"Pod name", podWrapper.Pod.Name)
		return err
	}

	r.kubegresContext.Log.Info("Pod is stuck. Deleting the stuck Pod to create an healthy one.", "Pod name", podWrapper.Pod.Name)
	err = r.kubegresContext.Client.Delete(r.kubegresContext.Ctx, &podWrapper.Pod)
	if err != nil {
		r.kubegresContext.Log.ErrorEvent("PodSpecEnforcementErr", err, "Unable to delete a stuck Pod in order to create an healthy one.", "Pod name", podWrapper.Pod.Name)
		r.blockingOperation.RemoveActiveOperation()
		return err
	}

	return nil
}

func (r *AllStatefulSetsSpecEnforcer) handleWhenPodSpecNotUpToDateOrNotReady(podWrapper statefulset.PodWrapper,
	isPodSpecUpToDate bool,
	specDifferences StatefulSetSpecDifferences) (err error) {

	if !isPodSpecUpToDate {
		r.kubegresContext.Log.Info("Pod has an OLD Spec. Waiting until it is up-to-date.",
			"Pod name", podWrapper.Pod.Name)

	} else if !podWrapper.IsReady {
		r.kubegresContext.Log.Info("Pod is not ready yet. Waiting until it is ready.",
			"Pod name", podWrapper.Pod.Name)
	}

	err = r.activateOperationStepStatefulSetPodSpecUpdating(podWrapper.InstanceIndex, specDifferences)
	if err != nil {
		r.kubegresContext.Log.ErrorEvent("PodSpecEnforcementOperationActivationErr", err,
			"Error while activating a blocking operation for the enforcement of a Pod's spec.",
			"Pod name", podWrapper.Pod.Name)
		return err
	}

	return nil
}

func (r *AllStatefulSetsSpecEnforcer) isStatefulSetSpecUpdating(statefulSetInstanceIndex int32) bool {
	if !r.blockingOperation.IsActiveOperationInTransition(operation.OperationIdStatefulSetSpecEnforcing) {
		return false
	}

	previouslyActiveOperation := r.blockingOperation.GetPreviouslyActiveOperation()
	return previouslyActiveOperation.StatefulSetOperation.InstanceIndex == statefulSetInstanceIndex
}

func (r *AllStatefulSetsSpecEnforcer) isStatefulSetSpecUpdated(operation postgresV1.KubegresBlockingOperation) bool {

	if r.blockingOperation.GetNbreSecondsSinceOperationHasStarted() < 10 {
		return false
	}

	statefulSetInstanceIndex := r.getUpdatingInstanceIndex(operation)
	statefulSetWrapper, err := r.resourcesStates.StatefulSets.All.GetByInstanceIndex(statefulSetInstanceIndex)
	if err != nil {
		r.kubegresContext.Log.InfoEvent("A StatefulSet's instanceIndex does not exist. As a result we will "+
			"return false inside a blocking operation completion checker 'isStatefulSetSpecUpdated()'",
			"instanceIndex", statefulSetInstanceIndex)
		return false
	}

	specDifferences := r.specsEnforcers.CheckForSpecDifferences(&statefulSetWrapper.StatefulSet)
	return !specDifferences.IsThereDifference()
}

func (r *AllStatefulSetsSpecEnforcer) isStatefulSetPodSpecUpdated(operation postgresV1.KubegresBlockingOperation) bool {

	statefulSetInstanceIndex := r.getUpdatingInstanceIndex(operation)
	statefulSetWrapper, err := r.resourcesStates.StatefulSets.All.GetByInstanceIndex(statefulSetInstanceIndex)
	if err != nil {
		r.kubegresContext.Log.InfoEvent("A StatefulSet's instanceIndex does not exist. As a result we will "+
			"return false inside a blocking operation completion checker 'isStatefulSetPodSpecUpdated()'",
			"instanceIndex", statefulSetInstanceIndex)
		return false
	}

	podWrapper := statefulSetWrapper.Pod

	if !podWrapper.IsReady {
		return false
	}

	podSpecComparator := comparator.PodSpecComparator{Pod: podWrapper.Pod, PostgresSpec: r.kubegresContext.Kubegres.Spec}
	return podSpecComparator.IsSpecUpToDate()
}

func (r *AllStatefulSetsSpecEnforcer) isStatefulSetPodNotStuck(operation postgresV1.KubegresBlockingOperation) bool {

	statefulSetInstanceIndex := r.getUpdatingInstanceIndex(operation)
	statefulSetWrapper, err := r.resourcesStates.StatefulSets.All.GetByInstanceIndex(statefulSetInstanceIndex)
	if err != nil {
		r.kubegresContext.Log.InfoEvent("A StatefulSet's instanceIndex does not exist. As a result we will "+
			"return false inside a blocking operation completion checker 'isStatefulSetPodNotStuck()'",
			"instanceIndex", statefulSetInstanceIndex)
		return false
	}

	podWrapper := statefulSetWrapper.Pod
	return !podWrapper.IsStuck
}

func (r *AllStatefulSetsSpecEnforcer) activateOperationStepStatefulSetSpecUpdating(statefulSetInstanceIndex int32,
	specDifferences StatefulSetSpecDifferences) error {

	return r.blockingOperation.ActivateOperationOnStatefulSetSpecUpdate(operation.OperationIdStatefulSetSpecEnforcing,
		operation.OperationStepIdStatefulSetSpecUpdating,
		statefulSetInstanceIndex,
		specDifferences.GetSpecDifferencesAsString())
}

func (r *AllStatefulSetsSpecEnforcer) activateOperationStepStatefulSetPodSpecUpdating(statefulSetInstanceIndex int32,
	specDifferences StatefulSetSpecDifferences) error {

	return r.blockingOperation.ActivateOperationOnStatefulSetSpecUpdate(operation.OperationIdStatefulSetSpecEnforcing,
		operation.OperationStepIdStatefulSetPodSpecUpdating,
		statefulSetInstanceIndex,
		specDifferences.GetSpecDifferencesAsString())
}

func (r *AllStatefulSetsSpecEnforcer) activateOperationStepWaitingUntilPodIsNotStuck(statefulSetInstanceIndex int32,
	specDifferences StatefulSetSpecDifferences) error {

	return r.blockingOperation.ActivateOperationOnStatefulSetSpecUpdate(operation.OperationIdStatefulSetSpecEnforcing,
		operation.OperationStepIdStatefulSetWaitingOnStuckPod,
		statefulSetInstanceIndex,
		specDifferences.GetSpecDifferencesAsString())
}

func (r *AllStatefulSetsSpecEnforcer) getUpdatingInstanceIndex(operation postgresV1.KubegresBlockingOperation) int32 {
	return operation.StatefulSetOperation.InstanceIndex
}
