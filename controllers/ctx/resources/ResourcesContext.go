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

package resources

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	postgresV1 "reactive-tech.io/kubegres/api/v1"
	ctx2 "reactive-tech.io/kubegres/controllers/ctx"
	"reactive-tech.io/kubegres/controllers/ctx/log"
	"reactive-tech.io/kubegres/controllers/ctx/status"
	"reactive-tech.io/kubegres/controllers/operation"
	log3 "reactive-tech.io/kubegres/controllers/operation/log"
	"reactive-tech.io/kubegres/controllers/spec/checker"
	"reactive-tech.io/kubegres/controllers/spec/defaultspec"
	"reactive-tech.io/kubegres/controllers/spec/enforcer/resources_count_spec"
	"reactive-tech.io/kubegres/controllers/spec/enforcer/resources_count_spec/statefulset"
	"reactive-tech.io/kubegres/controllers/spec/enforcer/resources_count_spec/statefulset/failover"
	"reactive-tech.io/kubegres/controllers/spec/enforcer/statefulset_spec"
	"reactive-tech.io/kubegres/controllers/spec/template"
	"reactive-tech.io/kubegres/controllers/states"
	log2 "reactive-tech.io/kubegres/controllers/states/log"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourcesContext struct {
	LogWrapper                   log.LogWrapper
	KubegresStatusWrapper        *status.KubegresStatusWrapper
	KubegresContext              ctx2.KubegresContext
	ResourcesStates              states.ResourcesStates
	ResourcesStatesLogger        log2.ResourcesStatesLogger
	SpecChecker                  checker.SpecChecker
	DefaultStorageClass          defaultspec.DefaultStorageClass
	CustomConfigSpecHelper       template.CustomConfigSpecHelper
	ResourcesCreatorFromTemplate template.ResourcesCreatorFromTemplate
	ResourcesCountSpecEnforcer   resources_count_spec.ResourcesCountSpecEnforcer
	AllStatefulSetsSpecEnforcer  statefulset_spec.AllStatefulSetsSpecEnforcer
	StatefulSetsSpecsEnforcer    statefulset_spec.StatefulSetsSpecsEnforcer

	BlockingOperation          *operation.BlockingOperation
	BlockingOperationLogger    log3.BlockingOperationLogger
	PrimaryToReplicaFailOver   failover.PrimaryToReplicaFailOver
	PrimaryDbCountSpecEnforcer statefulset.PrimaryDbCountSpecEnforcer
	ReplicaDbCountSpecEnforcer statefulset.ReplicaDbCountSpecEnforcer

	BaseConfigMapCountSpecEnforcer resources_count_spec.BaseConfigMapCountSpecEnforcer
	StatefulSetCountSpecEnforcer   resources_count_spec.StatefulSetCountSpecEnforcer
	ServicesCountSpecEnforcer      resources_count_spec.ServicesCountSpecEnforcer
	BackUpCronJobCountSpecEnforcer resources_count_spec.BackUpCronJobCountSpecEnforcer
}

func CreateResourcesContext(kubegres *postgresV1.Kubegres,
	ctx context.Context,
	logger logr.Logger,
	client client.Client,
	recorder record.EventRecorder) (rc *ResourcesContext, err error) {

	setReplicasFieldToZeroIfNil(kubegres)

	rc = &ResourcesContext{}

	rc.LogWrapper = log.LogWrapper{Kubegres: kubegres, Logger: logger, Recorder: recorder}
	rc.LogWrapper.Info("KUBEGRES", "name", kubegres.Name, "Status", kubegres.Status)

	rc.KubegresStatusWrapper = &status.KubegresStatusWrapper{
		Kubegres: kubegres,
		Ctx:      ctx,
		Log:      rc.LogWrapper,
		Client:   client,
	}

	rc.KubegresContext = ctx2.KubegresContext{
		Kubegres: kubegres,
		Status:   rc.KubegresStatusWrapper,
		Ctx:      ctx,
		Log:      rc.LogWrapper,
		Client:   client,
	}

	rc.DefaultStorageClass = defaultspec.CreateDefaultStorageClass(rc.KubegresContext)
	if err = defaultspec.SetDefaultForUndefinedSpecValues(rc.KubegresContext, rc.DefaultStorageClass); err != nil {
		return nil, err
	}

	rc.BlockingOperation = operation.CreateBlockingOperation(rc.KubegresContext)

	rc.BlockingOperationLogger = log3.CreateBlockingOperationLogger(rc.KubegresContext, rc.BlockingOperation)

	if rc.ResourcesStates, err = states.LoadResourcesStates(rc.KubegresContext); err != nil {
		return nil, err
	}

	rc.ResourcesStatesLogger = log2.CreateResourcesStatesLogger(rc.KubegresContext, rc.ResourcesStates)

	rc.SpecChecker = checker.CreateSpecChecker(rc.KubegresContext, rc.ResourcesStates)

	rc.CustomConfigSpecHelper = template.CreateCustomConfigSpecHelper(rc.KubegresContext, rc.ResourcesStates)

	resourceTemplateLoader := template.ResourceTemplateLoader{}
	rc.ResourcesCreatorFromTemplate = template.CreateResourcesCreatorFromTemplate(rc.KubegresContext, rc.CustomConfigSpecHelper, resourceTemplateLoader)

	addResourcesCountSpecEnforcers(rc)
	addStatefulSetSpecEnforcers(rc)
	addBlockingOperationConfigs(rc)

	return rc, nil
}

func setReplicasFieldToZeroIfNil(kubegres *postgresV1.Kubegres) {
	if kubegres.Spec.Replicas != nil {
		return
	}

	replica := int32(0)
	kubegres.Spec.Replicas = &replica
}

func addResourcesCountSpecEnforcers(rc *ResourcesContext) {

	rc.PrimaryToReplicaFailOver = failover.CreatePrimaryToReplicaFailOver(rc.KubegresContext, rc.ResourcesStates, rc.BlockingOperation)
	rc.PrimaryDbCountSpecEnforcer = statefulset.CreatePrimaryDbCountSpecEnforcer(rc.KubegresContext, rc.ResourcesStates, rc.ResourcesCreatorFromTemplate, rc.BlockingOperation, rc.PrimaryToReplicaFailOver)
	rc.ReplicaDbCountSpecEnforcer = statefulset.CreateReplicaDbCountSpecEnforcer(rc.KubegresContext, rc.ResourcesStates, rc.ResourcesCreatorFromTemplate, rc.BlockingOperation)
	rc.StatefulSetCountSpecEnforcer = resources_count_spec.CreateStatefulSetCountSpecEnforcer(rc.PrimaryDbCountSpecEnforcer, rc.ReplicaDbCountSpecEnforcer)

	rc.BaseConfigMapCountSpecEnforcer = resources_count_spec.CreateBaseConfigMapCountSpecEnforcer(rc.KubegresContext, rc.ResourcesStates, rc.ResourcesCreatorFromTemplate, rc.BlockingOperation)
	rc.ServicesCountSpecEnforcer = resources_count_spec.CreateServicesCountSpecEnforcer(rc.KubegresContext, rc.ResourcesStates, rc.ResourcesCreatorFromTemplate)
	rc.BackUpCronJobCountSpecEnforcer = resources_count_spec.CreateBackUpCronJobCountSpecEnforcer(rc.KubegresContext, rc.ResourcesStates, rc.ResourcesCreatorFromTemplate)

	rc.ResourcesCountSpecEnforcer = resources_count_spec.ResourcesCountSpecEnforcer{}
	rc.ResourcesCountSpecEnforcer.AddSpecEnforcer(&rc.BaseConfigMapCountSpecEnforcer)
	rc.ResourcesCountSpecEnforcer.AddSpecEnforcer(&rc.StatefulSetCountSpecEnforcer)
	rc.ResourcesCountSpecEnforcer.AddSpecEnforcer(&rc.ServicesCountSpecEnforcer)
	rc.ResourcesCountSpecEnforcer.AddSpecEnforcer(&rc.BackUpCronJobCountSpecEnforcer)
}

func addStatefulSetSpecEnforcers(rc *ResourcesContext) {
	imageSpecEnforcer := statefulset_spec.CreateImageSpecEnforcer(rc.KubegresContext)
	portSpecEnforcer := statefulset_spec.CreatePortSpecEnforcer(rc.KubegresContext, rc.ResourcesStates)
	storageClassSizeSpecEnforcer := statefulset_spec.CreateStorageClassSizeSpecEnforcer(rc.KubegresContext, rc.ResourcesStates)
	customConfigSpecEnforcer := statefulset_spec.CreateCustomConfigSpecEnforcer(rc.CustomConfigSpecHelper)
	affinitySpecEnforcer := statefulset_spec.CreateAffinitySpecEnforcer(rc.KubegresContext)
	tolerationsSpecEnforcer := statefulset_spec.CreateTolerationsSpecEnforcer(rc.KubegresContext)
	resourcesSpecEnforcer := statefulset_spec.CreateResourcesSpecEnforcer(rc.KubegresContext)
	volumeSpecEnforcer := statefulset_spec.CreateVolumeSpecEnforcer(rc.KubegresContext)
	securityContextSpecEnforcer := statefulset_spec.CreateSecurityContextSpecEnforcer(rc.KubegresContext)
	livenessProbeSpecEnforcer := statefulset_spec.CreateLivenessProbeSpecEnforcer(rc.KubegresContext)
	readinessProbeSpecEnforcer := statefulset_spec.CreateReadinessProbeSpecEnforcer(rc.KubegresContext)

	rc.StatefulSetsSpecsEnforcer = statefulset_spec.CreateStatefulSetsSpecsEnforcer(rc.KubegresContext)
	rc.StatefulSetsSpecsEnforcer.AddSpecEnforcer(&imageSpecEnforcer)
	rc.StatefulSetsSpecsEnforcer.AddSpecEnforcer(&portSpecEnforcer)
	rc.StatefulSetsSpecsEnforcer.AddSpecEnforcer(&storageClassSizeSpecEnforcer)
	rc.StatefulSetsSpecsEnforcer.AddSpecEnforcer(&customConfigSpecEnforcer)
	rc.StatefulSetsSpecsEnforcer.AddSpecEnforcer(&affinitySpecEnforcer)
	rc.StatefulSetsSpecsEnforcer.AddSpecEnforcer(&tolerationsSpecEnforcer)
	rc.StatefulSetsSpecsEnforcer.AddSpecEnforcer(&resourcesSpecEnforcer)
	rc.StatefulSetsSpecsEnforcer.AddSpecEnforcer(&volumeSpecEnforcer)
	rc.StatefulSetsSpecsEnforcer.AddSpecEnforcer(&securityContextSpecEnforcer)
	rc.StatefulSetsSpecsEnforcer.AddSpecEnforcer(&livenessProbeSpecEnforcer)
	rc.StatefulSetsSpecsEnforcer.AddSpecEnforcer(&readinessProbeSpecEnforcer)

	rc.AllStatefulSetsSpecEnforcer = statefulset_spec.CreateAllStatefulSetsSpecEnforcer(rc.KubegresContext, rc.ResourcesStates, rc.BlockingOperation, rc.StatefulSetsSpecsEnforcer)
}

func addBlockingOperationConfigs(rc *ResourcesContext) {

	rc.BlockingOperation.AddConfig(rc.BaseConfigMapCountSpecEnforcer.CreateOperationConfig())

	rc.BlockingOperation.AddConfig(rc.PrimaryDbCountSpecEnforcer.CreateOperationConfigForPrimaryDbDeploying())
	rc.BlockingOperation.AddConfig(rc.PrimaryToReplicaFailOver.CreateOperationConfigWaitingBeforeForFailingOver())
	rc.BlockingOperation.AddConfig(rc.PrimaryToReplicaFailOver.CreateOperationConfigForFailingOver())

	rc.BlockingOperation.AddConfig(rc.ReplicaDbCountSpecEnforcer.CreateOperationConfigForReplicaDbDeploying())
	rc.BlockingOperation.AddConfig(rc.ReplicaDbCountSpecEnforcer.CreateOperationConfigForReplicaDbUndeploying())

	rc.BlockingOperation.AddConfig(rc.AllStatefulSetsSpecEnforcer.CreateOperationConfigForStatefulSetSpecUpdating())
	rc.BlockingOperation.AddConfig(rc.AllStatefulSetsSpecEnforcer.CreateOperationConfigForStatefulSetSpecPodUpdating())
	rc.BlockingOperation.AddConfig(rc.AllStatefulSetsSpecEnforcer.CreateOperationConfigForStatefulSetWaitingOnStuckPod())
}
