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

package ctx

import (
	"context"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	postgresV1 "reactive-tech.io/kubegres/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateOwnerKeyIndexation(mgr ctrl.Manager,
	ctx context.Context) error {

	if err := mgr.GetFieldIndexer().IndexField(ctx, &apps.StatefulSet{}, DeploymentOwnerKey, func(rawObj client.Object) []string {
		// grab the StatefultSet object, extract the owner...
		depl := rawObj.(*apps.StatefulSet)
		owner := metav1.GetControllerOf(depl)
		if owner == nil {
			return nil
		}
		// ...make sure it's a Kubegres...
		if owner.APIVersion != postgresV1.GroupVersion.String() || owner.Kind != KindKubegres {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &core.Service{}, DeploymentOwnerKey, func(rawObj client.Object) []string {
		// grab the Service object, extract the owner...
		depl := rawObj.(*core.Service)
		owner := metav1.GetControllerOf(depl)
		if owner == nil {
			return nil
		}
		// ...make sure it's a Kubegres...
		if owner.APIVersion != postgresV1.GroupVersion.String() || owner.Kind != KindKubegres {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &core.ConfigMap{}, DeploymentOwnerKey, func(rawObj client.Object) []string {
		// grab the CustomConfig object, extract the owner...
		depl := rawObj.(*core.ConfigMap)
		owner := metav1.GetControllerOf(depl)
		if owner == nil {
			return nil
		}
		// ...make sure it's a Kubegres...
		if owner.APIVersion != postgresV1.GroupVersion.String() || owner.Kind != KindKubegres {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &core.Pod{}, DeploymentOwnerKey, func(rawObj client.Object) []string {
		// grab the Pod object, extract the owner...
		depl := rawObj.(*core.Pod)
		owner := metav1.GetControllerOf(depl)
		if owner == nil {
			return nil
		}
		// ...make sure it's a Kubegres...
		if owner.APIVersion != postgresV1.GroupVersion.String() || owner.Kind != KindKubegres {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return nil
}
