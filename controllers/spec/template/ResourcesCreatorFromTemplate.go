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

package template

import (
	apps "k8s.io/api/apps/v1"
	"k8s.io/api/batch/v1beta1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	postgresV1 "reactive-tech.io/kubegres/api/v1"
	"reactive-tech.io/kubegres/controllers/ctx"
)

type ResourcesCreatorFromTemplate struct {
	kubegresContext        ctx.KubegresContext
	customConfigSpecHelper CustomConfigSpecHelper
	templateFromFiles      ResourceTemplateLoader
}

const (
	KubegresInternalAnnotationKey = "kubectl.kubernetes.io/last-applied-configuration"
)

func CreateResourcesCreatorFromTemplate(kubegresContext ctx.KubegresContext,
	customConfigSpecHelper CustomConfigSpecHelper,
	resourceTemplateLoader ResourceTemplateLoader) ResourcesCreatorFromTemplate {

	return ResourcesCreatorFromTemplate{
		kubegresContext:        kubegresContext,
		customConfigSpecHelper: customConfigSpecHelper,
		templateFromFiles:      resourceTemplateLoader,
	}
}

func (r *ResourcesCreatorFromTemplate) CreateBaseConfigMap() (core.ConfigMap, error) {
	baseConfigMap, err := r.templateFromFiles.LoadBaseConfigMap()
	if err != nil {
		return core.ConfigMap{}, err
	}

	baseConfigMap.Namespace = r.kubegresContext.Kubegres.Namespace
	//baseConfigMap.OwnerReferences = r.getOwnerReference()

	return baseConfigMap, nil
}

func (r *ResourcesCreatorFromTemplate) CreatePrimaryService() (core.Service, error) {
	primaryService, err := r.templateFromFiles.LoadPrimaryService()
	if err != nil {
		return core.Service{}, err
	}

	r.initService(&primaryService)

	primaryService.Name = r.kubegresContext.GetServiceResourceName(true)

	return primaryService, nil
}

func (r *ResourcesCreatorFromTemplate) CreateReplicaService() (core.Service, error) {
	replicaService, err := r.templateFromFiles.LoadReplicaService()
	if err != nil {
		return core.Service{}, err
	}

	r.initService(&replicaService)

	replicaService.Name = r.kubegresContext.GetServiceResourceName(false)

	return replicaService, nil
}

func (r *ResourcesCreatorFromTemplate) CreatePrimaryStatefulSet(instance string) (apps.StatefulSet, error) {
	statefulSetTemplate, err := r.templateFromFiles.LoadPrimaryStatefulSet()
	if err != nil {
		return apps.StatefulSet{}, err
	}

	primaryServiceName := r.kubegresContext.GetServiceResourceName(true)
	r.initStatefulSet(primaryServiceName, instance, &statefulSetTemplate)
	r.customConfigSpecHelper.ConfigureStatefulSet(&statefulSetTemplate)
	return statefulSetTemplate, nil
}

func (r *ResourcesCreatorFromTemplate) CreateReplicaStatefulSet(instance string) (apps.StatefulSet, error) {
	statefulSetTemplate, err := r.templateFromFiles.LoadReplicaStatefulSet()
	if err != nil {
		return apps.StatefulSet{}, err
	}

	primaryServiceName := r.kubegresContext.GetServiceResourceName(true)
	replicaServiceName := r.kubegresContext.GetServiceResourceName(false)

	r.initStatefulSet(replicaServiceName, instance, &statefulSetTemplate)
	r.customConfigSpecHelper.ConfigureStatefulSet(&statefulSetTemplate)

	initContainer := &statefulSetTemplate.Spec.Template.Spec.InitContainers[0]
	postgresSpec := r.kubegresContext.Kubegres.Spec
	initContainer.Image = postgresSpec.Image
	initContainer.Env[0].Value = primaryServiceName
	initContainer.Env[1].ValueFrom = r.getEnvVar(ctx.EnvVarNameOfPostgresReplicationUserPsw).ValueFrom
	initContainer.Env[2].Value = postgresSpec.Database.VolumeMount + "/" + ctx.DefaultDatabaseFolder
	initContainer.VolumeMounts[0].MountPath = postgresSpec.Database.VolumeMount

	return statefulSetTemplate, nil
}

func (r *ResourcesCreatorFromTemplate) CreateBackUpCronJob(configMapNameForBackUp string) (v1beta1.CronJob, error) {
	backUpCronJob, err := r.templateFromFiles.LoadBackUpCronJob()
	if err != nil {
		return v1beta1.CronJob{}, err
	}

	postgres := r.kubegresContext.Kubegres
	backupSpec := postgres.Spec.Backup
	backUpName := ctx.CronJobNamePrefix + postgres.Name

	backUpCronJob.Name = backUpName
	backUpCronJob.Namespace = postgres.Namespace
	backUpCronJob.OwnerReferences = r.getOwnerReference()

	backUpCronJob.Spec.Schedule = backupSpec.Schedule

	backUpCronJob.Spec.JobTemplate.Spec.Template.Annotations = r.getCustomAnnotations()

	backUpCronJobSpec := &backUpCronJob.Spec.JobTemplate.Spec.Template.Spec

	backUpCronJobSpec.Volumes[0].PersistentVolumeClaim.ClaimName = backupSpec.PvcName
	backUpCronJobSpec.Volumes[1].ConfigMap.Name = configMapNameForBackUp

	backUpCronJobContainer := &backUpCronJobSpec.Containers[0]
	backUpCronJobContainer.Image = postgres.Spec.Image
	backUpCronJobContainer.VolumeMounts[0].MountPath = backupSpec.VolumeMount
	backUpCronJobContainer.Env[0].ValueFrom = r.getEnvVar(ctx.EnvVarNameOfPostgresSuperUserPsw).ValueFrom
	backUpCronJobContainer.Env[1].Value = postgres.Name
	backUpCronJobContainer.Env[2].Value = backupSpec.VolumeMount
	backUpCronJobContainer.Env = append(backUpCronJobContainer.Env, r.kubegresContext.Kubegres.Spec.Env...)

	backSourceDbHostName := r.kubegresContext.GetServiceResourceName(false)
	if r.kubegresContext.ReplicasCount() == 1 {
		backSourceDbHostName = r.kubegresContext.GetServiceResourceName(true)
	}
	backUpCronJobContainer.Env[3].Value = backSourceDbHostName

	return backUpCronJob, nil
}

func (r *ResourcesCreatorFromTemplate) initService(service *core.Service) {
	resourceName := r.kubegresContext.Kubegres.Name
	service.Namespace = r.kubegresContext.Kubegres.Namespace
	service.OwnerReferences = r.getOwnerReference()
	service.Labels[ctx.NameLabelKey] = resourceName
	service.Spec.Selector[ctx.NameLabelKey] = resourceName
	service.Spec.Ports[0].Port = r.kubegresContext.Kubegres.Spec.Port
}

func (r *ResourcesCreatorFromTemplate) initStatefulSet(serviceName string, instance string,
	statefulSetTemplate *apps.StatefulSet) {
	resourceName := r.kubegresContext.Kubegres.Name
	resourceSpec := r.kubegresContext.Kubegres.Spec
	statefulSetResourceName := r.kubegresContext.GetStatefulSetResourceName(instance)

	statefulSetTemplate.Name = statefulSetResourceName
	statefulSetTemplate.Namespace = r.kubegresContext.Kubegres.Namespace
	statefulSetTemplate.Annotations = r.getCustomAnnotations()
	statefulSetTemplate.Labels[ctx.NameLabelKey] = resourceName
	statefulSetTemplate.Labels[ctx.InstanceLabelKey] = instance
	statefulSetTemplate.OwnerReferences = r.getOwnerReference()

	statefulSetTemplate.Spec.ServiceName = serviceName
	statefulSetTemplate.Spec.Selector.MatchLabels[ctx.NameLabelKey] = resourceName
	statefulSetTemplate.Spec.Selector.MatchLabels[ctx.InstanceLabelKey] = instance
	statefulSetTemplate.Spec.Template.Labels[ctx.NameLabelKey] = resourceName
	statefulSetTemplate.Spec.Template.Labels[ctx.InstanceLabelKey] = instance
	statefulSetTemplate.Spec.Template.Annotations = r.getCustomAnnotations()

	statefulSetTemplateSpec := &statefulSetTemplate.Spec.Template.Spec

	if resourceSpec.ImagePullSecrets != nil {
		statefulSetTemplateSpec.ImagePullSecrets = append(statefulSetTemplateSpec.ImagePullSecrets, resourceSpec.ImagePullSecrets...)
	}

	container := &statefulSetTemplateSpec.Containers[0]
	container.Name = statefulSetResourceName
	container.Image = resourceSpec.Image
	container.Ports[0].ContainerPort = resourceSpec.Port
	container.VolumeMounts[0].MountPath = resourceSpec.Database.VolumeMount
	container.Env = append(container.Env, core.EnvVar{Name: ctx.EnvVarNamePgData, Value: resourceSpec.Database.VolumeMount + "/" + ctx.DefaultDatabaseFolder})
	container.Env = append(container.Env, r.kubegresContext.Kubegres.Spec.Env...)

	statefulSetTemplate.Spec.VolumeClaimTemplates[0].Spec.StorageClassName = resourceSpec.Database.StorageClassName
	statefulSetTemplate.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests = core.ResourceList{core.ResourceStorage: resource.MustParse(resourceSpec.Database.Size)}

	nodeSpec := r.kubegresContext.GetNodeSetSpecFromInstance(instance)
	if nodeSpec.Affinity != nil {
		statefulSetTemplateSpec.Affinity = nodeSpec.Affinity
	} else if resourceSpec.Scheduler.Affinity != nil {
		statefulSetTemplateSpec.Affinity = resourceSpec.Scheduler.Affinity
	}
	if len(nodeSpec.Tolerations) > 0 {
		statefulSetTemplateSpec.Tolerations = nodeSpec.Tolerations
	} else if len(resourceSpec.Scheduler.Tolerations) > 0 {
		statefulSetTemplateSpec.Tolerations = resourceSpec.Scheduler.Tolerations
	}

	if resourceSpec.Resources.Requests != nil || resourceSpec.Resources.Limits != nil {
		statefulSetTemplate.Spec.Template.Spec.Containers[0].Resources = resourceSpec.Resources
	}

	if resourceSpec.Volume.VolumeClaimTemplates != nil {

		for _, volumeClaimTemplate := range resourceSpec.Volume.VolumeClaimTemplates {
			persistentVolumeClaim := core.PersistentVolumeClaim{}
			persistentVolumeClaim.Name = volumeClaimTemplate.Name
			persistentVolumeClaim.Namespace = r.kubegresContext.Kubegres.Namespace
			persistentVolumeClaim.Spec = volumeClaimTemplate.Spec
			statefulSetTemplate.Spec.VolumeClaimTemplates = append(statefulSetTemplate.Spec.VolumeClaimTemplates, persistentVolumeClaim)
		}
	}

	if resourceSpec.Volume.Volumes != nil {
		statefulSetTemplate.Spec.Template.Spec.Volumes = append(statefulSetTemplate.Spec.Template.Spec.Volumes, r.kubegresContext.Kubegres.Spec.Volume.Volumes...)
	}

	if resourceSpec.Volume.VolumeMounts != nil {
		statefulSetTemplate.Spec.Template.Spec.Containers[0].VolumeMounts = append(statefulSetTemplate.Spec.Template.Spec.Containers[0].VolumeMounts, r.kubegresContext.Kubegres.Spec.Volume.VolumeMounts...)
	}

	if resourceSpec.SecurityContext != nil {
		statefulSetTemplate.Spec.Template.Spec.SecurityContext = resourceSpec.SecurityContext
	}

	if resourceSpec.Probe.LivenessProbe != nil {
		statefulSetTemplate.Spec.Template.Spec.Containers[0].LivenessProbe = resourceSpec.Probe.LivenessProbe
	}

	if resourceSpec.Probe.ReadinessProbe != nil {
		statefulSetTemplate.Spec.Template.Spec.Containers[0].ReadinessProbe = resourceSpec.Probe.ReadinessProbe
	}
}

// Extract annotations set in Kubegres YAML by
// excluding the internal annotation "kubectl.kubernetes.io/last-applied-configuration"
func (r *ResourcesCreatorFromTemplate) getCustomAnnotations() map[string]string {

	var customSpecAnnotations = make(map[string]string)

	for key, value := range r.kubegresContext.Kubegres.ObjectMeta.Annotations {
		if key == KubegresInternalAnnotationKey {
			continue
		}
		customSpecAnnotations[key] = value
	}

	return customSpecAnnotations
}

func (r *ResourcesCreatorFromTemplate) getOwnerReference() []metav1.OwnerReference {
	return []metav1.OwnerReference{*metav1.NewControllerRef(r.kubegresContext.Kubegres, postgresV1.GroupVersion.WithKind(ctx.KindKubegres))}
}

func (r *ResourcesCreatorFromTemplate) getEnvVar(envName string) core.EnvVar {
	for _, envVar := range r.kubegresContext.Kubegres.Spec.Env {
		if envVar.Name == envName {
			return envVar
		}
	}
	return core.EnvVar{}
}

func (r *ResourcesCreatorFromTemplate) doesCustomConfigExist() bool {
	return r.kubegresContext.Kubegres.Spec.CustomConfig != "" &&
		r.kubegresContext.Kubegres.Spec.CustomConfig != ctx.BaseConfigMapName
}
