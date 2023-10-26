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

package defaultspec

import (
	"errors"
	storage "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"reactive-tech.io/kubegres/internal/controller/ctx"
)

type DefaultStorageClass struct {
	kubegresContext ctx.KubegresContext
}

func CreateDefaultStorageClass(kubegresContext ctx.KubegresContext) DefaultStorageClass {
	return DefaultStorageClass{kubegresContext}
}

func (r *DefaultStorageClass) GetDefaultStorageClassName() (string, error) {

	list := &storage.StorageClassList{}
	err := r.kubegresContext.Client.List(r.kubegresContext.Ctx, list)

	if err != nil {
		if apierrors.IsNotFound(err) {
			err = nil
		} else {
			r.kubegresContext.Log.ErrorEvent("DefaultStorageClassNameErr", err, "Unable to retrieve the name of the default storage class in the Kubernetes cluster.", "Namespace", r.kubegresContext.Kubegres.Namespace)
		}
	}

	for _, storageClass := range list.Items {
		if storageClass.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
			return storageClass.Name, nil
		}
	}

	err = errors.New("Unable to retrieve the name of the default storage class in the Kubernetes cluster. Namespace: " + r.kubegresContext.Kubegres.Namespace)
	r.kubegresContext.Log.ErrorEvent("DefaultStorageClassNameErr", err, "Unable to retrieve the name of the default storage class in the Kubernetes cluster.", "Namespace", r.kubegresContext.Kubegres.Namespace)

	return "", err
}
