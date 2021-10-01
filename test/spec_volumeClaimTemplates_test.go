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
package test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v12 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"log"
	postgresv1 "reactive-tech.io/kubegres/api/v1"
	"reactive-tech.io/kubegres/controllers/ctx"
	"reactive-tech.io/kubegres/test/resourceConfigs"
	"reactive-tech.io/kubegres/test/util"
	"reactive-tech.io/kubegres/test/util/testcases"
	"reflect"
	"time"
)

var _ = Describe("Setting Kubegres spec 'volume.volumeClaimTemplates'", func() {

	var test = SpecVolumeClaimTemplatesTest{}

	BeforeEach(func() {
		//Skip("Temporarily skipping test")

		namespace := resourceConfigs.DefaultNamespace
		test.resourceRetriever = util.CreateTestResourceRetriever(k8sClientTest, namespace)
		test.resourceCreator = util.CreateTestResourceCreator(k8sClientTest, test.resourceRetriever, namespace)
		test.dbQueryTestCases = testcases.InitDbQueryTestCases(test.resourceCreator, resourceConfigs.KubegresResourceName)
	})

	AfterEach(func() {
		if !test.keepCreatedResourcesForNextTest {
			test.resourceCreator.DeleteAllTestResources()
		} else {
			test.keepCreatedResourcesForNextTest = false
		}
	})

	Context("GIVEN new Kubegres is created with a 'volume.volumeClaimTemplate' and 'volume.volumeMount' which have a reserved name", func() {

		It("THEN 2 error events should be logged as it is not possible to use a reserved name", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with a 'volume.volumeClaimTemplate' and 'volume.volumeMount' which have a reserved name'")

			reservedVolumeName := ctx.DatabaseVolumeName

			cacheVolumeClaim := test.givenVolumeClaimTemplate(reservedVolumeName, "10Mi")
			customVolumeClaims := []postgresv1.VolumeClaimTemplate{cacheVolumeClaim}

			cacheVolumeMount := test.givenVolumeMount(reservedVolumeName, "/cache")
			customVolumeMounts := []v12.VolumeMount{cacheVolumeMount}

			test.givenNewKubegresSpecIsSetTo(customVolumeClaims, customVolumeMounts, 3)

			test.whenKubegresIsCreated()

			test.thenErrorEventShouldBeLoggedAboutVolumeClaimTemplateName()
			test.thenErrorEventShouldBeLoggedAboutVolumeMountName()

			log.Print("END OF: Test 'GIVEN new Kubegres is created with a 'volume.volumeClaimTemplate' and 'volume.volumeMount' which have a reserved name'")
		})
	})

	Context("GIVEN new Kubegres is created with specs 'volume.volumeClaimTemplates' and 'volume.volumeMount' and later "+
		"it is updated by adding/removing 'volume.volumeClaimTemplates'", func() {

		It("GIVEN new Kubegres is created with a new custom 'volume.volumeClaimTemplates' and 'volume.volumeMount' "+
			"AND spec 'replica' set to 3 THEN 1 primary and 2 replica should be created with one custom volumeClaimTemplate "+
			"and volumeMount in StatefulSets", func() {

			log.Print("GIVEN new Kubegres is created with a new custom 'volume.volumeClaimTemplates' and " +
				"'volume.volumeMount' and spec 'replica' set to 3")

			cacheVolumeClaim := test.givenVolumeClaimTemplate("cache", "10Mi")
			customVolumeClaims := []postgresv1.VolumeClaimTemplate{cacheVolumeClaim}

			cacheVolumeMount := test.givenVolumeMount("cache", "/cache")
			customVolumeMounts := []v12.VolumeMount{cacheVolumeMount}

			test.givenNewKubegresSpecIsSetTo(customVolumeClaims, customVolumeMounts, 3)

			test.whenKubegresIsCreated()

			test.thenStatefulSetsStatesShouldBe(customVolumeClaims, customVolumeMounts, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(customVolumeClaims, customVolumeMounts)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.keepCreatedResourcesForNextTest = true

			log.Print("END OF: Test 'GIVEN new Kubegres is created with a new custom 'volume.volumeClaimTemplates' and " +
				"'volume.volumeMount' and spec 'replica' set to 3'")
		})

		It("GIVEN existing Kubegres is updated by adding new and updating existing custom 'volume.volumeClaimTemplates' "+
			"and 'volume.volumeMount' THEN StatefulSets should be updated too", func() {

			log.Print("START OF: Test 'GIVEN existing Kubegres is updated by adding new and updating existing custom " +
				"'volume.volumeClaimTemplates' and 'volume.volumeMount'")

			customVolumeClaimsToUpdate := []postgresv1.VolumeClaimTemplate{}

			cacheVolumeMount := test.givenVolumeMount("cache", "/cachetwo")
			customVolumeMountsToUpdate := []v12.VolumeMount{cacheVolumeMount}

			test.givenVolumesAreUpdatedOrAddedToTheExistingKubegresSpec(customVolumeClaimsToUpdate, customVolumeMountsToUpdate)

			test.whenKubernetesIsUpdated()

			cacheVolumeClaim := test.givenVolumeClaimTemplate("cache", "10Mi")
			expectedCustomVolumeClaims := []postgresv1.VolumeClaimTemplate{cacheVolumeClaim}
			expectedCustomVolumeMounts := []v12.VolumeMount{cacheVolumeMount}

			test.thenStatefulSetsStatesShouldBe(expectedCustomVolumeClaims, expectedCustomVolumeMounts, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(expectedCustomVolumeClaims, expectedCustomVolumeMounts)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.keepCreatedResourcesForNextTest = true

			log.Print("END OF: Test 'GIVEN existing Kubegres is updated by adding new and updating existing custom " +
				"'volume.volumeClaimTemplates' and 'volume.volumeMount'")
		})

		It("GIVEN existing Kubegres is updated with the update of one custom 'volume.volumeClaimTemplates' from 10Mi to 15Mi "+
			"THEN an error event should be logged as it is not possible to update 'volume.volumeClaimTemplates'", func() {

			log.Print("START OF: Test 'GIVEN existing Kubegres is updated with the update of one custom 'volume.volumeClaimTemplates' from 10Mi to 15Mi'")

			cacheVolume := test.givenVolumeClaimTemplate("cache", "15Mi")
			customVolumesUpdate := []postgresv1.VolumeClaimTemplate{cacheVolume}

			customVolumeMountsToUpdate := []v12.VolumeMount{}

			test.givenVolumesAreUpdatedOrAddedToTheExistingKubegresSpec(customVolumesUpdate, customVolumeMountsToUpdate)

			eventRecorderTest.RemoveAllEvents()

			test.whenKubernetesIsUpdated()

			test.thenErrorEventShouldBeLoggedAboutVolumeClaimTemplateSpecChanged()

			test.keepCreatedResourcesForNextTest = true

			log.Print("END OF: Test 'GIVEN existing Kubegres is updated with the update of one custom 'volume.volumeClaimTemplates' from 10Mi to 15Mi'")
		})

		It("GIVEN existing Kubegres is updated with the removal of one custom 'volume.volumeClaimTemplates' "+
			"THEN an error event should be logged as it is not possible to update 'volume.volumeClaimTemplates'", func() {

			log.Print("START OF: Test 'GIVEN existing Kubegres is updated with the removal of one custom 'volume.volumeClaimTemplates'")

			cacheVolume := test.givenVolumeClaimTemplate("cache", "10Mi")
			customVolumesToRemove := []postgresv1.VolumeClaimTemplate{cacheVolume}

			customVolumeMountsToRemove := []v12.VolumeMount{}

			test.givenVolumesAreRemovedFromTheExistingKubegresSpec(customVolumesToRemove, customVolumeMountsToRemove)

			eventRecorderTest.RemoveAllEvents()

			test.whenKubernetesIsUpdated()

			test.thenErrorEventShouldBeLoggedAboutVolumeClaimTemplateSpecChanged()

			log.Print("END OF: Test 'GIVEN existing Kubegres is updated with the removal of one custom 'volume.volumeClaimTemplates'")
		})

	})

})

type SpecVolumeClaimTemplatesTest struct {
	keepCreatedResourcesForNextTest bool
	kubegresResource                *postgresv1.Kubegres
	dbQueryTestCases                testcases.DbQueryTestCases
	resourceCreator                 util.TestResourceCreator
	resourceRetriever               util.TestResourceRetriever
}

func (r *SpecVolumeClaimTemplatesTest) givenVolumeClaimTemplate(volumeName, volumeSize string) postgresv1.VolumeClaimTemplate {

	storageClassName := "standard"
	quantity := resource.MustParse(volumeSize)
	requests := map[v12.ResourceName]resource.Quantity{}
	requests[v12.ResourceStorage] = quantity

	return postgresv1.VolumeClaimTemplate{
		Name: volumeName,
		Spec: v12.PersistentVolumeClaimSpec{
			AccessModes:      []v12.PersistentVolumeAccessMode{v12.ReadWriteOnce},
			StorageClassName: &storageClassName,
			Resources: v12.ResourceRequirements{
				Requests: requests,
			},
		},
	}
}

func (r *SpecVolumeClaimTemplatesTest) givenVolumeMount(volumeName, mountPath string) v12.VolumeMount {

	return v12.VolumeMount{
		Name:      volumeName,
		MountPath: mountPath,
	}
}

func (r *SpecVolumeClaimTemplatesTest) givenNewKubegresSpecIsSetTo(
	customVolumeClaims []postgresv1.VolumeClaimTemplate,
	customVolumeMounts []v12.VolumeMount,
	specNbreReplicas int32) {

	r.kubegresResource = resourceConfigs.LoadKubegresYaml()
	r.kubegresResource.Spec.Volume.VolumeMounts = append(r.kubegresResource.Spec.Volume.VolumeMounts, customVolumeMounts...)
	r.kubegresResource.Spec.Volume.VolumeClaimTemplates = append(r.kubegresResource.Spec.Volume.VolumeClaimTemplates, customVolumeClaims...)
	r.kubegresResource.Spec.Replicas = &specNbreReplicas
}

func (r *SpecVolumeClaimTemplatesTest) givenVolumesAreUpdatedOrAddedToTheExistingKubegresSpec(
	customVolumeClaimsToAddOrReplace []postgresv1.VolumeClaimTemplate,
	customVolumeMountsToAddOrReplace []v12.VolumeMount) {

	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()
	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	for _, customVolumeClaim := range customVolumeClaimsToAddOrReplace {
		volumeClaimIndex := r.getVolumeClaimIndex(customVolumeClaim)
		if volumeClaimIndex >= 0 {
			r.kubegresResource.Spec.Volume.VolumeClaimTemplates[volumeClaimIndex] = customVolumeClaim
		} else {
			r.kubegresResource.Spec.Volume.VolumeClaimTemplates = append(r.kubegresResource.Spec.Volume.VolumeClaimTemplates, customVolumeClaim)
		}
	}

	for _, customVolumeMount := range customVolumeMountsToAddOrReplace {
		volumeMountIndex := r.getVolumeMountIndex(customVolumeMount)
		if volumeMountIndex >= 0 {
			r.kubegresResource.Spec.Volume.VolumeMounts[volumeMountIndex] = customVolumeMount
		} else {
			r.kubegresResource.Spec.Volume.VolumeMounts = append(r.kubegresResource.Spec.Volume.VolumeMounts, customVolumeMount)
		}
	}
}

func (r *SpecVolumeClaimTemplatesTest) givenVolumesAreRemovedFromTheExistingKubegresSpec(
	customVolumesToRemove []postgresv1.VolumeClaimTemplate,
	customVolumeMountsToRemove []v12.VolumeMount) {

	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()
	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	for _, customVolume := range customVolumesToRemove {
		volumeIndex := r.getVolumeClaimIndex(customVolume)
		if volumeIndex >= 0 {
			r.kubegresResource.Spec.Volume.VolumeClaimTemplates = append(r.kubegresResource.Spec.Volume.VolumeClaimTemplates[:volumeIndex],
				r.kubegresResource.Spec.Volume.VolumeClaimTemplates[volumeIndex+1:]...)
		}
	}

	for _, customVolumeMount := range customVolumeMountsToRemove {
		volumeMountIndex := r.getVolumeMountIndex(customVolumeMount)
		if volumeMountIndex >= 0 {
			r.kubegresResource.Spec.Volume.VolumeMounts = append(r.kubegresResource.Spec.Volume.VolumeMounts[:volumeMountIndex],
				r.kubegresResource.Spec.Volume.VolumeMounts[volumeMountIndex+1:]...)
		}
	}
}

func (r *SpecVolumeClaimTemplatesTest) getVolumeClaimIndex(customVolumeClaim postgresv1.VolumeClaimTemplate) int {
	index := 0
	for _, volumeClaim := range r.kubegresResource.Spec.Volume.VolumeClaimTemplates {
		if customVolumeClaim.Name == volumeClaim.Name {
			return index
		}
		index++
	}
	return -1
}

func (r *SpecVolumeClaimTemplatesTest) getVolumeMountIndex(customVolumeMount v12.VolumeMount) int {
	index := 0
	for _, volumeMount := range r.kubegresResource.Spec.Volume.VolumeMounts {
		if customVolumeMount.Name == volumeMount.Name {
			return index
		}
		index++
	}
	return -1
}

func (r *SpecVolumeClaimTemplatesTest) whenKubegresIsCreated() {
	r.resourceCreator.CreateKubegres(r.kubegresResource)
}

func (r *SpecVolumeClaimTemplatesTest) whenKubernetesIsUpdated() {
	r.resourceCreator.UpdateResource(r.kubegresResource, "Kubegres")
}

//

func (r *SpecVolumeClaimTemplatesTest) thenErrorEventShouldBeLoggedAboutVolumeClaimTemplateName() {
	expectedErrorEvent := util.EventRecord{
		Eventtype: v12.EventTypeWarning,
		Reason:    "SpecCheckErr",
		Message: "In the Resources Spec the value of 'spec.Volume.VolumeClaimTemplates' has an entry with a volume name " +
			"which is a reserved name: " + ctx.DatabaseVolumeName + " . That name cannot be used and it is reserved for " +
			"Kubegres internal usages. Please change that name in the YAML.",
	}
	Eventually(func() bool {
		_, err := r.resourceRetriever.GetKubegres()
		if err != nil {
			return false
		}
		return eventRecorderTest.CheckEventExist(expectedErrorEvent)

	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *SpecVolumeClaimTemplatesTest) thenErrorEventShouldBeLoggedAboutVolumeMountName() {
	expectedErrorEvent := util.EventRecord{
		Eventtype: v12.EventTypeWarning,
		Reason:    "SpecCheckErr",
		Message: "In the Resources Spec the value of 'spec.Volume.VolumeMounts' has an entry with a volume name " +
			"which is a reserved name: " + ctx.DatabaseVolumeName + " . That name cannot be used and it is reserved for " +
			"Kubegres internal usages. Please change that name in the YAML.",
	}
	Eventually(func() bool {
		_, err := r.resourceRetriever.GetKubegres()
		if err != nil {
			return false
		}
		return eventRecorderTest.CheckEventExist(expectedErrorEvent)

	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *SpecVolumeClaimTemplatesTest) thenErrorEventShouldBeLoggedAboutVolumeClaimTemplateSpecChanged() {
	expectedErrorEvent := util.EventRecord{
		Eventtype: v12.EventTypeWarning,
		Reason:    "SpecCheckErr",
		Message: "In the Resources Spec, the array 'spec.Volume.VolumeClaimTemplates' has changed. Kubernetes does not " +
			"allow to update that field in StatefulSet specification. Please rollback your changes in the YAML.",
	}
	Eventually(func() bool {
		_, err := r.resourceRetriever.GetKubegres()
		if err != nil {
			return false
		}
		return eventRecorderTest.CheckEventExist(expectedErrorEvent)

	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *SpecVolumeClaimTemplatesTest) thenStatefulSetsStatesShouldBe(
	expectedCustomVolumeClaims []postgresv1.VolumeClaimTemplate,
	expectedCustomVolumeMounts []v12.VolumeMount,
	nbrePrimary,
	nbreReplicas int) bool {

	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		kubegresContext := ctx.KubegresContext{}

		for _, resource := range kubegresResources.Resources {

			for _, customVolumeClaim := range expectedCustomVolumeClaims {
				if !r.doesCustomVolumeExistsInStatefulSet(customVolumeClaim, resource.StatefulSet.Spec.VolumeClaimTemplates) {
					log.Println("StatefulSet '" + resource.StatefulSet.Name + "' doesn't have the expected custom volume " +
						"claim with name: '" + customVolumeClaim.Name + "'. Waiting...")
					return false
				}
			}

			for _, volumeClaimInStatefulSet := range resource.StatefulSet.Spec.VolumeClaimTemplates {
				if r.isCustomVolumeClaim(volumeClaimInStatefulSet, kubegresContext) &&
					!r.isVolumeClaimAnExpectedCustomVolume(volumeClaimInStatefulSet, expectedCustomVolumeClaims) {
					log.Println("StatefulSet '" + resource.StatefulSet.Name + "' still has custom volume claim " +
						"with name: '" + volumeClaimInStatefulSet.Name + "'. Waiting...")
					return false
				}
			}

			for _, customVolumeMount := range expectedCustomVolumeMounts {
				if !r.doesCustomVolumeMountExistsInStatefulSet(customVolumeMount, resource.StatefulSet.Spec.Template.Spec.Containers[0].VolumeMounts) {
					log.Println("StatefulSet '" + resource.StatefulSet.Name + "' doesn't have the expected custom volumeMount " +
						"with name: '" + customVolumeMount.Name + "'. Waiting...")
					return false
				}
			}

			for _, volumeMountInStatefulSet := range resource.StatefulSet.Spec.Template.Spec.Containers[0].VolumeMounts {
				if r.isCustomVolumeMount(volumeMountInStatefulSet, kubegresContext) &&
					!r.isVolumeMountAnExpectedCustomVolumeMount(volumeMountInStatefulSet, expectedCustomVolumeMounts) {
					log.Println("StatefulSet '" + resource.StatefulSet.Name + "' still has custom volumeMount with name: '" + volumeMountInStatefulSet.Name + "'. Waiting...")
					return false
				}
			}
		}

		if kubegresResources.AreAllReady &&
			kubegresResources.NbreDeployedPrimary == nbrePrimary &&
			kubegresResources.NbreDeployedReplicas == nbreReplicas {

			time.Sleep(resourceConfigs.TestRetryInterval)
			log.Println("Deployed and Ready StatefulSet check successful")
			return true
		}

		return false

	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *SpecVolumeClaimTemplatesTest) thenDeployedKubegresSpecShouldBeSetTo(
	expectedCustomVolumeClaims []postgresv1.VolumeClaimTemplate,
	expectedCustomVolumeMounts []v12.VolumeMount) {

	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	Expect(r.kubegresResource.Spec.Volume.VolumeClaimTemplates).Should(Equal(expectedCustomVolumeClaims))
	Expect(r.kubegresResource.Spec.Volume.VolumeMounts).Should(Equal(expectedCustomVolumeMounts))
}

func (r *SpecVolumeClaimTemplatesTest) isCustomVolumeClaim(volumeClaim v12.PersistentVolumeClaim, kubegresContext ctx.KubegresContext) bool {
	return !kubegresContext.IsReservedVolumeName(volumeClaim.Name)
}

func (r *SpecVolumeClaimTemplatesTest) isCustomVolumeMount(volumeMount v12.VolumeMount, kubegresContext ctx.KubegresContext) bool {
	return !kubegresContext.IsReservedVolumeName(volumeMount.Name)
}

func (r *SpecVolumeClaimTemplatesTest) doesCustomVolumeExistsInStatefulSet(
	customVolumeClaim postgresv1.VolumeClaimTemplate,
	statefulSetVolumeClaims []v12.PersistentVolumeClaim) bool {

	for _, statefulSetVolumeClaim := range statefulSetVolumeClaims {
		if r.areClaimsEqual(statefulSetVolumeClaim, customVolumeClaim) {
			return true
		}
	}
	return false
}

func (r *SpecVolumeClaimTemplatesTest) isVolumeClaimAnExpectedCustomVolume(
	volumeClaimToCheck v12.PersistentVolumeClaim,
	expectedCustomVolumeClaims []postgresv1.VolumeClaimTemplate) bool {

	for _, expectedCustomVolumeClaim := range expectedCustomVolumeClaims {
		if r.areClaimsEqual(volumeClaimToCheck, expectedCustomVolumeClaim) {
			return true
		}
	}
	return false
}

func (r *SpecVolumeClaimTemplatesTest) areClaimsEqual(
	persistentVolumeClaim v12.PersistentVolumeClaim,
	volumeClaimTemplate postgresv1.VolumeClaimTemplate) bool {

	return persistentVolumeClaim.Name == volumeClaimTemplate.Name &&
		reflect.DeepEqual(persistentVolumeClaim.Spec.StorageClassName, volumeClaimTemplate.Spec.StorageClassName) &&
		reflect.DeepEqual(persistentVolumeClaim.Spec.Resources, volumeClaimTemplate.Spec.Resources)
}

func (r *SpecVolumeClaimTemplatesTest) isVolumeMountAnExpectedCustomVolumeMount(
	volumeMountToCheck v12.VolumeMount,
	expectedCustomVolumeMounts []v12.VolumeMount) bool {

	for _, expectedCustomVolumeMount := range expectedCustomVolumeMounts {
		if reflect.DeepEqual(expectedCustomVolumeMount, volumeMountToCheck) {
			return true
		}
	}
	return false
}

func (r *SpecVolumeClaimTemplatesTest) doesCustomVolumeMountExistsInStatefulSet(customVolumeMount v12.VolumeMount, statefulSetVolumeMounts []v12.VolumeMount) bool {
	for _, statefulSetVolumeMount := range statefulSetVolumeMounts {
		if reflect.DeepEqual(statefulSetVolumeMount, customVolumeMount) {
			return true
		}
	}
	return false
}
