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
	"reactive-tech.io/kubegres/test/resourceConfigs"
	"reactive-tech.io/kubegres/test/util"
	"reactive-tech.io/kubegres/test/util/testcases"
	"time"
)

var _ = Describe("Setting Kubegres specs 'database.storageClassName'", func() {

	var test = SpecDatabaseStorageClassTest{}

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

	Context("GIVEN new Kubegres is created without spec 'database.storageClassName'", func() {

		It("THEN it should be set to the 'standard' storage class (which is the default one in our test cluster)", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created without spec 'database.storageClassName''")

			test.givenNewKubegresSpecIsSetTo(nil, 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe("standard", 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo("standard")

			test.thenEventShouldBeLoggedSayingStorageClassNameWasSetToDefaultValue()

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.keepCreatedResourcesForNextTest = false

			log.Print("END OF: Test 'GIVEN new Kubegres is created without spec 'database.storageClassName'")
		})
	})

	Context("GIVEN new Kubegres is created with spec 'database.storageClassName' set to 'standard' and spec 'replica' set to 3 and later 'database.storageClassName' is updated to 'anything'", func() {

		It("GIVEN new Kubegres is created with spec 'database.storageClassName' set to 'standard' and spec 'replica' set to 3 THEN 1 primary and 2 replica should be created with spec 'database.storageClassName' set to 'standard'", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'database.storageClassName' set to 'standard' and spec 'replica' set to 3")

			customStorageClassName := "standard"

			test.givenNewKubegresSpecIsSetTo(&customStorageClassName, 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(customStorageClassName, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(customStorageClassName)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.keepCreatedResourcesForNextTest = true

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'database.storageClassName' set to 'standard' and spec 'replica' set to 3'")
		})

		It("GIVEN existing Kubegres is updated with spec 'database.storageClassName' set from 'standard' to 'anything' THEN an error event should be logged", func() {

			log.Print("START OF: Test 'GIVEN existing Kubegres is updated with spec 'database.storageClassName' set from 'standard' to 'anything'")

			test.givenExistingKubegresSpecIsSetTo("anything")

			test.whenKubernetesIsUpdated()

			test.thenErrorEventShouldBeLoggedSayingCannotChangeStorageClassName("standard", "anything")

			test.thenPodsStatesShouldBe("standard", 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo("standard")

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN existing Kubegres is updated with spec 'database.storageClassName' set from 'standard' to 'anything'")
		})
	})

})

type SpecDatabaseStorageClassTest struct {
	keepCreatedResourcesForNextTest bool
	kubegresResource                *postgresv1.Kubegres
	dbQueryTestCases                testcases.DbQueryTestCases
	resourceCreator                 util.TestResourceCreator
	resourceRetriever               util.TestResourceRetriever
}

func (r *SpecDatabaseStorageClassTest) givenNewKubegresSpecIsSetTo(databaseStorageClassName *string, specNbreReplicas int32) {
	r.kubegresResource = resourceConfigs.LoadKubegresYaml()
	r.kubegresResource.Spec.Database.StorageClassName = databaseStorageClassName
	r.kubegresResource.Spec.Replicas = &specNbreReplicas
}

func (r *SpecDatabaseStorageClassTest) givenExistingKubegresSpecIsSetTo(databaseStorageClassName string) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	r.kubegresResource.Spec.Database.StorageClassName = &databaseStorageClassName
}

func (r *SpecDatabaseStorageClassTest) whenKubegresIsCreated() {
	r.resourceCreator.CreateKubegres(r.kubegresResource)
}

func (r *SpecDatabaseStorageClassTest) whenKubernetesIsUpdated() {
	r.resourceCreator.UpdateResource(r.kubegresResource, "Kubegres")
}

func (r *SpecDatabaseStorageClassTest) thenErrorEventShouldBeLogged() {
	expectedErrorEvent := util.EventRecord{
		Eventtype: v12.EventTypeWarning,
		Reason:    "SpecCheckErr",
		Message:   "In the Resources Spec the value of 'spec.database.storageClassName' is undefined. Please set a value otherwise this operator cannot work correctly.",
	}
	Eventually(func() bool {
		_, err := r.resourceRetriever.GetKubegres()
		if err != nil {
			return false
		}
		return eventRecorderTest.CheckEventExist(expectedErrorEvent)

	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *SpecDatabaseStorageClassTest) thenPodsStatesShouldBe(databaseStorageClassName string, nbrePrimary, nbreReplicas int) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		for _, resource := range kubegresResources.Resources {
			currentDatabaseStorageClassName := resource.StatefulSet.Spec.VolumeClaimTemplates[0].Spec.StorageClassName
			if *currentDatabaseStorageClassName != databaseStorageClassName {
				log.Println("Pod '" + resource.Pod.Name + "' doesn't have the expected database.storageClassName: '" + databaseStorageClassName + "'. " +
					"Current value: '" + *currentDatabaseStorageClassName + "'. Waiting...")
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

func (r *SpecDatabaseStorageClassTest) thenDeployedKubegresSpecShouldBeSetTo(databaseStorageClassName string) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	Expect(*r.kubegresResource.Spec.Database.StorageClassName).Should(Equal(databaseStorageClassName))
}

func (r *SpecDatabaseStorageClassTest) thenEventShouldBeLoggedSayingStorageClassNameWasSetToDefaultValue() {

	expectedErrorEvent := util.EventRecord{
		Eventtype: v12.EventTypeNormal,
		Reason:    "DefaultSpecValue",
		Message:   "A default value was set for a field in Kubegres YAML spec. 'spec.Database.StorageClassName': New value: standard",
	}
	Eventually(func() bool {
		_, err := r.resourceRetriever.GetKubegres()
		if err != nil {
			return false
		}
		return eventRecorderTest.CheckEventExist(expectedErrorEvent)

	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *SpecDatabaseStorageClassTest) thenErrorEventShouldBeLoggedSayingCannotChangeStorageClassName(currentValue, newValue string) {
	expectedErrorEvent := util.EventRecord{
		Eventtype: v12.EventTypeWarning,
		Reason:    "SpecCheckErr",
		Message: "In the Resources Spec the value of 'spec.database.storageClassName' cannot be changed from '" + currentValue + "' to '" + newValue + "' after Pods were created. " +
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
