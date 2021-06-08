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

package corrector

import (
	"reactive-tech.io/kubegres/controllers/ctx"
	"strconv"
)

type specCorrector struct {
	kubegresContext     ctx.KubegresContext
	defaultStorageClass DefaultStorageClass
}

func CorrectSpec(kubegresContext ctx.KubegresContext, defaultStorageClass DefaultStorageClass) error {
	specCorrector := specCorrector{kubegresContext: kubegresContext, defaultStorageClass: defaultStorageClass}
	return specCorrector.correctSpec()
}

func (r *specCorrector) correctSpec() error {

	wasSpecCorrected := false
	spec := &r.kubegresContext.Kubegres.Spec
	const emptyStr = ""

	if spec.Port <= 0 {
		wasSpecCorrected = true
		spec.Port = 5432
		r.createSpecCorrectionLog("spec.port", strconv.Itoa(int(spec.Port)))
	}

	if spec.Database.VolumeMount == emptyStr {
		wasSpecCorrected = true
		spec.Database.VolumeMount = ctx.DefaultContainerVolumeMount
		r.createSpecCorrectionLog("spec.database.volumeMount", spec.Database.VolumeMount)
	}

	if spec.CustomConfig == emptyStr {
		wasSpecCorrected = true
		spec.CustomConfig = ctx.BaseConfigMapName
		r.createSpecCorrectionLog("spec.customConfig", spec.CustomConfig)
	}

	if r.isStorageClassNameUndefinedInSpec() {

		wasSpecCorrected = true
		defaultStorageClassName, err := r.defaultStorageClass.GetDefaultStorageClassName()
		if err != nil {
			return err
		}

		spec.Database.StorageClassName = &defaultStorageClassName
		r.createSpecCorrectionLog("spec.Database.StorageClassName", defaultStorageClassName)
	}

	if wasSpecCorrected {
		return r.updateSpec()
	}

	return nil
}

func (r *specCorrector) createSpecCorrectionLog(specName string, specValue string) {
	r.kubegresContext.Log.InfoEvent("SpecCheckCorrection", "Corrected an undefined value in Spec.", specName, "New value: "+specValue+"")
}

func (r *specCorrector) isStorageClassNameUndefinedInSpec() bool {
	storageClassName := r.kubegresContext.Kubegres.Spec.Database.StorageClassName
	return storageClassName == nil || *storageClassName == ""
}

func (r *specCorrector) updateSpec() error {
	r.kubegresContext.Log.Info("Updating Kubegres Spec", "name", r.kubegresContext.Kubegres.Name)
	return r.kubegresContext.Client.Update(r.kubegresContext.Ctx, r.kubegresContext.Kubegres)
}
