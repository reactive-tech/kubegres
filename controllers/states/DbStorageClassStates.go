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
	storage "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"reactive-tech.io/kubegres/controllers/ctx"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DbStorageClassStates struct {
	IsDeployed       bool
	StorageClassName string

	kubegresContext ctx.KubegresContext
}

func loadDbStorageClass(kubegresContext ctx.KubegresContext) (DbStorageClassStates, error) {
	dbStorageClassStates := DbStorageClassStates{kubegresContext: kubegresContext}
	err := dbStorageClassStates.loadStates()

	return dbStorageClassStates, err
}

func (r *DbStorageClassStates) loadStates() error {

	dbStorageClass, err := r.GetStorageClass()
	if err != nil {
		return err
	}

	if dbStorageClass.Name != "" {
		r.IsDeployed = true
		r.StorageClassName = dbStorageClass.Name
	}

	return nil
}

func (r *DbStorageClassStates) GetStorageClass() (*storage.StorageClass, error) {

	namespace := ""
	resourceName := r.getSpecStorageClassName()
	resourceKey := client.ObjectKey{Namespace: namespace, Name: resourceName}
	storageClass := &storage.StorageClass{}

	err := r.kubegresContext.Client.Get(r.kubegresContext.Ctx, resourceKey, storageClass)

	if err != nil {
		if apierrors.IsNotFound(err) {
			err = nil
		} else {
			r.kubegresContext.Log.ErrorEvent("DatabaseStorageClassLoadingErr", err, "Unable to load any deployed Database StorageClass.", "StorageClass name", resourceName)
		}
	}

	return storageClass, err
}

func (r *DbStorageClassStates) getSpecStorageClassName() string {
	return *r.kubegresContext.Kubegres.Spec.Database.StorageClassName
}
