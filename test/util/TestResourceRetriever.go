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

package util

import (
	"context"
	v1 "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"log"
	postgresv1 "reactive-tech.io/kubegres/api/v1"
	"reactive-tech.io/kubegres/controllers/ctx"
	"reactive-tech.io/kubegres/test/resourceConfigs"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TestResourceRetriever struct {
	client    client.Client
	namespace string
}

type TestKubegresResources struct {
	NbreDeployedPrimary  int
	NbreDeployedReplicas int
	AreAllReady          bool
	Resources            []TestKubegresResource
	BackUpCronJob        TestKubegresBackUpCronJob
}

type TestKubegresResource struct {
	IsPrimary   bool
	IsReady     bool
	Pod         TestKubegresPod
	StatefulSet TestKubegresStatefulSet
	Pvc         TestKubegresPvc
}

type TestKubegresBackUpCronJob struct {
	Name string
	Spec batch.CronJobSpec
}

type TestKubegresPod struct {
	Name     string
	Metadata metav1.ObjectMeta
	Spec     core.PodSpec
	Resource *core.Pod
}

type TestKubegresStatefulSet struct {
	Name     string
	Metadata metav1.ObjectMeta
	Spec     v1.StatefulSetSpec
	Resource *v1.StatefulSet
}

type TestKubegresPvc struct {
	Name     string
	Spec     core.PersistentVolumeClaimSpec
	Resource *core.PersistentVolumeClaim
}

func CreateTestResourceRetriever(k8sClient client.Client, namespace string) TestResourceRetriever {
	resourceRetriever := TestResourceRetriever{client: k8sClient, namespace: namespace}
	return resourceRetriever
}

func (r *TestResourceRetriever) GetServiceNameAllowingToSqlQueryDb(kubegresName string, isPrimaryDb bool) string {
	if isPrimaryDb {
		return resourceConfigs.ServiceToSqlQueryPrimaryDbResourceName + "-" + kubegresName
	}
	return resourceConfigs.ServiceToSqlQueryReplicaDbServiceResourceName + "-" + kubegresName
}

func (r *TestResourceRetriever) GetKubegres() (*postgresv1.Kubegres, error) {
	return r.GetKubegresByName(resourceConfigs.KubegresResourceName)
}

func (r *TestResourceRetriever) GetKubegresByName(resourceName string) (*postgresv1.Kubegres, error) {
	resourceToRetrieve := &postgresv1.Kubegres{}
	err := r.getResource(resourceName, resourceToRetrieve)
	return resourceToRetrieve, err
}

func (r *TestResourceRetriever) GetService(serviceResourceName string) (*core.Service, error) {
	resourceToRetrieve := &core.Service{}
	err := r.getResource(serviceResourceName, resourceToRetrieve)
	return resourceToRetrieve, err
}

func (r *TestResourceRetriever) GetBackUpPvc() (*core.PersistentVolumeClaim, error) {
	resourceToRetrieve := &core.PersistentVolumeClaim{}
	err := r.getResource(resourceConfigs.BackUpPvcResourceName, resourceToRetrieve)
	return resourceToRetrieve, err
}

func (r *TestResourceRetriever) GetKubegresPvc() (*core.PersistentVolumeClaimList, error) {
	return r.GetKubegresPvcByKubegresName(resourceConfigs.KubegresResourceName)
}

func (r *TestResourceRetriever) GetKubegresPvcByKubegresName(kubegresName string) (*core.PersistentVolumeClaimList, error) {

	list := &core.PersistentVolumeClaimList{}
	opts := []client.ListOption{
		client.InNamespace(r.namespace),
		client.MatchingLabels{"app": kubegresName},
	}
	ctx := context.Background()
	err := r.client.List(ctx, list, opts...)
	return list, err
}

func (r *TestResourceRetriever) getResource(resourceNameToRetrieve string, resourceToRetrieve client.Object) error {
	ctx := context.Background()
	lookupKey := types.NamespacedName{Name: resourceNameToRetrieve, Namespace: r.namespace}
	return r.client.Get(ctx, lookupKey, resourceToRetrieve)
}

/*
func (r *TestResourceRetriever) doesResourceExist(resourceName string, resourceType runtime.Object) bool {
	err := r.getResource(resourceName, resourceType)
	return err != nil
}*/

func (r *TestResourceRetriever) GetKubegresResources() (TestKubegresResources, error) {
	return r.GetKubegresResourcesByName(resourceConfigs.KubegresResourceName)
}

func (r *TestResourceRetriever) GetKubegresResourcesByName(kubegresName string) (TestKubegresResources, error) {

	testKubegresResources := TestKubegresResources{}
	statefulSetsList := &v1.StatefulSetList{}
	err := r.getResourcesList(kubegresName, statefulSetsList)

	if err != nil {
		return r.logAndReturnError("StatefulSetList", "-", err)
	}

	cronJobName := ctx.CronJobNamePrefix + kubegresName
	cronJob := &batch.CronJob{}
	err = r.getResource(cronJobName, cronJob)
	if err == nil {
		testKubegresResources.BackUpCronJob = TestKubegresBackUpCronJob{
			Name: cronJobName,
			Spec: cronJob.Spec,
		}
	}

	nbrePodsReady := 0

	for _, statefulSet := range statefulSetsList.Items {

		podName := statefulSet.Name + "-0"
		pod := &core.Pod{}
		err = r.getResource(podName, pod)
		if err != nil {
			return r.logAndReturnError("Pod", podName, err)
		}

		pvcName := "postgres-db-" + statefulSet.Name + "-0"
		pvc := &core.PersistentVolumeClaim{}
		err = r.getResource(pvcName, pvc)
		if err != nil {
			return r.logAndReturnError("Pvc", pvcName, err)
		}

		isPrimaryPod := r.isPrimaryPod(pod)
		isPodReady := r.isPodReady(pod)

		if isPrimaryPod {
			testKubegresResources.NbreDeployedPrimary += 1
		} else {
			testKubegresResources.NbreDeployedReplicas += 1
		}

		if isPodReady {
			nbrePodsReady += 1
		}

		statefulSetCopy := statefulSet

		kubegresPostgres := TestKubegresResource{
			IsReady:   isPodReady,
			IsPrimary: isPrimaryPod,
			Pod: TestKubegresPod{
				Name:     pod.Name,
				Metadata: pod.ObjectMeta,
				Spec:     pod.Spec,
				Resource: pod,
			},
			StatefulSet: TestKubegresStatefulSet{
				Name:     statefulSet.Name,
				Metadata: statefulSet.ObjectMeta,
				Spec:     statefulSet.Spec,
				Resource: &statefulSetCopy,
			},
			Pvc: TestKubegresPvc{
				Name:     pvc.Name,
				Spec:     pvc.Spec,
				Resource: pvc,
			},
		}

		testKubegresResources.Resources = append(testKubegresResources.Resources, kubegresPostgres)
	}

	if testKubegresResources.NbreDeployedPrimary > 0 {
		testKubegresResources.AreAllReady = nbrePodsReady == (testKubegresResources.NbreDeployedPrimary + testKubegresResources.NbreDeployedReplicas)
	}

	return testKubegresResources, nil
}

func (r *TestResourceRetriever) getResourcesList(kubegresName string, resourceTypeToRetrieve client.ObjectList) error {
	ctx := context.Background()
	opts := []client.ListOption{
		client.InNamespace(r.namespace),
		client.MatchingLabels{"app": kubegresName},
	}
	return r.client.List(ctx, resourceTypeToRetrieve, opts...)
}

func (r *TestResourceRetriever) isPodReady(pod *core.Pod) bool {

	if len(pod.Status.ContainerStatuses) == 0 {
		return false
	}

	return pod.Status.ContainerStatuses[0].Ready
}

func (r *TestResourceRetriever) isPrimaryPod(pod *core.Pod) bool {
	return pod.Labels["replicationRole"] == resourceConfigs.PrimaryReplicationRole
}

func (r *TestResourceRetriever) logAndReturnError(resourceType, resourceName string, err error) (TestKubegresResources, error) {
	if apierrors.IsNotFound(err) {
		log.Println("There is not any deployed Kubegres " + resourceType + " with name '" + resourceName + "' yet.")
		err = nil
	} else {
		log.Println("Error while retrieving Kubegres "+resourceType+" with name '"+resourceName+"'. Given error: ", err)
	}
	return TestKubegresResources{}, err
}
