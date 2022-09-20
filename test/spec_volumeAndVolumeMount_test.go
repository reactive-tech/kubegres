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

var _ = Describe("Setting Kubegres specs 'volume.volume' and 'volume.volumeMount'", func() {

	var test = SpecVolumeAndVolumeMountTest{}

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

	Context("GIVEN new Kubegres is created with a 'volume.volume' and 'volume.volumeMount' which have a reserved name", func() {

		It("THEN 2 error events should be logged as it is not possible to use a reserved name", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with a 'volume.volume' and 'volume.volumeMount' " +
				"which have a reserved name'")

			reservedVolumeName := ctx.DatabaseVolumeName

			cacheVolume := test.givenVolumeWithEmptyDir(reservedVolumeName)
			customVolumes := []v12.Volume{cacheVolume}

			cacheVolumeMount := test.givenVolumeMount(reservedVolumeName, "/cache")
			customVolumeMounts := []v12.VolumeMount{cacheVolumeMount}

			test.givenNewKubegresSpecIsSetTo(customVolumes, customVolumeMounts, 3)

			test.whenKubegresIsCreated()

			test.thenErrorEventShouldBeLoggedAboutVolumeName()
			test.thenErrorEventShouldBeLoggedAboutVolumeMountName()

			log.Print("END OF: Test 'GIVEN new Kubegres is created with a 'volume.volume' and 'volume.volumeMount' " +
				"which have a reserved name'")
		})
	})

	Context("GIVEN new Kubegres is created with a 'volume.volumeMount' which has a mountPath set to the path of "+
		"Postgres database folder", func() {

		It("THEN an error event should be logged as it is not possible to use Postgres database as a mountPath", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with a 'volume.volumeMount' which has a mountPath " +
				"set to the path of Postgres database folder'")

			cacheVolume := test.givenVolumeWithEmptyDir("cache-volume")
			customVolumes := []v12.Volume{cacheVolume}

			postgresDatabasePath := test.kubegresResource.Spec.Database.VolumeMount
			cacheVolumeMount := test.givenVolumeMount("cache-volume", postgresDatabasePath)
			customVolumeMounts := []v12.VolumeMount{cacheVolumeMount}

			test.givenNewKubegresSpecIsSetTo(customVolumes, customVolumeMounts, 3)

			test.whenKubegresIsCreated()

			test.thenErrorEventShouldBeLoggedAboutVolumeMountPath()

			log.Print("END OF: Test 'GIVEN new Kubegres is created with a 'volume.volumeMount' which has a mountPath " +
				"set to the path of Postgres database folder'")
		})
	})

	Context("GIVEN new Kubegres is created with specs 'volume.volume' and 'volume.volumeMount' and later "+
		"we update them and add additional volumes and finally we delete them", func() {

		It("GIVEN new Kubegres is created with a new custom 'volume.volume' and 'volume.volumeMount' AND spec 'replica' "+
			"set to 3 THEN 1 primary and 2 replica should be created with one custom volume and volumeMount in StatefulSets", func() {

			log.Print("GIVEN new Kubegres is created with a new custom 'volume.volume' and 'volume.volumeMount' and spec 'replica' set to 3")

			shmVolume := test.givenVolumeWithMemory("dshm", "200Mi")
			customVolumes := []v12.Volume{shmVolume}

			shmVolumeMount := test.givenVolumeMount("dshm", "/dev/shm")
			customVolumeMounts := []v12.VolumeMount{shmVolumeMount}

			test.givenNewKubegresSpecIsSetTo(customVolumes, customVolumeMounts, 3)

			test.whenKubegresIsCreated()

			test.thenStatefulSetsStatesShouldBe(customVolumes, customVolumeMounts, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(customVolumes, customVolumeMounts)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.keepCreatedResourcesForNextTest = true

			log.Print("END OF: Test 'GIVEN new Kubegres is created with a new custom 'volume.volume' and 'volume.volumeMount' " +
				"and spec 'replica' set to 3'")
		})

		It("GIVEN existing Kubegres is updated by updating by adding new and updating existing custom 'volume.volume' and 'volume.volumeMount' THEN "+
			"StatefulSets should be updated too", func() {

			log.Print("START OF: Test 'GIVEN existing Kubegres is updated by adding new and updating existing custom 'volume.volume' and 'volume.volumeMount'")

			shmVolume := test.givenVolumeWithMemory("dshm", "300Mi")
			cacheVolume := test.givenVolumeWithEmptyDir("cache-volume")
			customVolumesToAddOrUpdate := []v12.Volume{shmVolume, cacheVolume}

			cacheVolumeMount := test.givenVolumeMount("cache-volume", "/cache")
			customVolumeMountsToAddOrUpdate := []v12.VolumeMount{cacheVolumeMount}

			test.givenVolumesAreUpdatedOrAddedToTheExistingKubegresSpec(customVolumesToAddOrUpdate, customVolumeMountsToAddOrUpdate)

			test.whenKubernetesIsUpdated()

			shmVolume = test.givenVolumeWithMemory("dshm", "300Mi")
			cacheVolume = test.givenVolumeWithEmptyDir("cache-volume")
			expectedCustomVolumes := []v12.Volume{shmVolume, cacheVolume}

			shmVolumeMount := test.givenVolumeMount("dshm", "/dev/shm")
			cacheVolumeMount = test.givenVolumeMount("cache-volume", "/cache")
			expectedCustomVolumeMounts := []v12.VolumeMount{shmVolumeMount, cacheVolumeMount}

			test.thenStatefulSetsStatesShouldBe(expectedCustomVolumes, expectedCustomVolumeMounts, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(expectedCustomVolumes, expectedCustomVolumeMounts)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.keepCreatedResourcesForNextTest = true

			log.Print("END OF: Test 'GIVEN existing Kubegres is updated by adding new and updating existing custom 'volume.volume' and 'volume.volumeMount'")
		})

		It("GIVEN existing Kubegres is updated with the removal of one custom 'volume.volume' and 'volume.volumeMount' "+
			"THEN StatefulSets should be updated too", func() {

			log.Print("START OF: Test 'GIVEN existing Kubegres is updated with the removal of one custom 'volume.volume' and 'volume.volumeMount'")

			shmVolume := test.givenVolumeWithMemory("dshm", "300Mi")
			customVolumesToRemove := []v12.Volume{shmVolume}

			shmVolumeMount := test.givenVolumeMount("dshm", "/dev/shm")
			customVolumeMountsToRemove := []v12.VolumeMount{shmVolumeMount}

			test.givenVolumesAreRemovedFromTheExistingKubegresSpec(customVolumesToRemove, customVolumeMountsToRemove)

			test.whenKubernetesIsUpdated()

			cacheVolume := test.givenVolumeWithEmptyDir("cache-volume")
			expectedCustomVolumes := []v12.Volume{cacheVolume}

			cacheVolumeMount := test.givenVolumeMount("cache-volume", "/cache")
			expectedCustomVolumeMounts := []v12.VolumeMount{cacheVolumeMount}

			test.thenStatefulSetsStatesShouldBe(expectedCustomVolumes, expectedCustomVolumeMounts, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(expectedCustomVolumes, expectedCustomVolumeMounts)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.keepCreatedResourcesForNextTest = true

			log.Print("END OF: Test 'GIVEN existing Kubegres is updated with the removal of one custom 'volume.volume' and 'volume.volumeMount'")
		})

		It("GIVEN existing Kubegres is updated with the removal of all custom 'volume.volume' and 'volume.volumeMount' "+
			"THEN StatefulSets should be updated too", func() {

			log.Print("START OF: Test 'GIVEN existing Kubegres is updated with the removal of all custom 'volume.volume' and 'volume.volumeMount'")

			cacheVolume := test.givenVolumeWithEmptyDir("cache-volume")
			customVolumesToRemove := []v12.Volume{cacheVolume}

			cacheVolumeMount := test.givenVolumeMount("cache-volume", "/cache")
			customVolumeMountsToRemove := []v12.VolumeMount{cacheVolumeMount}

			test.givenVolumesAreRemovedFromTheExistingKubegresSpec(customVolumesToRemove, customVolumeMountsToRemove)

			test.whenKubernetesIsUpdated()

			expectedCustomVolumes := []v12.Volume{}
			expectedCustomVolumeMounts := []v12.VolumeMount{}
			test.thenStatefulSetsStatesShouldBe(expectedCustomVolumes, expectedCustomVolumeMounts, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(expectedCustomVolumes, expectedCustomVolumeMounts)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN existing Kubegres is updated with the removal of all custom 'volume.volume' and 'volume.volumeMount'")
		})
	})

})

type SpecVolumeAndVolumeMountTest struct {
	keepCreatedResourcesForNextTest bool
	kubegresResource                *postgresv1.Kubegres
	dbQueryTestCases                testcases.DbQueryTestCases
	resourceCreator                 util.TestResourceCreator
	resourceRetriever               util.TestResourceRetriever
}

func (r *SpecVolumeAndVolumeMountTest) givenVolumeWithMemory(volumeName, memoryQuantity string) v12.Volume {

	memQuantity := resource.MustParse(memoryQuantity)

	return v12.Volume{
		Name: volumeName,
		VolumeSource: v12.VolumeSource{
			EmptyDir: &v12.EmptyDirVolumeSource{
				Medium:    v12.StorageMediumMemory,
				SizeLimit: &memQuantity,
			},
		},
	}
}

func (r *SpecVolumeAndVolumeMountTest) givenVolumeWithEmptyDir(volumeName string) v12.Volume {

	return v12.Volume{
		Name: volumeName,
		VolumeSource: v12.VolumeSource{
			EmptyDir: &v12.EmptyDirVolumeSource{},
		},
	}
}

func (r *SpecVolumeAndVolumeMountTest) givenVolumeMount(volumeName, mountPath string) v12.VolumeMount {

	return v12.VolumeMount{
		Name:      volumeName,
		MountPath: mountPath,
	}
}

func (r *SpecVolumeAndVolumeMountTest) givenNewKubegresSpecIsSetTo(
	customVolumes []v12.Volume,
	customVolumeMounts []v12.VolumeMount,
	specNbreReplicas int32) {

	r.kubegresResource = resourceConfigs.LoadKubegresYaml()
	r.kubegresResource.Spec.Volume.Volumes = append(r.kubegresResource.Spec.Volume.Volumes, customVolumes...)
	r.kubegresResource.Spec.Volume.VolumeMounts = append(r.kubegresResource.Spec.Volume.VolumeMounts, customVolumeMounts...)
	r.kubegresResource.Spec.Replicas = &specNbreReplicas
}

func (r *SpecVolumeAndVolumeMountTest) givenVolumesAreUpdatedOrAddedToTheExistingKubegresSpec(
	customVolumesToAddOrReplace []v12.Volume,
	customVolumeMountsToAddOrReplace []v12.VolumeMount) {

	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()
	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	for _, customVolume := range customVolumesToAddOrReplace {
		volumeIndex := r.getVolumeIndex(customVolume)
		if volumeIndex >= 0 {
			r.kubegresResource.Spec.Volume.Volumes[volumeIndex] = customVolume
		} else {
			r.kubegresResource.Spec.Volume.Volumes = append(r.kubegresResource.Spec.Volume.Volumes, customVolume)
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

func (r *SpecVolumeAndVolumeMountTest) givenVolumesAreRemovedFromTheExistingKubegresSpec(
	customVolumesToRemove []v12.Volume,
	customVolumeMountsToRemove []v12.VolumeMount) {

	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()
	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	for _, customVolume := range customVolumesToRemove {
		volumeIndex := r.getVolumeIndex(customVolume)
		if volumeIndex >= 0 {
			r.kubegresResource.Spec.Volume.Volumes = append(r.kubegresResource.Spec.Volume.Volumes[:volumeIndex], r.kubegresResource.Spec.Volume.Volumes[volumeIndex+1:]...)
		}
	}

	for _, customVolumeMount := range customVolumeMountsToRemove {
		volumeMountIndex := r.getVolumeMountIndex(customVolumeMount)
		if volumeMountIndex >= 0 {
			r.kubegresResource.Spec.Volume.VolumeMounts = append(r.kubegresResource.Spec.Volume.VolumeMounts[:volumeMountIndex], r.kubegresResource.Spec.Volume.VolumeMounts[volumeMountIndex+1:]...)
		}
	}
}

func (r *SpecVolumeAndVolumeMountTest) getVolumeIndex(customVolume v12.Volume) int {
	index := 0
	for _, volume := range r.kubegresResource.Spec.Volume.Volumes {
		if customVolume.Name == volume.Name {
			return index
		}
		index++
	}
	return -1
}

func (r *SpecVolumeAndVolumeMountTest) getVolumeMountIndex(customVolumeMount v12.VolumeMount) int {
	index := 0
	for _, volumeMount := range r.kubegresResource.Spec.Volume.VolumeMounts {
		if customVolumeMount.Name == volumeMount.Name {
			return index
		}
		index++
	}
	return -1
}

func (r *SpecVolumeAndVolumeMountTest) whenKubegresIsCreated() {
	r.resourceCreator.CreateKubegres(r.kubegresResource)
}

func (r *SpecVolumeAndVolumeMountTest) whenKubernetesIsUpdated() {
	r.resourceCreator.UpdateResource(r.kubegresResource, "Kubegres")
}

func (r *SpecVolumeAndVolumeMountTest) thenErrorEventShouldBeLoggedAboutVolumeName() {
	expectedErrorEvent := util.EventRecord{
		Eventtype: v12.EventTypeWarning,
		Reason:    "SpecCheckErr",
		Message: "In the Resources Spec the value of 'spec.Volume.Volumes' has an entry with a volume name " +
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

func (r *SpecVolumeAndVolumeMountTest) thenErrorEventShouldBeLoggedAboutVolumeMountName() {
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

func (r *SpecVolumeAndVolumeMountTest) thenErrorEventShouldBeLoggedAboutVolumeMountPath() {
	expectedErrorEvent := util.EventRecord{
		Eventtype: v12.EventTypeWarning,
		Reason:    "SpecCheckErr",
		Message: "In the Resources Spec the value of 'spec.Volume.VolumeMounts' has an entry with a 'mountPath' value " +
			"which is reserved for the Postgres database: " + r.kubegresResource.Spec.Database.VolumeMount + " . " +
			"That value cannot be used and it is reserved for Kubegres internal usages. Please change that value in the YAML.",
	}
	Eventually(func() bool {
		_, err := r.resourceRetriever.GetKubegres()
		if err != nil {
			return false
		}
		return eventRecorderTest.CheckEventExist(expectedErrorEvent)

	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *SpecVolumeAndVolumeMountTest) thenStatefulSetsStatesShouldBe(
	expectedCustomVolumes []v12.Volume,
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

			for _, customVolume := range expectedCustomVolumes {
				if !r.doesCustomVolumeExistsInStatefulSet(customVolume, resource.StatefulSet.Spec.Template.Spec.Volumes) {
					log.Println("StatefulSet '" + resource.StatefulSet.Name + "' doesn't have the expected custom volume with name: '" + customVolume.Name + "'. Waiting...")
					return false
				}
			}

			for _, volumeInStatefulSet := range resource.StatefulSet.Spec.Template.Spec.Volumes {
				if r.isCustomVolume(volumeInStatefulSet, kubegresContext) &&
					!r.isVolumeAnExpectedCustomVolume(volumeInStatefulSet, expectedCustomVolumes) {
					log.Println("StatefulSet '" + resource.StatefulSet.Name + "' still has custom volume with name: '" + volumeInStatefulSet.Name + "'. Waiting...")
					return false
				}
			}

			for _, customVolumeMount := range expectedCustomVolumeMounts {
				if !r.doesCustomVolumeMountExistsInStatefulSet(customVolumeMount, resource.StatefulSet.Spec.Template.Spec.Containers[0].VolumeMounts) {
					log.Println("StatefulSet '" + resource.StatefulSet.Name + "' doesn't have the expected custom volumeMount with name: '" + customVolumeMount.Name + "'. Waiting...")
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

func (r *SpecVolumeAndVolumeMountTest) thenDeployedKubegresSpecShouldBeSetTo(
	expectedCustomVolumes []v12.Volume,
	expectedCustomVolumeMounts []v12.VolumeMount) {

	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	if len(expectedCustomVolumes) == 0 && len(expectedCustomVolumeMounts) == 0 {
		return
	}

	Expect(r.kubegresResource.Spec.Volume.Volumes).Should(Equal(expectedCustomVolumes))
	Expect(r.kubegresResource.Spec.Volume.VolumeMounts).Should(Equal(expectedCustomVolumeMounts))
}

func (r *SpecVolumeAndVolumeMountTest) isCustomVolume(volume v12.Volume, kubegresContext ctx.KubegresContext) bool {
	return !kubegresContext.IsReservedVolumeName(volume.Name)
}

func (r *SpecVolumeAndVolumeMountTest) isCustomVolumeMount(volumeMount v12.VolumeMount, kubegresContext ctx.KubegresContext) bool {
	return !kubegresContext.IsReservedVolumeName(volumeMount.Name)
}

func (r *SpecVolumeAndVolumeMountTest) doesCustomVolumeExistsInStatefulSet(customVolume v12.Volume, statefulSetVolumes []v12.Volume) bool {
	for _, statefulSetVolume := range statefulSetVolumes {
		if reflect.DeepEqual(statefulSetVolume, customVolume) {
			return true
		}
	}
	return false
}

func (r *SpecVolumeAndVolumeMountTest) isVolumeAnExpectedCustomVolume(volumeToCheck v12.Volume, expectedCustomVolumes []v12.Volume) bool {
	for _, expectedCustomVolume := range expectedCustomVolumes {
		if reflect.DeepEqual(expectedCustomVolume, volumeToCheck) {
			return true
		}
	}
	return false
}

func (r *SpecVolumeAndVolumeMountTest) isVolumeMountAnExpectedCustomVolumeMount(volumeMountToCheck v12.VolumeMount, expectedCustomVolumeMounts []v12.VolumeMount) bool {
	for _, expectedCustomVolumeMount := range expectedCustomVolumeMounts {
		if reflect.DeepEqual(expectedCustomVolumeMount, volumeMountToCheck) {
			return true
		}
	}
	return false
}

func (r *SpecVolumeAndVolumeMountTest) doesCustomVolumeMountExistsInStatefulSet(customVolumeMount v12.VolumeMount, statefulSetVolumeMounts []v12.VolumeMount) bool {
	for _, statefulSetVolumeMount := range statefulSetVolumeMounts {
		if reflect.DeepEqual(statefulSetVolumeMount, customVolumeMount) {
			return true
		}
	}
	return false
}
