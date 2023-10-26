/*
Copyright 2023 Reactive Tech Limited.
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

package operation

import (
	v1 "reactive-tech.io/kubegres/api/v1"
	"reactive-tech.io/kubegres/internal/controller/ctx"
	"time"
)

type BlockingOperation struct {
	kubegresContext           ctx.KubegresContext
	configs                   map[string]BlockingOperationConfig
	activeOperation           v1.KubegresBlockingOperation
	previouslyActiveOperation v1.KubegresBlockingOperation
}

func CreateBlockingOperation(kubegresContext ctx.KubegresContext) *BlockingOperation {

	return &BlockingOperation{
		kubegresContext: kubegresContext,
		configs:         make(map[string]BlockingOperationConfig),
		activeOperation: v1.KubegresBlockingOperation{},
	}
}

func (r *BlockingOperation) AddConfig(config BlockingOperationConfig) {
	configId := r.generateConfigId(config.OperationId, config.StepId)
	r.configs[configId] = config
}

func (r *BlockingOperation) LoadActiveOperation() int64 {

	r.activeOperation = r.kubegresContext.Status.GetBlockingOperation()
	r.previouslyActiveOperation = r.kubegresContext.Status.GetPreviousBlockingOperation()
	r.removeOperationIfNotActive()

	nbreSecondsLeftBeforeTimeOut := r.GetNbreSecondsLeftBeforeTimeOut()
	if nbreSecondsLeftBeforeTimeOut > 20 {
		return 20
	}

	return nbreSecondsLeftBeforeTimeOut
}

func (r *BlockingOperation) IsActiveOperationIdDifferentOf(operationId string) bool {
	return r.isThereActiveOperation() && r.activeOperation.OperationId != operationId
}

func (r *BlockingOperation) IsActiveOperationInTransition(operationId string) bool {
	return r.activeOperation.OperationId == operationId &&
		r.activeOperation.StepId == TransitionOperationStepId
}

func (r *BlockingOperation) HasActiveOperationIdTimedOut(operationId string) bool {
	return r.activeOperation.OperationId == operationId &&
		r.activeOperation.HasTimedOut
}

func (r *BlockingOperation) GetActiveOperation() v1.KubegresBlockingOperation {
	return r.activeOperation
}

func (r *BlockingOperation) GetPreviouslyActiveOperation() v1.KubegresBlockingOperation {
	return r.previouslyActiveOperation
}

func (r *BlockingOperation) ActivateOperation(operationId string, stepId string) error {
	return r.activateOperation(r.createOperationObj(operationId, stepId))
}

func (r *BlockingOperation) ActivateOperationOnStatefulSet(operationId string, stepId string, statefulSetInstanceIndex int32) error {
	blockingOperation := r.createOperationObj(operationId, stepId)
	blockingOperation.StatefulSetOperation = r.createStatefulSetOperationObj(statefulSetInstanceIndex)
	return r.activateOperation(blockingOperation)
}

func (r *BlockingOperation) ActivateOperationOnStatefulSetSpecUpdate(operationId string, stepId string,
	statefulSetInstanceIndex int32, specDifferences string) error {

	blockingOperation := r.createOperationObj(operationId, stepId)
	blockingOperation.StatefulSetOperation = r.createStatefulSetOperationObj(statefulSetInstanceIndex)
	blockingOperation.StatefulSetSpecUpdateOperation = r.createStatefulSetSpecUpdateOperationObj(specDifferences)

	return r.activateOperation(blockingOperation)
}

func (r *BlockingOperation) RemoveActiveOperation() {
	r.removeActiveOperation(false)
}

func (r *BlockingOperation) isThereActiveOperation() bool {
	return r.activeOperation.OperationId != ""
}

func (r *BlockingOperation) GetNbreSecondsSinceOperationHasStarted() int64 {
	configTimeOutInSeconds := r.getConfig(r.activeOperation).TimeOutInSeconds
	nbreSecondsLeftBeforeTimeOut := r.GetNbreSecondsLeftBeforeTimeOut()
	return configTimeOutInSeconds - nbreSecondsLeftBeforeTimeOut
}

func (r *BlockingOperation) GetNbreSecondsLeftBeforeTimeOut() int64 {
	if r.activeOperation.TimeOutEpocInSeconds == 0 {
		return 0
	}

	nbreSecondsLeftBeforeTimeOut := r.activeOperation.TimeOutEpocInSeconds - r.getNbreSecondsSinceEpoc()
	if nbreSecondsLeftBeforeTimeOut < 0 {
		return 0
	}

	return nbreSecondsLeftBeforeTimeOut
}

func (r *BlockingOperation) GetNbreSecondsSinceTimedOut() int64 {
	return r.getNbreSecondsSinceEpoc() - r.activeOperation.TimeOutEpocInSeconds
}

func (r *BlockingOperation) activateOperation(operation v1.KubegresBlockingOperation) error {

	if r.isThereActiveOperation() && r.activeOperation.OperationId != operation.OperationId {
		return CreateBlockingOperationError(errorTypeThereIsAlreadyAnActiveOperation, operation.OperationId)

	} else if !r.doesOperationHaveAssociatedConfig(operation) {
		return CreateBlockingOperationError(errorTypeOperationIdHasNoAssociatedConfig, operation.OperationId)
	}

	config := r.getConfig(operation)
	operation.TimeOutEpocInSeconds = r.calculateTimeOutInEpochSeconds(config.TimeOutInSeconds)
	r.activeOperation = operation
	r.kubegresContext.Status.SetBlockingOperation(operation)
	return nil
}

func (r *BlockingOperation) createOperationObj(operationId string, stepId string) v1.KubegresBlockingOperation {
	return v1.KubegresBlockingOperation{OperationId: operationId, StepId: stepId}
}

func (r *BlockingOperation) createStatefulSetOperationObj(statefulSetInstanceIndex int32) v1.KubegresStatefulSetOperation {
	return v1.KubegresStatefulSetOperation{
		InstanceIndex: statefulSetInstanceIndex,
		Name:          r.kubegresContext.GetStatefulSetResourceName(statefulSetInstanceIndex),
	}
}

func (r *BlockingOperation) createStatefulSetSpecUpdateOperationObj(specDifferences string) v1.KubegresStatefulSetSpecUpdateOperation {
	return v1.KubegresStatefulSetSpecUpdateOperation{SpecDifferences: specDifferences}
}

func (r *BlockingOperation) removeActiveOperation(hasOperationTimedOut bool) {
	if !r.isOperationInTransition() {
		r.previouslyActiveOperation = r.activeOperation
		r.previouslyActiveOperation.HasTimedOut = hasOperationTimedOut
		r.kubegresContext.Status.SetPreviousBlockingOperation(r.previouslyActiveOperation)
	}

	r.activeOperation = v1.KubegresBlockingOperation{}
	r.kubegresContext.Status.SetBlockingOperation(r.activeOperation)
}

func (r *BlockingOperation) setActiveOperationInTransition(hasOperationTimedOut bool) {
	r.previouslyActiveOperation = r.activeOperation
	r.previouslyActiveOperation.HasTimedOut = hasOperationTimedOut
	r.kubegresContext.Status.SetPreviousBlockingOperation(r.previouslyActiveOperation)

	r.activeOperation.StepId = TransitionOperationStepId
	r.activeOperation.HasTimedOut = false
	r.activeOperation.TimeOutEpocInSeconds = 0
	r.kubegresContext.Status.SetBlockingOperation(r.activeOperation)
}

func (r *BlockingOperation) removeOperationIfNotActive() {
	if !r.isThereActiveOperation() || r.isOperationInTransition() {
		return
	}

	hasOperationTimedOut := r.hasOperationTimedOut()

	if hasOperationTimedOut {
		r.activeOperation.HasTimedOut = true

		if !r.hasCompletionChecker() {

			if r.shouldOperationBeInTransition() {
				r.setActiveOperationInTransition(hasOperationTimedOut)
			} else {
				r.removeActiveOperation(hasOperationTimedOut)
			}

		} else {
			r.kubegresContext.Log.InfoEvent("BlockingOperationTimedOut", "Blocking-Operation timed-out.", "OperationId", r.activeOperation.OperationId, "StepId", r.activeOperation.StepId)
		}

	} else if r.isOperationCompletionConditionReached() {

		r.kubegresContext.Log.InfoEvent("BlockingOperationCompleted", "Blocking-Operation is successfully completed.", "OperationId", r.activeOperation.OperationId, "StepId", r.activeOperation.StepId)

		if r.shouldOperationBeInTransition() {
			r.setActiveOperationInTransition(hasOperationTimedOut)
		} else {
			r.removeActiveOperation(hasOperationTimedOut)
		}
	}
}

func (r *BlockingOperation) isOperationInTransition() bool {
	return r.activeOperation.StepId == TransitionOperationStepId
}

func (r *BlockingOperation) hasCompletionChecker() bool {
	operationConfig := r.getConfig(r.activeOperation)
	return operationConfig.CompletionChecker != nil
}

func (r *BlockingOperation) shouldOperationBeInTransition() bool {
	operationConfig := r.getConfig(r.activeOperation)
	return operationConfig.AfterCompletionMoveToTransitionStep
}

func (r *BlockingOperation) hasOperationTimedOut() bool {
	return r.activeOperation.TimeOutEpocInSeconds-r.getNbreSecondsSinceEpoc() <= 0
}

func (r *BlockingOperation) isOperationCompletionConditionReached() bool {

	operationConfig := r.getConfig(r.activeOperation)
	if operationConfig.CompletionChecker != nil {
		return operationConfig.CompletionChecker(r.activeOperation)
	}

	return false
}

func (r *BlockingOperation) calculateTimeOutInEpochSeconds(timeOutInSeconds int64) int64 {
	return r.getNbreSecondsSinceEpoc() + timeOutInSeconds
}

func (r *BlockingOperation) getNbreSecondsSinceEpoc() int64 {
	return time.Now().Unix()
}

func (r *BlockingOperation) getConfig(operation v1.KubegresBlockingOperation) BlockingOperationConfig {
	configId := r.generateConfigId(operation.OperationId, operation.StepId)
	return r.configs[configId]
}

func (r *BlockingOperation) generateConfigId(operationId, stepId string) string {
	return operationId + "_" + stepId
}

func (r *BlockingOperation) doesOperationHaveAssociatedConfig(operation v1.KubegresBlockingOperation) bool {
	config := r.getConfig(operation)
	return config.OperationId == operation.OperationId
}
