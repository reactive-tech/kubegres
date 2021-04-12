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

package resources_count_spec

import (
	"errors"
	core "k8s.io/api/core/v1"
	v1 "reactive-tech.io/kubegres/api/v1"
	"reactive-tech.io/kubegres/controllers/ctx"
	"reactive-tech.io/kubegres/controllers/operation"
	"reactive-tech.io/kubegres/controllers/spec/template"
	"reactive-tech.io/kubegres/controllers/states"
	"strconv"
)

type BaseConfigMapCountSpecEnforcer struct {
	kubegresContext   ctx.KubegresContext
	resourcesStates   states.ResourcesStates
	resourcesCreator  template.ResourcesCreatorFromTemplate
	blockingOperation *operation.BlockingOperation
}

func CreateBaseConfigMapCountSpecEnforcer(kubegresContext ctx.KubegresContext,
	resourcesStates states.ResourcesStates,
	resourcesCreator template.ResourcesCreatorFromTemplate,
	blockingOperation *operation.BlockingOperation) BaseConfigMapCountSpecEnforcer {

	return BaseConfigMapCountSpecEnforcer{
		kubegresContext:   kubegresContext,
		resourcesStates:   resourcesStates,
		resourcesCreator:  resourcesCreator,
		blockingOperation: blockingOperation,
	}
}

func (r *BaseConfigMapCountSpecEnforcer) CreateOperationConfig() operation.BlockingOperationConfig {

	return operation.BlockingOperationConfig{
		OperationId:       operation.OperationIdBaseConfigCountSpecEnforcement,
		StepId:            operation.OperationStepIdBaseConfigDeploying,
		TimeOutInSeconds:  10,
		CompletionChecker: func(operation v1.KubegresBlockingOperation) bool { return r.isBaseConfigDeployed() },
	}
}

func (r *BaseConfigMapCountSpecEnforcer) EnforceSpec() error {

	if r.blockingOperation.IsActiveOperationIdDifferentOf(operation.OperationIdBaseConfigCountSpecEnforcement) {
		return nil
	}

	if r.hasLastAttemptTimedOut() {

		if r.isBaseConfigDeployed() {
			r.blockingOperation.RemoveActiveOperation()
			r.logKubegresFeaturesAreReEnabled()

		} else {
			r.logTimedOut()
			return nil
		}
	}

	if r.isBaseConfigDeployed() {
		return nil
	}

	baseConfigMap, err := r.resourcesCreator.CreateBaseConfigMap()
	if err != nil {
		r.kubegresContext.Log.ErrorEvent("BaseConfigMapTemplateErr", err,
			"Unable to create a Base ConfigMap object from template.",
			"Based ConfigMap name", ctx.BaseConfigMapName)
		return err
	}

	return r.deployBaseConfigMap(baseConfigMap)
}

func (r *BaseConfigMapCountSpecEnforcer) isBaseConfigDeployed() bool {
	return r.resourcesStates.Config.IsBaseConfigDeployed
}

func (r *BaseConfigMapCountSpecEnforcer) hasLastAttemptTimedOut() bool {
	return r.blockingOperation.HasActiveOperationIdTimedOut(operation.OperationIdBaseConfigCountSpecEnforcement)
}

func (r *BaseConfigMapCountSpecEnforcer) logKubegresFeaturesAreReEnabled() {
	r.kubegresContext.Log.InfoEvent("KubegresReEnabled", "Base ConfigMap is available again. "+
		"We can safely re-enable all features of Kubegres.")
}

func (r *BaseConfigMapCountSpecEnforcer) logTimedOut() {

	operationTimeOutStr := strconv.FormatInt(r.CreateOperationConfig().TimeOutInSeconds, 10)

	err := errors.New("Base ConfigMap deployment timed-out")
	r.kubegresContext.Log.ErrorEvent("BaseConfigMapDeploymentTimedOutErr", err,
		"Last Base ConfigMap deployment attempt has timed-out after "+operationTimeOutStr+" seconds. "+
			"The new Base ConfigMap is still NOT ready. It must be fixed manually. "+
			"Until Base ConfigMap is ready, most of the features of Kubegres are disabled for safety reason. ",
		"Based ConfigMap to fix", ctx.BaseConfigMapName)
}

func (r *BaseConfigMapCountSpecEnforcer) deployBaseConfigMap(configMap core.ConfigMap) error {

	r.kubegresContext.Log.Info("Deploying Base ConfigMap", "name", configMap.Name)

	if err := r.activateBlockingOperationForDeployment(); err != nil {
		r.kubegresContext.Log.ErrorEvent("BaseConfigMapDeploymentOperationActivationErr", err,
			"Error while activating blocking operation for the deployment of a Base ConfigMap.",
			"ConfigMap name", configMap.Name)
		return err
	}

	if err := r.kubegresContext.Client.Create(r.kubegresContext.Ctx, &configMap); err != nil {
		r.kubegresContext.Log.ErrorEvent("BaseConfigMapDeploymentErr", err,
			"Unable to deploy Base ConfigMap.",
			"ConfigMap name", configMap.Name)
		r.blockingOperation.RemoveActiveOperation()
		return err
	}

	r.kubegresContext.Log.InfoEvent("BaseConfigMapDeployment", "Deployed Base ConfigMap.",
		"ConfigMap name", configMap.Name)
	return nil
}

func (r *BaseConfigMapCountSpecEnforcer) activateBlockingOperationForDeployment() error {
	return r.blockingOperation.ActivateOperation(operation.OperationIdBaseConfigCountSpecEnforcement,
		operation.OperationStepIdBaseConfigDeploying)
}
