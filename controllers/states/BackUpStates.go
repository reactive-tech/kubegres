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

package states

import (
	"k8s.io/api/batch/v1beta1"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"reactive-tech.io/kubegres/controllers/ctx"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BackUpStates struct {
	IsCronJobDeployed       bool
	IsPvcDeployed           bool
	CronJobLastScheduleTime string
	DeployedCronJob         *v1beta1.CronJob

	kubegresContext ctx.KubegresContext
}

func loadBackUpStates(kubegresContext ctx.KubegresContext) (BackUpStates, error) {
	backUpStates := BackUpStates{kubegresContext: kubegresContext}
	err := backUpStates.loadStates()
	return backUpStates, err
}

func (r *BackUpStates) loadStates() (err error) {

	r.DeployedCronJob, err = r.getDeployedCronJob()
	if err != nil {
		return err
	}

	if r.DeployedCronJob.Name != "" {
		r.IsCronJobDeployed = true
		if r.DeployedCronJob.Status.LastScheduleTime != nil {
			r.CronJobLastScheduleTime = r.DeployedCronJob.Status.LastScheduleTime.String()
		}
	}

	backUpPvc, err := r.getDeployedPvc()
	if err != nil {
		return err
	}

	if backUpPvc.Name != "" {
		r.IsPvcDeployed = true
	}

	return nil
}

func (r *BackUpStates) getDeployedCronJob() (*v1beta1.CronJob, error) {

	namespace := r.kubegresContext.Kubegres.Namespace
	resourceName := ctx.CronJobNamePrefix + r.kubegresContext.Kubegres.Name
	resourceKey := client.ObjectKey{Namespace: namespace, Name: resourceName}
	cronJob := &v1beta1.CronJob{}

	err := r.kubegresContext.Client.Get(r.kubegresContext.Ctx, resourceKey, cronJob)

	if err != nil {
		if apierrors.IsNotFound(err) {
			err = nil
		} else {
			r.kubegresContext.Log.ErrorEvent("BackUpCronJobLoadingErr", err, "Unable to load any deployed BackUp CronJob.", "CronJob name", resourceName)
		}
	}

	return cronJob, err
}

func (r *BackUpStates) getDeployedPvc() (*v1.PersistentVolumeClaim, error) {

	namespace := r.kubegresContext.Kubegres.Namespace
	resourceName := r.kubegresContext.Kubegres.Spec.Backup.PvcName
	resourceKey := client.ObjectKey{Namespace: namespace, Name: resourceName}
	pvc := &v1.PersistentVolumeClaim{}

	err := r.kubegresContext.Client.Get(r.kubegresContext.Ctx, resourceKey, pvc)

	if err != nil {
		if apierrors.IsNotFound(err) {
			err = nil
		} else {
			r.kubegresContext.Log.ErrorEvent("BackUpPersistentVolumeClaimLoadingErr", err, "Unable to load any deployed BackUp PersistentVolumeClaim.", "PersistentVolumeClaim name", resourceName)
		}
	}

	return pvc, err
}
