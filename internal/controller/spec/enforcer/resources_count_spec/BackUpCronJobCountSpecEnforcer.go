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

package resources_count_spec

import (
	batch "k8s.io/api/batch/v1"
	"reactive-tech.io/kubegres/internal/controller/ctx"
	"reactive-tech.io/kubegres/internal/controller/spec/template"
	states2 "reactive-tech.io/kubegres/internal/controller/states"
)

type BackUpCronJobCountSpecEnforcer struct {
	kubegresContext  ctx.KubegresContext
	resourcesStates  states2.ResourcesStates
	resourcesCreator template.ResourcesCreatorFromTemplate
}

func CreateBackUpCronJobCountSpecEnforcer(kubegresContext ctx.KubegresContext,
	resourcesStates states2.ResourcesStates,
	resourcesCreator template.ResourcesCreatorFromTemplate) BackUpCronJobCountSpecEnforcer {

	return BackUpCronJobCountSpecEnforcer{
		kubegresContext:  kubegresContext,
		resourcesStates:  resourcesStates,
		resourcesCreator: resourcesCreator,
	}
}

func (r *BackUpCronJobCountSpecEnforcer) EnforceSpec() error {

	if r.isCronJobDeployed() {

		if !r.hasSpecChanged() {
			return nil
		}

		err := r.deleteCronJob()
		if err != nil {
			return err
		}
	}

	if !r.isBackUpConfigured() {
		return nil
	}

	configMapNameForBackUp := r.getConfigMapNameForBackUp(r.resourcesStates.Config)
	cronJob, err := r.resourcesCreator.CreateBackUpCronJob(configMapNameForBackUp)
	if err != nil {
		r.kubegresContext.Log.ErrorEvent("BackUpCronJobTemplateErr", err, "Unable to create a BackUp CronJob object from template.")
		return err
	}

	return r.deployCronJob(cronJob)
}

func (r *BackUpCronJobCountSpecEnforcer) getConfigMapNameForBackUp(configStates states2.ConfigStates) string {
	if configStates.ConfigLocations.BackUpScript == ctx.BaseConfigMapVolumeName {
		return configStates.BaseConfigName
	}
	return configStates.CustomConfigName
}

func (r *BackUpCronJobCountSpecEnforcer) deployCronJob(cronJob batch.CronJob) error {

	r.kubegresContext.Log.Info("Deploying BackUp CronJob.", "CronJob name", cronJob.Name)

	if err := r.kubegresContext.Client.Create(r.kubegresContext.Ctx, &cronJob); err != nil {
		r.kubegresContext.Log.ErrorEvent("BackUpCronJobDeploymentErr", err, "Unable to deploy BackUp CronJob.", "CronJob name", cronJob.Name)
		return err
	}

	r.kubegresContext.Log.InfoEvent("BackUpCronJobDeployment", "Deployed BackUp CronJob.", "CronJob name", cronJob.Name)
	return nil
}

func (r *BackUpCronJobCountSpecEnforcer) isBackUpConfigured() bool {
	return r.kubegresContext.Kubegres.Spec.Backup.Schedule != ""
}

func (r *BackUpCronJobCountSpecEnforcer) isCronJobDeployed() bool {
	return r.resourcesStates.BackUp.IsCronJobDeployed
}

func (r *BackUpCronJobCountSpecEnforcer) hasSpecChanged() (hasSpecChanged bool) {

	cronJob := r.resourcesStates.BackUp.DeployedCronJob
	cronJobSpec := &cronJob.Spec
	cronJobTemplateSpec := cronJob.Spec.JobTemplate.Spec.Template.Spec
	kubegresBackUpSpec := r.kubegresContext.Kubegres.Spec.Backup

	currentSchedule := cronJobSpec.Schedule
	expectedSchedule := kubegresBackUpSpec.Schedule
	if currentSchedule != expectedSchedule {
		hasSpecChanged = true
		r.logSpecChange("spec.backup.schedule")
	}

	currentVolumeMount := cronJobTemplateSpec.Containers[0].VolumeMounts[0].MountPath
	expectedVolumeMount := kubegresBackUpSpec.VolumeMount
	if currentVolumeMount != expectedVolumeMount {
		hasSpecChanged = true
		r.logSpecChange("spec.backup.volumeMount")
	}

	currentPvcName := cronJobTemplateSpec.Volumes[0].PersistentVolumeClaim.ClaimName
	expectedPvcName := kubegresBackUpSpec.PvcName
	if currentPvcName != expectedPvcName {
		hasSpecChanged = true
		r.logSpecChange("spec.backup.pvcName")
	}

	currentCustomConfig := cronJobTemplateSpec.Volumes[1].ConfigMap.Name
	expectedCustomConfig := r.getConfigMapNameForBackUp(r.resourcesStates.Config)
	if currentCustomConfig != expectedCustomConfig {
		hasSpecChanged = true
		r.logSpecChange("spec.backup.customConfig")
	}

	return hasSpecChanged
}

func (r *BackUpCronJobCountSpecEnforcer) deleteCronJob() error {

	err := r.kubegresContext.Client.Delete(r.kubegresContext.Ctx, r.resourcesStates.BackUp.DeployedCronJob)
	if err != nil {
		r.kubegresContext.Log.ErrorEvent("PodSpecEnforcementErr", err,
			"Unable to delete a Backup CronJob which had its spec changed. "+
				"The aim of the deletion is to trigger a re-creation of the CronJob.",
			"CronJob name:", r.resourcesStates.BackUp.DeployedCronJob.Name)
		return err
	}

	r.kubegresContext.Log.InfoEvent("BackUpCronJobDeletion", "Deleted BackUp CronJob.", "CronJob name", r.resourcesStates.BackUp.DeployedCronJob.Name)
	return nil
}

func (r *BackUpCronJobCountSpecEnforcer) logSpecChange(specName string) {
	r.kubegresContext.Log.Info("BackUp spec '"+specName+"' has changed. "+
		"We will delete Backup CronJob resource so that it gets re-created by Kubegres "+
		"and the spec change will be applied on creation.",
		"CronJob name:", r.resourcesStates.BackUp.DeployedCronJob.Name)
}
