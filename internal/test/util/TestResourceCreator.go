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

package util

import (
	"context"
	"github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"log"
	postgresv1 "reactive-tech.io/kubegres/api/v1"
	"reactive-tech.io/kubegres/internal/test/resourceConfigs"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"time"
)

type TestResourceCreator struct {
	client            client.Client
	resourceRetriever TestResourceRetriever
	namespace         string
}

func CreateTestResourceCreator(k8sClient client.Client,
	resourceRetriever TestResourceRetriever,
	namespace string) TestResourceCreator {

	if namespace == "" {
		namespace = resourceConfigs.DefaultNamespace
	}

	resourceCreator := TestResourceCreator{client: k8sClient, resourceRetriever: resourceRetriever, namespace: namespace}

	resourceCreator.CreateNamespace()
	resourceCreator.CreateConfigMapWithPrimaryInitScript()
	resourceCreator.CreateSecret()

	return resourceCreator
}

func (r *TestResourceCreator) CreateKubegres(resourceToCreate *postgresv1.Kubegres) {
	ctx := context.Background()
	err := r.client.Create(ctx, resourceToCreate)
	if err != nil {
		log.Println("Error while creating Kubegres resource : ", err)
		gomega.Expect(err).Should(gomega.Succeed())
	} else {
		log.Println("Kubegres resource created")
	}
}

func (r *TestResourceCreator) UpdateResource(resourceToUpdate client.Object, resourceName string) {
	ctx := context.Background()
	err := r.client.Update(ctx, resourceToUpdate)
	if err != nil {
		log.Println("Error while updating resource '"+resourceName+"': ", err)
		gomega.Expect(err).Should(gomega.Succeed())
	} else {
		log.Println("Resources '" + resourceName + "' updated")
	}
}

func (r *TestResourceCreator) CreateSecret() {
	existingResource := v1.Secret{}
	resourceToCreate := resourceConfigs.LoadSecretYaml()
	resourceToCreate.Namespace = r.namespace
	r.createResourceFromYaml("Secret", resourceConfigs.SecretResourceName, &existingResource, &resourceToCreate)
}

func (r *TestResourceCreator) CreateBackUpPvc() {
	existingResource := v1.PersistentVolumeClaim{}
	resourceToCreate := resourceConfigs.LoadBackUpPvcYaml()
	resourceToCreate.Namespace = r.namespace
	r.createResourceFromYaml("BackUp PVC", resourceConfigs.BackUpPvcResourceName, &existingResource, resourceToCreate)
}

func (r *TestResourceCreator) CreateBackUpPvc2() {
	existingResource := v1.PersistentVolumeClaim{}
	resourceToCreate := resourceConfigs.LoadBackUpPvcYaml()
	resourceToCreate.Name = resourceConfigs.BackUpPvcResourceName2
	resourceToCreate.Namespace = r.namespace
	r.createResourceFromYaml("BackUp PVC 2", resourceConfigs.BackUpPvcResourceName2, &existingResource, resourceToCreate)
}

func (r *TestResourceCreator) CreateConfigMapEmpty() {
	existingResource := v1.ConfigMap{}
	resourceToCreate := resourceConfigs.LoadCustomConfigMapYaml(resourceConfigs.CustomConfigMapEmptyYamlFile)
	resourceToCreate.Namespace = r.namespace
	r.createResourceFromYaml("Custom ConfigMap empty", resourceConfigs.CustomConfigMapEmptyResourceName, &existingResource, &resourceToCreate)
}

func (r *TestResourceCreator) CreateConfigMapWithAllConfigs() {
	existingResource := v1.ConfigMap{}
	resourceToCreate := resourceConfigs.LoadCustomConfigMapYaml(resourceConfigs.CustomConfigMapWithAllConfigsYamlFile)
	resourceToCreate.Namespace = r.namespace
	r.createResourceFromYaml("Custom ConfigMap with all configs", resourceConfigs.CustomConfigMapWithAllConfigsResourceName, &existingResource, &resourceToCreate)
}

func (r *TestResourceCreator) CreateConfigMapWithBackupDatabaseScript() {
	existingResource := v1.ConfigMap{}
	resourceToCreate := resourceConfigs.LoadCustomConfigMapYaml(resourceConfigs.CustomConfigMapWithBackupDatabaseScriptYamlFile)
	resourceToCreate.Namespace = r.namespace
	r.createResourceFromYaml("Custom ConfigMap with backup database script", resourceConfigs.CustomConfigMapWithBackupDatabaseScriptResourceName, &existingResource, &resourceToCreate)
}

func (r *TestResourceCreator) CreateConfigMapWithPgHbaConf() {
	existingResource := v1.ConfigMap{}
	resourceToCreate := resourceConfigs.LoadCustomConfigMapYaml(resourceConfigs.CustomConfigMapWithPgHbaConfYamlFile)
	resourceToCreate.Namespace = r.namespace
	r.createResourceFromYaml("Custom ConfigMap with pg-hba configuration resource", resourceConfigs.CustomConfigMapWithPgHbaConfResourceName, &existingResource, &resourceToCreate)
}

func (r *TestResourceCreator) CreateConfigMapWithPostgresConf() {
	existingResource := v1.ConfigMap{}
	resourceToCreate := resourceConfigs.LoadCustomConfigMapYaml(resourceConfigs.CustomConfigMapWithPostgresConfYamlFile)
	resourceToCreate.Namespace = r.namespace
	r.createResourceFromYaml("Custom ConfigMap with postgres conf", resourceConfigs.CustomConfigMapWithPostgresConfResourceName, &existingResource, &resourceToCreate)
}

func (r *TestResourceCreator) CreateConfigMapWithPostgresConfAndWalLevelSetToLogical() {
	existingResource := v1.ConfigMap{}
	resourceToCreate := resourceConfigs.LoadCustomConfigMapYaml(resourceConfigs.CustomConfigMapWithPostgresConfAndWalLevelSetToLogicalYamlFile)
	resourceToCreate.Namespace = r.namespace
	r.createResourceFromYaml("Custom ConfigMap with postgres conf and wal_level=logical", resourceConfigs.CustomConfigMapWithPostgresConfAndWalLevelSetToLogicalResourceName, &existingResource, &resourceToCreate)
}

func (r *TestResourceCreator) CreateNamespace() {
	if r.namespace == resourceConfigs.DefaultNamespace {
		return
	}

	if r.doesNamespaceExist() {
		log.Println("Namespace '" + r.namespace + "' already exist. All good.")
		return
	}

	existingResource := v1.Namespace{}
	resourceToCreate := v1.Namespace{}
	resourceToCreate.Name = r.namespace
	resourceToCreate.Labels = make(map[string]string)
	resourceToCreate.Labels["name"] = r.namespace
	r.createResourceFromYaml("Custom Namespace", resourceConfigs.CustomConfigMapWithPrimaryInitScriptResourceName, &existingResource, &resourceToCreate)
}

func (r *TestResourceCreator) CreateConfigMapWithPrimaryInitScript() {
	existingResource := v1.ConfigMap{}
	resourceToCreate := resourceConfigs.LoadCustomConfigMapYaml(resourceConfigs.CustomConfigMapWithPrimaryInitScriptYamlFile)
	resourceToCreate.Namespace = r.namespace
	r.createResourceFromYaml("Custom ConfigMap with primary init script", resourceConfigs.CustomConfigMapWithPrimaryInitScriptResourceName, &existingResource, &resourceToCreate)
}

func (r *TestResourceCreator) CreateServiceToSqlQueryDb(kubegresName string, nodePort int, isPrimaryDb bool) (runtime.Object, error) {
	existingResource := v1.Service{}
	serviceResourceName := r.resourceRetriever.GetServiceNameAllowingToSqlQueryDb(kubegresName, isPrimaryDb)
	resourceToCreate := r.loadYamlServiceToSqlQueryDb(isPrimaryDb)
	resourceToCreate.Spec.Ports[0].Port = resourceConfigs.DbPort
	resourceToCreate.Spec.Ports[0].NodePort = int32(nodePort)
	resourceToCreate.Namespace = r.namespace
	resourceToCreate.Name = serviceResourceName
	resourceToCreate.Spec.Selector["app"] = kubegresName

	resourceLabel := "Service " + serviceResourceName + " (nodePort: " + strconv.Itoa(nodePort) + ")"

	return r.createResourceFromYaml(resourceLabel, serviceResourceName, &existingResource, &resourceToCreate)
}

func (r *TestResourceCreator) DeleteResource(resourceToDelete client.Object, resourceName string) bool {
	ctx := context.Background()
	err := r.client.Delete(ctx, resourceToDelete)
	if err != nil {
		log.Println("Error while deleting resource with name: '"+resourceName+"' ", err)
		return false
	}
	log.Println("Deleted resource with name: '" + resourceName + "'")
	return true
}

func (r *TestResourceCreator) DeleteAllTestResources(resourceNamesToNotDelete ...string) {

	log.Println("Deleting all resources created during tests")

	configMapsList := &v1.ConfigMapList{}
	r.searchList(configMapsList)
	for _, resourceToDelete := range configMapsList.Items {
		if !r.doesArrayContain(resourceToDelete.Name, resourceNamesToNotDelete...) {
			r.DeleteResource(&resourceToDelete, resourceToDelete.Name)
		}
	}

	servicesList := &v1.ServiceList{}
	r.searchList(servicesList)
	for _, resourceToDelete := range servicesList.Items {
		if !r.doesArrayContain(resourceToDelete.Name, resourceNamesToNotDelete...) {
			r.DeleteResource(&resourceToDelete, resourceToDelete.Name)
		}
	}

	pvcList := &v1.PersistentVolumeClaimList{}
	r.searchList(pvcList)
	for _, resourceToDelete := range pvcList.Items {
		if !r.doesArrayContain(resourceToDelete.Name, resourceNamesToNotDelete...) {
			r.DeleteResource(&resourceToDelete, resourceToDelete.Name)
		}
	}

	kubegresList := &postgresv1.KubegresList{}
	r.searchList(kubegresList)
	for _, resourceToDelete := range kubegresList.Items {

		kubegresPvcList, err := r.resourceRetriever.GetKubegresPvcByKubegresName(resourceToDelete.Name)
		r.DeleteResource(&resourceToDelete, resourceToDelete.Name)

		if err == nil {
			for _, pvcToDelete := range kubegresPvcList.Items {
				r.DeleteResource(&pvcToDelete, pvcToDelete.Name)
			}
		} else {
			log.Println("No PVC found for kubegres resource '" + resourceToDelete.Name + "'")
			continue
		}
	}

	log.Println("Deleted all resources created during tests. Waiting for 30 seconds...")
	time.Sleep(30 * time.Second)
}

func (r *TestResourceCreator) doesArrayContain(valueToSearch string, resourceNamesToNotDelete ...string) bool {
	for _, value := range resourceNamesToNotDelete {
		if value == valueToSearch {
			return true
		}
	}
	return false
}

func (r *TestResourceCreator) loadYamlServiceToSqlQueryDb(isPrimaryDb bool) v1.Service {
	if isPrimaryDb {
		return resourceConfigs.LoadYamlServiceToSqlQueryPrimaryDb()
	}
	return resourceConfigs.LoadYamlServiceToSqlQueryReplicaDb()
}

func (r *TestResourceCreator) doesNamespaceExist() bool {
	ctx := context.Background()
	resourceKey := client.ObjectKey{Name: r.namespace}
	existingResource := v1.Namespace{}
	err := r.client.Get(ctx, resourceKey, &existingResource)
	return err == nil
}

func (r *TestResourceCreator) createResourceFromYaml(resourceLabel, resourceName string, existingResource,
	resourceToCreate client.Object) (runtime.Object, error) {

	ctx := context.Background()
	lookupKey := types.NamespacedName{Name: resourceName, Namespace: r.namespace}
	err := r.client.Get(ctx, lookupKey, existingResource)

	if err != nil && !apierrors.IsNotFound(err) {
		log.Fatal("Error while getting "+resourceLabel+" resource : ", err)
		return existingResource, err

	} else if err == nil {
		log.Println("'" + resourceLabel + "' resource already exist. All good.")
		return existingResource, nil
	}

	err = r.client.Create(ctx, resourceToCreate)
	if err != nil {
		log.Println("Error while creating "+resourceLabel+" resource : ", err)
		gomega.Expect(err).Should(gomega.Succeed())
		return existingResource, err
	}

	log.Println("'" + resourceLabel + "' resource created")
	return resourceToCreate, nil
}

func (r *TestResourceCreator) searchList(listToSearch client.ObjectList) {

	opts := []client.ListOption{
		client.InNamespace(r.namespace),
		client.MatchingLabels{"environment": "acceptancetesting"},
	}
	ctx := context.Background()
	_ = r.client.List(ctx, listToSearch, opts...)
}
