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

package test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v12 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"log"
	postgresv1 "reactive-tech.io/kubegres/api/v1"
	resourceConfigs2 "reactive-tech.io/kubegres/internal/test/resourceConfigs"
	util2 "reactive-tech.io/kubegres/internal/test/util"
	"reactive-tech.io/kubegres/internal/test/util/testcases"
	"time"
)

var _ = Describe("Setting Kubegres specs 'database.size'", func() {

	var test = SpecDatabaseSizeTest{}

	BeforeEach(func() {
		//Skip("Temporarily skipping test")

		namespace := resourceConfigs2.DefaultNamespace
		test.resourceRetriever = util2.CreateTestResourceRetriever(k8sClientTest, namespace)
		test.resourceCreator = util2.CreateTestResourceCreator(k8sClientTest, test.resourceRetriever, namespace)
		test.dbQueryTestCases = testcases.InitDbQueryTestCases(test.resourceCreator, resourceConfigs2.KubegresResourceName)
	})

	AfterEach(func() {
		if !test.keepCreatedResourcesForNextTest {
			test.resourceCreator.DeleteAllTestResources()
		} else {
			test.keepCreatedResourcesForNextTest = false
		}
	})

	Context("GIVEN new Kubegres is created without spec 'database.size'", func() {

		It("THEN an error event should be logged", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created without spec 'database.size''")

			test.givenNewKubegresSpecIsSetTo("", 3)

			test.whenKubegresIsCreated()

			test.thenErrorEventShouldBeLogged()

			log.Print("END OF: Test 'GIVEN new Kubegres is created without spec 'database.size'")
		})
	})

	Context("GIVEN new Kubegres is created with spec 'database.size' set to '300Mi' and spec 'replica' set to 3 and later 'database.size' is updated to '400Mi'", func() {

		It("GIVEN new Kubegres is created with spec 'database.size' set to '300Mi' and spec 'replica' set to 3 THEN 1 primary and 2 replica should be created with spec 'database.size' set to '300Mi'", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'database.size' set to '300Mi' and spec 'replica' set to 3")

			test.givenNewKubegresSpecIsSetTo("300Mi", 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe("300Mi", 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo("300Mi")

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.keepCreatedResourcesForNextTest = true

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'database.size' set to '300Mi' and spec 'replica' set to 3'")
		})

		It("GIVEN existing Kubegres is updated with spec 'database.size' set from '300Mi' to '400Mi' THEN an error event should be logged", func() {

			log.Print("START OF: Test 'GIVEN existing Kubegres is updated with spec 'database.size' set from '300Mi' to '400Mi'")

			test.givenExistingKubegresSpecIsSetTo("400Mi")

			test.whenKubernetesIsUpdated()

			test.thenErrorEventShouldBeLoggedSayingCannotChangeStorageSize("300Mi", "400Mi")

			test.thenPodsStatesShouldBe("300Mi", 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo("300Mi")

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN existing Kubegres is updated with spec 'database.size' set from '300Mi' to '400Mi'")
		})
	})

})

type SpecDatabaseSizeTest struct {
	keepCreatedResourcesForNextTest bool
	kubegresResource                *postgresv1.Kubegres
	dbQueryTestCases                testcases.DbQueryTestCases
	resourceCreator                 util2.TestResourceCreator
	resourceRetriever               util2.TestResourceRetriever
}

func (r *SpecDatabaseSizeTest) givenNewKubegresSpecIsSetTo(databaseSize string, specNbreReplicas int32) {
	r.kubegresResource = resourceConfigs2.LoadKubegresYaml()
	r.kubegresResource.Spec.Database.Size = databaseSize
	r.kubegresResource.Spec.Replicas = &specNbreReplicas
}

func (r *SpecDatabaseSizeTest) givenExistingKubegresSpecIsSetTo(databaseSize string) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	r.kubegresResource.Spec.Database.Size = databaseSize
}

func (r *SpecDatabaseSizeTest) whenKubegresIsCreated() {
	r.resourceCreator.CreateKubegres(r.kubegresResource)
}

func (r *SpecDatabaseSizeTest) whenKubernetesIsUpdated() {
	r.resourceCreator.UpdateResource(r.kubegresResource, "Kubegres")
}

func (r *SpecDatabaseSizeTest) thenErrorEventShouldBeLogged() {
	expectedErrorEvent := util2.EventRecord{
		Eventtype: v12.EventTypeWarning,
		Reason:    "SpecCheckErr",
		Message:   "In the Resources Spec the value of 'spec.database.size' is undefined. Please set a value otherwise this operator cannot work correctly.",
	}
	Eventually(func() bool {
		_, err := r.resourceRetriever.GetKubegres()
		if err != nil {
			return false
		}
		return eventRecorderTest.CheckEventExist(expectedErrorEvent)

	}, resourceConfigs2.TestTimeout, resourceConfigs2.TestRetryInterval).Should(BeTrue())
}

func (r *SpecDatabaseSizeTest) thenPodsStatesShouldBe(databaseSize string, nbrePrimary, nbreReplicas int) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		for _, resource := range kubegresResources.Resources {
			currentDatabaseSize := resource.Pvc.Spec.Resources.Requests["storage"]
			if currentDatabaseSize.String() != databaseSize {
				log.Println("Pod '" + resource.Pod.Name + "' doesn't have the expected database.size: '" + databaseSize + "'. " +
					"Current value: '" + currentDatabaseSize.String() + "'. Waiting...")
				return false
			}
		}

		if kubegresResources.AreAllReady &&
			kubegresResources.NbreDeployedPrimary == nbrePrimary &&
			kubegresResources.NbreDeployedReplicas == nbreReplicas {

			time.Sleep(resourceConfigs2.TestRetryInterval)
			log.Println("Deployed and Ready StatefulSets check successful")
			return true
		}

		return false

	}, resourceConfigs2.TestTimeout, resourceConfigs2.TestRetryInterval).Should(BeTrue())
}

func (r *SpecDatabaseSizeTest) thenDeployedKubegresSpecShouldBeSetTo(databaseSize string) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	Expect(r.kubegresResource.Spec.Database.Size).Should(Equal(databaseSize))
}

func (r *SpecDatabaseSizeTest) thenErrorEventShouldBeLoggedSayingCannotChangeStorageSize(currentValue, newValue string) {
	expectedErrorEvent := util2.EventRecord{
		Eventtype: v12.EventTypeWarning,
		Reason:    "SpecCheckErr",
		Message: "In the Resources Spec the value of 'spec.database.size' cannot be changed from '" + currentValue + "' to '" + newValue + "' after Pods were created. " +
			"The StorageClass does not allow volume expansion. The option AllowVolumeExpansion is set to false. " +
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
