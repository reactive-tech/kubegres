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

package resources

import (
	"context"

	ctx2 "reactive-tech.io/kubegres/internal/controller/ctx"
	"reactive-tech.io/kubegres/internal/controller/ctx/log"
	"reactive-tech.io/kubegres/internal/controller/ctx/status"
	"reactive-tech.io/kubegres/internal/controller/operation"
	log3 "reactive-tech.io/kubegres/internal/controller/operation/log"
	"reactive-tech.io/kubegres/internal/controller/spec/checker"
	defaultspec2 "reactive-tech.io/kubegres/internal/controller/spec/defaultspec"
	resources_count_spec2 "reactive-tech.io/kubegres/internal/controller/spec/enforcer/resources_count_spec"
	statefulset2 "reactive-tech.io/kubegres/internal/controller/spec/enforcer/resources_count_spec/statefulset"
	"reactive-tech.io/kubegres/internal/controller/spec/enforcer/resources_count_spec/statefulset/failover"
	statefulset_spec2 "reactive-tech.io/kubegres/internal/controller/spec/enforcer/statefulset_spec"
	template2 "reactive-tech.io/kubegres/internal/controller/spec/template"
	"reactive-tech.io/kubegres/internal/controller/states"
	log2 "reactive-tech.io/kubegres/internal/controller/states/log"

	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	postgresV1 "reactive-tech.io/kubegres/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourcesContext struct {
	LogWrapper                   log.LogWrapper
	KubegresStatusWrapper        *status.KubegresStatusWrapper
	KubegresContext              ctx2.KubegresContext
	ResourcesStates              states.ResourcesStates
	ResourcesStatesLogger        log2.ResourcesStatesLogger
	SpecChecker                  checker.SpecChecker
	DefaultStorageClass          defaultspec2.DefaultStorageClass
	CustomConfigSpecHelper       template2.CustomConfigSpecHelper
	ResourcesCreatorFromTemplate template2.ResourcesCreatorFromTemplate
	ResourcesCountSpecEnforcer   resources_count_spec2.ResourcesCountSpecEnforcer
	AllStatefulSetsSpecEnforcer  statefulset_spec2.AllStatefulSetsSpecEnforcer
	StatefulSetsSpecsEnforcer    statefulset_spec2.StatefulSetsSpecsEnforcer

	BlockingOperation          *operation.BlockingOperation
	BlockingOperationLogger    log3.BlockingOperationLogger
	PrimaryToReplicaFailOver   failover.PrimaryToReplicaFailOver
	PrimaryDbCountSpecEnforcer statefulset2.PrimaryDbCountSpecEnforcer
	ReplicaDbCountSpecEnforcer statefulset2.ReplicaDbCountSpecEnforcer

	BaseConfigMapCountSpecEnforcer resources_count_spec2.BaseConfigMapCountSpecEnforcer
	StatefulSetCountSpecEnforcer   resources_count_spec2.StatefulSetCountSpecEnforcer
	ServicesCountSpecEnforcer      resources_count_spec2.ServicesCountSpecEnforcer
	BackUpCronJobCountSpecEnforcer resources_count_spec2.BackUpCronJobCountSpecEnforcer
}

func CreateResourcesContext(kubegres *postgresV1.Kubegres,
	ctx context.Context,
	logger logr.Logger,
	client client.Client,
	recorder record.EventRecorder) (rc *ResourcesContext, err error) {

	setReplicaFieldToZeroIfNil(kubegres)

	rc = &ResourcesContext{}

	rc.LogWrapper = log.LogWrapper{Kubegres: kubegres, Logger: logger, Recorder: recorder}
	rc.LogWrapper.Info("KUBEGRES", "name", kubegres.Name, "Status", kubegres.Status)
	//rc.LogWrapper.WithName(kubegres.Name)

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

	rc.DefaultStorageClass = defaultspec2.CreateDefaultStorageClass(rc.KubegresContext)
	if err = defaultspec2.SetDefaultForUndefinedSpecValues(rc.KubegresContext, rc.DefaultStorageClass); err != nil {
		return nil, err
	}

	rc.BlockingOperation = operation.CreateBlockingOperation(rc.KubegresContext)

	rc.BlockingOperationLogger = log3.CreateBlockingOperationLogger(rc.KubegresContext, rc.BlockingOperation)

	if rc.ResourcesStates, err = states.LoadResourcesStates(rc.KubegresContext); err != nil {
		return nil, err
	}

	rc.ResourcesStatesLogger = log2.CreateResourcesStatesLogger(rc.KubegresContext, rc.ResourcesStates)

	rc.SpecChecker = checker.CreateSpecChecker(rc.KubegresContext, rc.ResourcesStates)

	rc.CustomConfigSpecHelper = template2.CreateCustomConfigSpecHelper(rc.KubegresContext, rc.ResourcesStates)

	resourceTemplateLoader := template2.ResourceTemplateLoader{}
	rc.ResourcesCreatorFromTemplate = template2.CreateResourcesCreatorFromTemplate(rc.KubegresContext, rc.CustomConfigSpecHelper, resourceTemplateLoader)

	addResourcesCountSpecEnforcers(rc)
	addStatefulSetSpecEnforcers(rc)
	addBlockingOperationConfigs(rc)

	return rc, nil
}

func setReplicaFieldToZeroIfNil(kubegres *postgresV1.Kubegres) {
	if kubegres.Spec.Replicas != nil {
		return
	}

	replica := int32(0)
	kubegres.Spec.Replicas = &replica
}

func addResourcesCountSpecEnforcers(rc *ResourcesContext) {

	rc.PrimaryToReplicaFailOver = failover.CreatePrimaryToReplicaFailOver(rc.KubegresContext, rc.ResourcesStates, rc.BlockingOperation)
	rc.PrimaryDbCountSpecEnforcer = statefulset2.CreatePrimaryDbCountSpecEnforcer(rc.KubegresContext, rc.ResourcesStates, rc.ResourcesCreatorFromTemplate, rc.BlockingOperation, rc.PrimaryToReplicaFailOver)
	rc.ReplicaDbCountSpecEnforcer = statefulset2.CreateReplicaDbCountSpecEnforcer(rc.KubegresContext, rc.ResourcesStates, rc.ResourcesCreatorFromTemplate, rc.BlockingOperation)
	rc.StatefulSetCountSpecEnforcer = resources_count_spec2.CreateStatefulSetCountSpecEnforcer(rc.PrimaryDbCountSpecEnforcer, rc.ReplicaDbCountSpecEnforcer)

	rc.BaseConfigMapCountSpecEnforcer = resources_count_spec2.CreateBaseConfigMapCountSpecEnforcer(rc.KubegresContext, rc.ResourcesStates, rc.ResourcesCreatorFromTemplate, rc.BlockingOperation)
	rc.ServicesCountSpecEnforcer = resources_count_spec2.CreateServicesCountSpecEnforcer(rc.KubegresContext, rc.ResourcesStates, rc.ResourcesCreatorFromTemplate)
	rc.BackUpCronJobCountSpecEnforcer = resources_count_spec2.CreateBackUpCronJobCountSpecEnforcer(rc.KubegresContext, rc.ResourcesStates, rc.ResourcesCreatorFromTemplate)

	rc.ResourcesCountSpecEnforcer = resources_count_spec2.ResourcesCountSpecEnforcer{}
	rc.ResourcesCountSpecEnforcer.AddSpecEnforcer(&rc.BaseConfigMapCountSpecEnforcer)
	rc.ResourcesCountSpecEnforcer.AddSpecEnforcer(&rc.StatefulSetCountSpecEnforcer)
	rc.ResourcesCountSpecEnforcer.AddSpecEnforcer(&rc.ServicesCountSpecEnforcer)
	rc.ResourcesCountSpecEnforcer.AddSpecEnforcer(&rc.BackUpCronJobCountSpecEnforcer)
}

func addStatefulSetSpecEnforcers(rc *ResourcesContext) {
	imageSpecEnforcer := statefulset_spec2.CreateImageSpecEnforcer(rc.KubegresContext)
	portSpecEnforcer := statefulset_spec2.CreatePortSpecEnforcer(rc.KubegresContext, rc.ResourcesStates)
	storageClassSizeSpecEnforcer := statefulset_spec2.CreateStorageClassSizeSpecEnforcer(rc.KubegresContext, rc.ResourcesStates)
	customConfigSpecEnforcer := statefulset_spec2.CreateCustomConfigSpecEnforcer(rc.CustomConfigSpecHelper)
	affinitySpecEnforcer := statefulset_spec2.CreateAffinitySpecEnforcer(rc.KubegresContext)
	tolerationsSpecEnforcer := statefulset_spec2.CreateTolerationsSpecEnforcer(rc.KubegresContext)
	resourcesSpecEnforcer := statefulset_spec2.CreateResourcesSpecEnforcer(rc.KubegresContext)
	volumeSpecEnforcer := statefulset_spec2.CreateVolumeSpecEnforcer(rc.KubegresContext)
	containerSecurityContextSpecEnforcer := statefulset_spec2.CreateContainerSecurityContextSpecEnforcer(rc.KubegresContext)
	securityContextSpecEnforcer := statefulset_spec2.CreateSecurityContextSpecEnforcer(rc.KubegresContext)
	livenessProbeSpecEnforcer := statefulset_spec2.CreateLivenessProbeSpecEnforcer(rc.KubegresContext)
	readinessProbeSpecEnforcer := statefulset_spec2.CreateReadinessProbeSpecEnforcer(rc.KubegresContext)
	serviAccountNameSpecEnforcer := statefulset_spec2.CreateServiceAccountNameSpecEnforcer(rc.KubegresContext)

	rc.StatefulSetsSpecsEnforcer = statefulset_spec2.CreateStatefulSetsSpecsEnforcer(rc.KubegresContext)
	rc.StatefulSetsSpecsEnforcer.AddSpecEnforcer(&imageSpecEnforcer)
	rc.StatefulSetsSpecsEnforcer.AddSpecEnforcer(&portSpecEnforcer)
	rc.StatefulSetsSpecsEnforcer.AddSpecEnforcer(&storageClassSizeSpecEnforcer)
	rc.StatefulSetsSpecsEnforcer.AddSpecEnforcer(&customConfigSpecEnforcer)
	rc.StatefulSetsSpecsEnforcer.AddSpecEnforcer(&affinitySpecEnforcer)
	rc.StatefulSetsSpecsEnforcer.AddSpecEnforcer(&tolerationsSpecEnforcer)
	rc.StatefulSetsSpecsEnforcer.AddSpecEnforcer(&resourcesSpecEnforcer)
	rc.StatefulSetsSpecsEnforcer.AddSpecEnforcer(&volumeSpecEnforcer)
	rc.StatefulSetsSpecsEnforcer.AddSpecEnforcer(&containerSecurityContextSpecEnforcer)
	rc.StatefulSetsSpecsEnforcer.AddSpecEnforcer(&securityContextSpecEnforcer)
	rc.StatefulSetsSpecsEnforcer.AddSpecEnforcer(&livenessProbeSpecEnforcer)
	rc.StatefulSetsSpecsEnforcer.AddSpecEnforcer(&readinessProbeSpecEnforcer)
	rc.StatefulSetsSpecsEnforcer.AddSpecEnforcer(&serviAccountNameSpecEnforcer)

	rc.AllStatefulSetsSpecEnforcer = statefulset_spec2.CreateAllStatefulSetsSpecEnforcer(rc.KubegresContext, rc.ResourcesStates, rc.BlockingOperation, rc.StatefulSetsSpecsEnforcer)
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
