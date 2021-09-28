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
	"log"
	postgresv1 "reactive-tech.io/kubegres/api/v1"
	"reactive-tech.io/kubegres/controllers/ctx"
	"reactive-tech.io/kubegres/test/resourceConfigs"
	"reactive-tech.io/kubegres/test/util"
	"reactive-tech.io/kubegres/test/util/testcases"
	"time"
)

var _ = Describe("Setting Kubegres specs 'database.volumeMount'", func() {

	var test = SpecDatabaseVolumeMountTest{}

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

	Context("GIVEN new Kubegres is created without spec 'database.volumeMount' and with spec 'replica' set to 3", func() {

		It("THEN 1 primary and 2 replica should be created with 'database.volumeMount' set to the value of the const 'KubegresContext.DefaultDatabaseVolumeMount' and a normal event should be logged", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created without spec 'database.volumeMount' and with spec 'replica' set to 3'")

			test.givenNewKubegresSpecIsSetTo("", 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(ctx.DefaultDatabaseVolumeMount, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(ctx.DefaultDatabaseVolumeMount)

			test.thenEventShouldBeLoggedSayingDatabaseVolumeMountWasSetToDefaultValue()

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN new Kubegres is created without spec 'database.volumeMount' and with spec 'replica' set to 3'")
		})
	})

	Context("GIVEN new Kubegres is created with spec 'database.volumeMount' set to '/tmp/folder1' and spec 'replica' set to 3 and later 'database.volumeMount' is updated to '/tmp/folder2'", func() {

		It("GIVEN new Kubegres is created with spec 'database.volumeMount' set to '/tmp/folder1' and spec 'replica' set to 3 THEN 1 primary and 2 replica should be created with spec 'database.volumeMount' set to '/tmp/folder1'", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'database.volumeMount' set to '/tmp/folder1' and spec 'replica' set to 3")

			test.givenNewKubegresSpecIsSetTo("/tmp/folder1", 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe("/tmp/folder1", 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo("/tmp/folder1")

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.keepCreatedResourcesForNextTest = true

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'database.volumeMount' set to '/tmp/folder1' and spec 'replica' set to 3'")
		})

		It("GIVEN existing Kubegres is updated with spec 'database.volumeMount' set from '/tmp/folder1' to '/tmp/folder2' THEN an error event should be logged", func() {

			log.Print("START OF: Test 'GIVEN existing Kubegres is updated with spec 'database.volumeMount' set from '/tmp/folder1' to '/tmp/folder2'")

			test.givenExistingKubegresSpecIsSetTo("/tmp/folder2")

			test.whenKubernetesIsUpdated()

			test.thenErrorEventShouldBeLoggedSayingCannotChangeDatabaseVolumeMount("/tmp/folder1", "/tmp/folder2")

			test.thenPodsStatesShouldBe("/tmp/folder1", 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo("/tmp/folder1")

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN existing Kubegres is updated with spec 'database.volumeMount' set from '/tmp/folder1' to '/tmp/folder2'")
		})
	})

})

type SpecDatabaseVolumeMountTest struct {
	keepCreatedResourcesForNextTest bool
	kubegresResource                *postgresv1.Kubegres
	dbQueryTestCases                testcases.DbQueryTestCases
	resourceCreator                 util.TestResourceCreator
	resourceRetriever               util.TestResourceRetriever
}

func (r *SpecDatabaseVolumeMountTest) givenNewKubegresSpecIsSetTo(databaseVolumeMount string, specNbreReplicas int32) {
	r.kubegresResource = resourceConfigs.LoadKubegresYaml()
	r.kubegresResource.Spec.Database.VolumeMount = databaseVolumeMount
	r.kubegresResource.Spec.Replicas = &specNbreReplicas
}

func (r *SpecDatabaseVolumeMountTest) givenExistingKubegresSpecIsSetTo(specDatabaseVolumeMount string) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	r.kubegresResource.Spec.Database.VolumeMount = specDatabaseVolumeMount
}

func (r *SpecDatabaseVolumeMountTest) whenKubegresIsCreated() {
	r.resourceCreator.CreateKubegres(r.kubegresResource)
}

func (r *SpecDatabaseVolumeMountTest) whenKubernetesIsUpdated() {
	r.resourceCreator.UpdateResource(r.kubegresResource, "Kubegres")
}

func (r *SpecDatabaseVolumeMountTest) thenEventShouldBeLoggedSayingDatabaseVolumeMountWasSetToDefaultValue() {

	expectedErrorEvent := util.EventRecord{
		Eventtype: v12.EventTypeNormal,
		Reason:    "DefaultSpecValue",
		Message:   "A default value was set for a field in Kubegres YAML spec. 'spec.database.volumeMount': New value: " + ctx.DefaultDatabaseVolumeMount,
	}
	Eventually(func() bool {
		_, err := r.resourceRetriever.GetKubegres()
		if err != nil {
			return false
		}
		return eventRecorderTest.CheckEventExist(expectedErrorEvent)

	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *SpecDatabaseVolumeMountTest) thenPodsStatesShouldBe(databaseVolumeMount string, nbrePrimary, nbreReplicas int) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		for _, resource := range kubegresResources.Resources {
			currentDatabaseVolumeMount := resource.Pod.Spec.Containers[0].VolumeMounts[0].MountPath
			if currentDatabaseVolumeMount != databaseVolumeMount {
				log.Println("Pod '" + resource.Pod.Name + "' doesn't have the expected database.volumeMount: '" + databaseVolumeMount + "'. " +
					"Current value: '" + currentDatabaseVolumeMount + "'. Waiting...")
				return false
			}
		}

		if kubegresResources.AreAllReady &&
			kubegresResources.NbreDeployedPrimary == nbrePrimary &&
			kubegresResources.NbreDeployedReplicas == nbreReplicas {

			time.Sleep(resourceConfigs.TestRetryInterval)
			log.Println("Deployed and Ready StatefulSets check successful")
			return true
		}

		return false

	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *SpecDatabaseVolumeMountTest) thenDeployedKubegresSpecShouldBeSetTo(databaseVolumeMount string) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	Expect(r.kubegresResource.Spec.Database.VolumeMount).Should(Equal(databaseVolumeMount))
}

func (r *SpecDatabaseVolumeMountTest) thenErrorEventShouldBeLoggedSayingCannotChangeDatabaseVolumeMount(currentValue, newValue string) {
	expectedErrorEvent := util.EventRecord{
		Eventtype: v12.EventTypeWarning,
		Reason:    "SpecCheckErr",
		Message: "In the Resources Spec the value of 'spec.database.volumeMount' cannot be changed from '" + currentValue + "' to '" + newValue + "' after Pods were created. " +
			"Otherwise, the cluster of PostgreSql servers risk of being inconsistent. " +
			"We roll-backed Kubegres spec to the currently working value '" + currentValue + "'. " +
			"If you know what you are doing, you can manually update that spec in every StatefulSet of your PostgreSql cluster and then Kubegres will automatically update itself.",
	}
	Eventually(func() bool {
		_, err := r.resourceRetriever.GetKubegres()
		if err != nil {
			return false
		}
		return eventRecorderTest.CheckEventExist(expectedErrorEvent)

	}, time.Second*10, time.Second*5).Should(BeTrue())
}
