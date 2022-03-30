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

package checker

import (
	"errors"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	postgresV1 "reactive-tech.io/kubegres/api/v1"
	"reactive-tech.io/kubegres/controllers/ctx"
	"reactive-tech.io/kubegres/controllers/states"
	"reactive-tech.io/kubegres/controllers/states/statefulset"
	"reflect"
)

type SpecChecker struct {
	kubegresContext ctx.KubegresContext
	resourcesStates states.ResourcesStates
}

type SpecCheckResult struct {
	HasSpecFatalError bool
	FatalErrorMessage string
}

func CreateSpecChecker(kubegresContext ctx.KubegresContext, resourcesStates states.ResourcesStates) SpecChecker {
	return SpecChecker{kubegresContext: kubegresContext, resourcesStates: resourcesStates}
}

func (r *SpecChecker) CheckSpec() (SpecCheckResult, error) {

	specCheckResult := SpecCheckResult{}

	spec := &r.kubegresContext.Kubegres.Spec
	primaryStatefulSet := r.getPrimaryStatefulSet()
	primaryStatefulSetSpec := primaryStatefulSet.StatefulSet.Spec
	const emptyStr = ""

	if primaryStatefulSet.Pod.IsReady {

		primaryVolumeMount := primaryStatefulSetSpec.Template.Spec.Containers[0].VolumeMounts[0].MountPath
		if spec.Database.VolumeMount != primaryVolumeMount {

			specCheckResult.HasSpecFatalError = true
			specCheckResult.FatalErrorMessage = r.createErrMsgSpecCannotBeChanged("spec.database.volumeMount",
				primaryVolumeMount, spec.Database.VolumeMount,
				"Otherwise, the cluster of PostgreSql servers risk of being inconsistent.")

			r.kubegresContext.Kubegres.Spec.Database.VolumeMount = primaryVolumeMount
			r.updateKubegresSpec("spec.database.volumeMount", primaryVolumeMount)
		}

		primaryStorageClassName := primaryStatefulSetSpec.VolumeClaimTemplates[0].Spec.StorageClassName
		if *spec.Database.StorageClassName != *primaryStorageClassName {

			specCheckResult.HasSpecFatalError = true
			specCheckResult.FatalErrorMessage = r.createErrMsgSpecCannotBeChanged("spec.database.storageClassName",
				*primaryStorageClassName,
				*spec.Database.StorageClassName,
				"Otherwise, the cluster of PostgreSql servers risk of being inconsistent.")

			r.kubegresContext.Kubegres.Spec.Database.StorageClassName = primaryStorageClassName
			r.updateKubegresSpec("spec.database.storageClassName", *primaryStorageClassName)
		}

		primaryStorageSizeQuantity := primaryStatefulSetSpec.VolumeClaimTemplates[0].Spec.Resources.Requests[v1.ResourceStorage]
		primaryStorageSize := primaryStorageSizeQuantity.String()
		if spec.Database.Size != primaryStorageSize && !r.doesStorageClassAllowVolumeExpansion() {

			specCheckResult.HasSpecFatalError = true
			specCheckResult.FatalErrorMessage = r.createErrMsgSpecCannotBeChanged("spec.database.size",
				primaryStorageSize,
				spec.Database.Size,
				"The StorageClass does not allow volume expansion. The option AllowVolumeExpansion is set to false.")

			spec.Database.Size = primaryStorageSize
			r.updateKubegresSpec("spec.database.size", primaryStorageSize)

			// TODO: condition to remove when Kubernetes allows updating storage size in StatefulSet (see https://github.com/kubernetes/enhancements/pull/2842)
		} else if spec.Database.Size != primaryStorageSize {
			specCheckResult.HasSpecFatalError = true
			specCheckResult.FatalErrorMessage = r.createErrMsgSpecCannotBeChanged("spec.database.size",
				primaryStorageSize,
				spec.Database.Size,
				"The database size cannot be modified after the creation of the Postgres cluster. "+
					"We are rolling back that value to its previous value.")

			spec.Database.Size = primaryStorageSize
			r.updateKubegresSpec("spec.database.size", primaryStorageSize)
		}

		if r.hasCustomVolumeClaimTemplatesChanged(primaryStatefulSetSpec) {
			specCheckResult.HasSpecFatalError = true
			specCheckResult.FatalErrorMessage = r.logSpecErrMsg("In the Resources Spec, the array 'spec.Volume.VolumeClaimTemplates' " +
				"has changed. Kubernetes does not allow to update that field in StatefulSet specification. Please rollback your changes in the YAML.")
		}

		if specCheckResult.HasSpecFatalError {
			return specCheckResult, nil
		}
	}

	if !r.dbStorageClassDeployed() {
		specCheckResult.HasSpecFatalError = true
		specCheckResult.FatalErrorMessage = r.logSpecErrMsg("In the Resources Spec the value of " +
			"'spec.database.storageClassName' has a StorageClass name which is not deployed. Please deploy this StorageClass, " +
			"otherwise this operator cannot work correctly.")
	}

	if spec.Database.Size == emptyStr {
		specCheckResult.HasSpecFatalError = true
		specCheckResult.FatalErrorMessage = r.createErrMsgSpecUndefined("spec.database.size")
	}

	if r.isCustomConfigNotDeployed(spec) {
		specCheckResult.HasSpecFatalError = true
		specCheckResult.FatalErrorMessage = r.logSpecErrMsg("In the Resources Spec the value of " +
			"'spec.customConfig' has a configMap name which is not deployed. Please deploy this configMap otherwise this " +
			"operator cannot work correctly.")
	}

	if !r.doesEnvVarExist(ctx.EnvVarNameOfPostgresSuperUserPsw) {
		specCheckResult.HasSpecFatalError = true
		specCheckResult.FatalErrorMessage = r.createErrMsgSpecUndefined("spec.env.POSTGRES_PASSWORD")
	}

	if !r.doesEnvVarExist(ctx.EnvVarNameOfPostgresReplicationUserPsw) {
		specCheckResult.HasSpecFatalError = true
		specCheckResult.FatalErrorMessage = r.createErrMsgSpecUndefined("spec.env.POSTGRES_REPLICATION_PASSWORD")
	}

	if *spec.Replicas <= 0 && len(spec.NodeSets) == 0 {
		specCheckResult.HasSpecFatalError = true
		specCheckResult.FatalErrorMessage = r.createErrMsgSpecUndefined("spec.replicas")
	}

	if *spec.Replicas > 0 && len(spec.NodeSets) > 0 {
		specCheckResult.HasSpecFatalError = true
		specCheckResult.FatalErrorMessage = r.logSpecErrMsg("In the Resources Spec the value of " +
			"'spec.replicas' and 'spec.nodeSets' are mutually exclusive. " +
			"Please set only one of the value otherwise this operator cannot work correctly.")
	}

	for _, nodeSet := range spec.NodeSets {
		if nodeSet.Name == "" {
			specCheckResult.HasSpecFatalError = true
			specCheckResult.FatalErrorMessage = r.createErrMsgSpecUndefined("spec.nodeSets[].Name")
		}
	}

	if spec.Image == "" {
		specCheckResult.HasSpecFatalError = true
		specCheckResult.FatalErrorMessage = r.createErrMsgSpecUndefined("spec.image")
	}

	if r.isBackUpConfigured(spec) {

		if spec.Backup.VolumeMount == emptyStr {
			specCheckResult.HasSpecFatalError = true
			specCheckResult.FatalErrorMessage = r.createErrMsgSpecUndefined("spec.Backup.VolumeMount")
		}

		if spec.Backup.PvcName == emptyStr {
			specCheckResult.HasSpecFatalError = true
			specCheckResult.FatalErrorMessage = r.createErrMsgSpecUndefined("spec.Backup.PvcName")
		}

		if spec.Backup.PvcName != emptyStr && !r.isBackUpPvcDeployed() {
			specCheckResult.HasSpecFatalError = true
			specCheckResult.FatalErrorMessage = r.logSpecErrMsg("In the Resources Spec the value of " +
				"'spec.Backup.PvcName' has a PersistentVolumeClaim name which is not deployed. Please deploy this " +
				"PersistentVolumeClaim, otherwise this operator cannot work correctly.")
		}
	}

	reservedVolumeName := r.doCustomVolumeClaimTemplatesHaveReservedName()
	if reservedVolumeName != "" {
		specCheckResult.HasSpecFatalError = true
		specCheckResult.FatalErrorMessage = r.logSpecErrMsg("In the Resources Spec the value of " +
			"'spec.Volume.VolumeClaimTemplates' has an entry with a volume name which is a reserved name: " + reservedVolumeName + " . " +
			"That name cannot be used and it is reserved for Kubegres internal usages. Please change that name in the YAML.")
	}

	reservedVolumeName = r.doCustomVolumesHaveReservedName()
	if reservedVolumeName != "" {
		specCheckResult.HasSpecFatalError = true
		specCheckResult.FatalErrorMessage = r.logSpecErrMsg("In the Resources Spec the value of " +
			"'spec.Volume.Volumes' has an entry with a volume name which is a reserved name: " + reservedVolumeName + " . " +
			"That name cannot be used and it is reserved for Kubegres internal usages. Please change that name in the YAML.")
	}

	reservedVolumeName = r.doCustomVolumeMountsHaveReservedName()
	if reservedVolumeName != "" {
		specCheckResult.HasSpecFatalError = true
		specCheckResult.FatalErrorMessage = r.logSpecErrMsg("In the Resources Spec the value of " +
			"'spec.Volume.VolumeMounts' has an entry with a volume name which is a reserved name: " + reservedVolumeName + " . " +
			"That name cannot be used and it is reserved for Kubegres internal usages. Please change that name in the YAML.")
	}

	if r.doCustomVolumeMountsHaveReservedPath() {
		specCheckResult.HasSpecFatalError = true
		specCheckResult.FatalErrorMessage = r.logSpecErrMsg("In the Resources Spec the value of 'spec.Volume.VolumeMounts' " +
			"has an entry with a 'mountPath' value which is reserved for the Postgres database: " + r.kubegresContext.Kubegres.Spec.Database.VolumeMount + " . " +
			"That value cannot be used and it is reserved for Kubegres internal usages. Please change that value in the YAML.")
	}

	return specCheckResult, nil
}

func (r *SpecChecker) updateKubegresSpec(specName string, specValue string) {
	err := r.kubegresContext.Client.Update(r.kubegresContext.Ctx, r.kubegresContext.Kubegres)
	if err != nil {
		r.kubegresContext.Log.Error(err, "Unable to rollback the value of '"+specName+"' to '"+specValue+"'")
	}
}

func (r *SpecChecker) isBackUpConfigured(spec *postgresV1.KubegresSpec) bool {
	return spec.Backup.Schedule != ""
}

func (r *SpecChecker) dbStorageClassDeployed() bool {
	return r.resourcesStates.DbStorageClass.IsDeployed
}

func (r *SpecChecker) isBackUpPvcDeployed() bool {
	return r.resourcesStates.BackUp.IsPvcDeployed
}

func (r *SpecChecker) isCustomConfigNotDeployed(spec *postgresV1.KubegresSpec) bool {
	return spec.CustomConfig != "" &&
		spec.CustomConfig != ctx.BaseConfigMapName &&
		!r.resourcesStates.Config.IsCustomConfigDeployed
}

func (r *SpecChecker) createErrMsgSpecUndefined(specName string) string {
	errorMsg := "In the Resources Spec the value of '" + specName + "' is undefined. Please set a value otherwise this operator cannot work correctly."
	return r.logSpecErrMsg(errorMsg)
}

func (r *SpecChecker) createErrMsgSpecCannotBeChanged(specName, currentValue, newValue, reason string) string {
	errorMsg := "In the Resources Spec the value of '" + specName + "' cannot be changed from '" + currentValue + "' to '" + newValue + "' after Pods were created. " +
		reason + " " +
		"We roll-backed Kubegres spec to the currently working value '" + currentValue + "'. " +
		"If you know what you are doing, you can manually update that spec in every StatefulSet of your PostgreSql cluster and then Kubegres will automatically update itself."
	return r.logSpecErrMsg(errorMsg)
}

func (r *SpecChecker) logSpecErrMsg(errorMsg string) string {
	r.kubegresContext.Log.ErrorEvent("SpecCheckErr", errors.New(errorMsg), "")
	return errorMsg
}

func (r *SpecChecker) doesEnvVarExist(envName string) bool {
	for _, envVar := range r.kubegresContext.Kubegres.Spec.Env {
		if envVar.Name == envName {
			return true
		}
	}
	return false
}

func (r *SpecChecker) getPrimaryStatefulSet() statefulset.StatefulSetWrapper {
	return r.resourcesStates.StatefulSets.Primary
}

func (r *SpecChecker) doesStorageClassAllowVolumeExpansion() bool {
	storageClass, err := r.resourcesStates.DbStorageClass.GetStorageClass()
	if err != nil {
		return false
	}
	return storageClass.AllowVolumeExpansion != nil && *storageClass.AllowVolumeExpansion
}

func (r *SpecChecker) doCustomVolumeClaimTemplatesHaveReservedName() string {
	for _, customVolumeClaimTemplate := range r.kubegresContext.Kubegres.Spec.Volume.VolumeClaimTemplates {
		if r.kubegresContext.IsReservedVolumeName(customVolumeClaimTemplate.Name) {
			return customVolumeClaimTemplate.Name
		}
	}
	return ""
}

func (r *SpecChecker) doCustomVolumesHaveReservedName() string {
	for _, customVolume := range r.kubegresContext.Kubegres.Spec.Volume.Volumes {
		if r.kubegresContext.IsReservedVolumeName(customVolume.Name) {
			return customVolume.Name
		}
	}
	return ""
}

func (r *SpecChecker) doCustomVolumeMountsHaveReservedName() string {
	for _, customVolumeMount := range r.kubegresContext.Kubegres.Spec.Volume.VolumeMounts {
		if r.kubegresContext.IsReservedVolumeName(customVolumeMount.Name) {
			return customVolumeMount.Name
		}
	}
	return ""
}

func (r *SpecChecker) doCustomVolumeMountsHaveReservedPath() bool {
	for _, customVolumeMount := range r.kubegresContext.Kubegres.Spec.Volume.VolumeMounts {
		if customVolumeMount.MountPath == r.kubegresContext.Kubegres.Spec.Database.VolumeMount {
			return true
		}
	}
	return false
}

func (r *SpecChecker) hasCustomVolumeClaimTemplatesChanged(primaryStatefulSetSpec apps.StatefulSetSpec) bool {

	customVolumeClaimTemplatesCount := 0
	for _, currentVolumeClaimTemplate := range primaryStatefulSetSpec.VolumeClaimTemplates {

		if !r.kubegresContext.IsReservedVolumeName(currentVolumeClaimTemplate.Name) {
			customVolumeClaimTemplatesCount++
			if !r.doesCurrentVolumeClaimTemplateExistInExpectedSpec(currentVolumeClaimTemplate) {
				return true
			}
		}
	}

	if customVolumeClaimTemplatesCount != len(r.kubegresContext.Kubegres.Spec.Volume.VolumeClaimTemplates) {
		return true
	}

	return false
}

func (r *SpecChecker) doesCurrentVolumeClaimTemplateExistInExpectedSpec(currentCustomVolumeClaimTemplate v1.PersistentVolumeClaim) bool {

	for _, expectedCustomVolumeClaimTemplate := range r.kubegresContext.Kubegres.Spec.Volume.VolumeClaimTemplates {

		if expectedCustomVolumeClaimTemplate.Name == currentCustomVolumeClaimTemplate.Name &&
			reflect.DeepEqual(expectedCustomVolumeClaimTemplate.Spec.StorageClassName, currentCustomVolumeClaimTemplate.Spec.StorageClassName) &&
			reflect.DeepEqual(expectedCustomVolumeClaimTemplate.Spec.Resources, currentCustomVolumeClaimTemplate.Spec.Resources) {
			return true
		}
	}
	return false
}
