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
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"reactive-tech.io/kubegres/controllers/ctx"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServicesStates struct {
	Primary ServiceWrapper
	Replica ServiceWrapper

	kubegresContext ctx.KubegresContext
}

type ServiceWrapper struct {
	Name       string
	IsDeployed bool
	Service    core.Service
}

func loadServicesStates(kubegresContext ctx.KubegresContext) (ServicesStates, error) {
	servicesStates := ServicesStates{kubegresContext: kubegresContext}
	err := servicesStates.loadStates()
	return servicesStates, err
}

func (r *ServicesStates) loadStates() (err error) {

	deployedServices, err := r.getDeployedServices()
	if err != nil {
		return err
	}

	for _, service := range deployedServices.Items {

		serviceWrapper := ServiceWrapper{IsDeployed: true, Service: service}

		if r.isPrimary(service) {
			serviceWrapper.Name = r.kubegresContext.GetServiceResourceName(true)
			r.Primary = serviceWrapper

		} else {
			serviceWrapper.Name = r.kubegresContext.GetServiceResourceName(false)
			r.Replica = serviceWrapper
		}
	}

	return nil
}

func (r *ServicesStates) getDeployedServices() (*core.ServiceList, error) {

	list := &core.ServiceList{}
	opts := []client.ListOption{
		client.InNamespace(r.kubegresContext.Kubegres.Namespace),
		client.MatchingFields{ctx.DeploymentOwnerKey: r.kubegresContext.Kubegres.Name},
	}
	err := r.kubegresContext.Client.List(r.kubegresContext.Ctx, list, opts...)

	if err != nil {
		if apierrors.IsNotFound(err) {
			err = nil
		} else {
			r.kubegresContext.Log.ErrorEvent("ServiceLoadingErr", err, "Unable to load any deployed Services.", "Kubegres name", r.kubegresContext.Kubegres.Name)
		}
	}

	return list, err
}

func (r *ServicesStates) isPrimary(service core.Service) bool {
	return service.Labels["replicationRole"] == ctx.PrimaryRoleName
}
