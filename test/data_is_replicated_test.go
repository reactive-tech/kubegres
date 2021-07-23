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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"log"
	postgresv1 "reactive-tech.io/kubegres/api/v1"
	"reactive-tech.io/kubegres/test/resourceConfigs"
	"reactive-tech.io/kubegres/test/util"
	"strconv"
	"time"
)

var _ = Describe("Checking changes in Primary DB is replicated in Replica DBs", func() {

	var test = DataIsReplicatedTest{}

	BeforeEach(func() {
		//Skip("Temporarily skipping test")

		namespace := resourceConfigs.DefaultNamespace
		test.resourceRetriever = util.CreateTestResourceRetriever(k8sClientTest, namespace)
		test.resourceCreator = util.CreateTestResourceCreator(k8sClientTest, test.resourceRetriever, namespace)
		test.connectionPrimaryDb = util.InitDbConnectionDbUtil(test.resourceCreator, resourceConfigs.KubegresResourceName, resourceConfigs.ServiceToSqlQueryPrimaryDbNodePort, true)
		test.connectionReplicaDb = util.InitDbConnectionDbUtil(test.resourceCreator, resourceConfigs.KubegresResourceName, resourceConfigs.ServiceToSqlQueryReplicaDbNodePort, false)
	})

	AfterEach(func() {
		test.resourceCreator.DeleteAllTestResources()
	})

	Context("GIVEN new Kubegres is created with spec 'replica' set to 3 AND data is inserted and then deleted from Primary DB", func() {

		It("THEN 1 primary and 2 replica should be created with inserted and deleted users should be replicated from Primary DB to Replica DB", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'replica' set to 3 AND data is inserted and then deleted from Primary DB'")

			test.givenNewKubegresSpecIsSetTo(3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 2)

			expectedNbreUsers := 0

			test.GivenUserAddedInPrimaryDb()
			expectedNbreUsers++

			test.GivenUserAddedInPrimaryDb()
			expectedNbreUsers++

			test.GivenUserAddedInPrimaryDb()
			expectedNbreUsers++

			test.ThenPrimaryDbContainsExpectedNbreUsers(expectedNbreUsers)
			test.ThenReplicaDbContainsExpectedNbreUsers(expectedNbreUsers)

			test.GivenUserDeletedFromPrimaryDb()
			expectedNbreUsers--

			test.ThenPrimaryDbContainsExpectedNbreUsers(expectedNbreUsers)
			test.ThenReplicaDbContainsExpectedNbreUsers(expectedNbreUsers)

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'replica' set to 3 AND data is inserted and then deleted from Primary DB'")
		})
	})

	Context("GIVEN existing Kubegres with 'replica' set to 1, with data in PrimaryDB is updated from replica 1 to 4", func() {

		It("THEN 1 primary and 3 replica should be created with all users present on all Replica DB", func() {

			log.Print("START OF: Test 'GIVEN existing Kubegres with 'replica' set to 1, with data in PrimaryDB is updated from replica 1 to 3'")

			test.givenNewKubegresSpecIsSetTo(1)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 0)

			expectedNbreUsers := 0

			test.GivenUserAddedInPrimaryDb()
			expectedNbreUsers++

			test.GivenUserAddedInPrimaryDb()
			expectedNbreUsers++

			test.givenExistingKubegresSpecIsSetTo(4)

			test.whenKubernetesIsUpdated()

			test.thenPodsStatesShouldBe(1, 3)

			test.ThenPrimaryDbContainsExpectedNbreUsers(expectedNbreUsers)
			test.ThenReplicaDbContainsExpectedNbreUsers(expectedNbreUsers)

			test.GivenUserAddedInPrimaryDb()
			expectedNbreUsers++

			test.GivenUserAddedInPrimaryDb()
			expectedNbreUsers++

			test.GivenUserAddedInPrimaryDb()
			expectedNbreUsers++

			test.GivenUserDeletedFromPrimaryDb()
			expectedNbreUsers--

			test.ThenPrimaryDbContainsExpectedNbreUsers(expectedNbreUsers)
			test.ThenReplicaDbContainsExpectedNbreUsers(expectedNbreUsers)

			log.Print("END OF: Test 'GIVEN existing Kubegres with 'replica' set to 1, with data in PrimaryDB is updated from replica 1 to 3'")
		})
	})
})

type DataIsReplicatedTest struct {
	kubegresResource      *postgresv1.Kubegres
	connectionPrimaryDb   util.DbConnectionDbUtil
	connectionReplicaDb   util.DbConnectionDbUtil
	resourceCreator       util.TestResourceCreator
	resourceRetriever     util.TestResourceRetriever
	customEnvVariableName string
	customEnvVariableKey  string
}

func (r *DataIsReplicatedTest) givenNewKubegresSpecIsSetTo(specNbreReplicas int32) {
	r.kubegresResource = resourceConfigs.LoadKubegresYaml()
	r.kubegresResource.Spec.Replicas = &specNbreReplicas
}

func (r *DataIsReplicatedTest) givenExistingKubegresSpecIsSetTo(specNbreReplicas int32) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	r.kubegresResource.Spec.Replicas = &specNbreReplicas
}

func (r *DataIsReplicatedTest) whenKubegresIsCreated() {
	r.resourceCreator.CreateKubegres(r.kubegresResource)
}

func (r *DataIsReplicatedTest) whenKubernetesIsUpdated() {
	r.resourceCreator.UpdateResource(r.kubegresResource, "Kubegres")
}

func (r *DataIsReplicatedTest) thenPodsStatesShouldBe(nbrePrimary, nbreReplicas int) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
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

func (r *DataIsReplicatedTest) GivenUserAddedInPrimaryDb() {
	Eventually(func() bool {
		return r.connectionPrimaryDb.InsertUser()
	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *DataIsReplicatedTest) GivenUserDeletedFromPrimaryDb() {
	Eventually(func() bool {
		return r.connectionPrimaryDb.DeleteUser(r.connectionPrimaryDb.LastInsertedUserId)
	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *DataIsReplicatedTest) ThenPrimaryDbContainsExpectedNbreUsers(expectedNbreUsers int) {
	Eventually(func() bool {

		users := r.connectionPrimaryDb.GetUsers()
		r.connectionPrimaryDb.Close()

		if len(users) != expectedNbreUsers ||
			r.connectionPrimaryDb.NbreInsertedUsers != expectedNbreUsers {
			log.Println("Primary DB does not contain the expected number of users: " + strconv.Itoa(expectedNbreUsers))
			return false
		}

		return true

	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *DataIsReplicatedTest) ThenReplicaDbContainsExpectedNbreUsers(expectedNbreUsers int) {
	Eventually(func() bool {

		users := r.connectionReplicaDb.GetUsers()
		r.connectionReplicaDb.Close()

		if len(users) != expectedNbreUsers ||
			r.connectionReplicaDb.NbreInsertedUsers != expectedNbreUsers {
			log.Println("Replica DB does not contain the expected number of users: " + strconv.Itoa(expectedNbreUsers))
			return false
		}

		return true

	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}
