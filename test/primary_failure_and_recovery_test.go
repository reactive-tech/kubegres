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

var _ = Describe("Primary instances is not available, checking recovery works", func() {

	var test = PrimaryFailureAndRecoveryTest{}

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

	Context("GIVEN Kubegres with 1 primary AND that primary is deleted while its associated PVC is still available", func() {

		It("THEN the primary should be automatically re-created and teh existing PVC should be associated to the new primary", func() {

			log.Print("START OF: Test 'GIVEN Kubegres with 1 primary AND that primary is deleted while its associated PVC is still available'")

			test.givenNewKubegresSpecIsSetTo(1)

			test.whenKubernetesIsCreated()

			test.thenPodsStatesShouldBe(1, 0)

			expectedNbreUsers := 0

			test.GivenUserAddedInPrimaryDb()
			expectedNbreUsers++

			test.GivenUserAddedInPrimaryDb()
			expectedNbreUsers++

			test.whenPrimaryIsDeleted()

			test.thenPodsStatesShouldBe(1, 0)

			test.ThenPrimaryDbContainsExpectedNbreUsers(expectedNbreUsers)

			log.Print("END OF: Test 'GIVEN Kubegres with 1 primary AND that primary is deleted while its associated PVC is still available'")
		})
	})

	Context("GIVEN Kubegres with 1 primary AND that primary is deleted including its associated PVC", func() {

		It("THEN the primary should be automatically re-created and a new PVC should be associated to the new primary", func() {

			log.Print("START OF: Test 'GIVEN Kubegres with 1 primary AND that primary is deleted including its associated PVC'")

			test.givenNewKubegresSpecIsSetTo(1)

			test.whenKubernetesIsCreated()

			test.thenPodsStatesShouldBe(1, 0)

			expectedNbreUsers := 0

			test.GivenUserAddedInPrimaryDb()
			expectedNbreUsers++

			test.GivenUserAddedInPrimaryDb()
			expectedNbreUsers++

			test.whenPrimaryAndItsPVCAreDeleted()

			test.thenPodsStatesShouldBe(1, 0)

			expectedNbreUsers = 0
			test.ThenPrimaryDbContainsExpectedNbreUsers(expectedNbreUsers)

			test.GivenUserAddedInPrimaryDb()
			expectedNbreUsers++

			test.ThenPrimaryDbContainsExpectedNbreUsers(expectedNbreUsers)

			log.Print("END OF: Test 'GIVEN Kubegres with 1 primary AND that primary is deleted including its associated PVC'")
		})
	})

	Context("GIVEN Kubegres with 1 primary and 2 replicas AND primary is deleted", func() {

		It("THEN the failover should take place with a replica becoming primary AND a new replica created AND existing data available", func() {

			log.Print("START OF: Test 'GIVEN Kubegres with 1 primary and 2 replicas AND primary is deleted'")

			test.givenNewKubegresSpecIsSetTo(3)

			test.whenKubernetesIsCreated()

			test.thenPodsStatesShouldBe(1, 2)

			expectedNbreUsers := 0

			test.GivenUserAddedInPrimaryDb()
			expectedNbreUsers++

			test.GivenUserAddedInPrimaryDb()
			expectedNbreUsers++

			test.whenPrimaryIsDeleted()

			test.thenPodsStatesShouldBe(1, 2)

			test.ThenPrimaryDbContainsExpectedNbreUsers(expectedNbreUsers)
			test.ThenReplicaDbContainsExpectedNbreUsers(expectedNbreUsers)

			test.GivenUserAddedInPrimaryDb()
			expectedNbreUsers++

			test.ThenPrimaryDbContainsExpectedNbreUsers(expectedNbreUsers)
			test.ThenReplicaDbContainsExpectedNbreUsers(expectedNbreUsers)

			log.Print("END OF: Test 'GIVEN Kubegres with 1 primary and 2 replicas AND primary is deleted'")
		})
	})

})

type PrimaryFailureAndRecoveryTest struct {
	kubegresResource      *postgresv1.Kubegres
	connectionPrimaryDb   util.DbConnectionDbUtil
	connectionReplicaDb   util.DbConnectionDbUtil
	resourceCreator       util.TestResourceCreator
	resourceRetriever     util.TestResourceRetriever
	customEnvVariableName string
	customEnvVariableKey  string
}

func (r *PrimaryFailureAndRecoveryTest) givenNewKubegresSpecIsSetTo(specNbreReplicas int32) {
	r.kubegresResource = resourceConfigs.LoadKubegresYaml()
	r.kubegresResource.Spec.Replicas = &specNbreReplicas
}

func (r *PrimaryFailureAndRecoveryTest) whenKubernetesIsCreated() {
	r.resourceCreator.CreateKubegres(r.kubegresResource)
}

func (r *PrimaryFailureAndRecoveryTest) whenPrimaryIsDeleted() {
	kubegresResources, err := r.resourceRetriever.GetKubegresResources()
	if err != nil {
		Expect(err).Should(Succeed())
		return
	}

	nbreDeleted := 0
	for _, kubegresResource := range kubegresResources.Resources {
		if kubegresResource.IsPrimary {
			log.Println("Attempting to delete StatefulSet: '" + kubegresResource.StatefulSet.Name + "'")
			if !r.resourceCreator.DeleteResource(kubegresResource.StatefulSet.Resource, kubegresResource.StatefulSet.Name) {
				log.Println("Replica StatefulSet CANNOT BE deleted: '" + kubegresResource.StatefulSet.Name + "'")
			} else {
				nbreDeleted++
				time.Sleep(5 * time.Second)
			}
		}
	}

	Expect(nbreDeleted).Should(Equal(1))
}

func (r *PrimaryFailureAndRecoveryTest) whenPrimaryAndItsPVCAreDeleted() {
	kubegresResources, err := r.resourceRetriever.GetKubegresResources()
	if err != nil {
		Expect(err).Should(Succeed())
		return
	}

	nbreDeletedStafulSet := 0
	nbreDeletedPvc := 0
	for _, kubegresResource := range kubegresResources.Resources {

		if kubegresResource.IsPrimary {

			log.Println("Attempting to delete PVC: '" + kubegresResource.Pvc.Name + "'")
			log.Println("Attempting to delete StatefulSet: '" + kubegresResource.StatefulSet.Name + "'")

			isPvcDeleted := r.resourceCreator.DeleteResource(kubegresResource.Pvc.Resource, kubegresResource.Pvc.Name)
			time.Sleep(10 * time.Second)

			isPrimaryStatefulSetDeleted := r.resourceCreator.DeleteResource(kubegresResource.StatefulSet.Resource, kubegresResource.StatefulSet.Name)

			if isPvcDeleted {
				nbreDeletedPvc++
			} else {
				log.Println("Primary PVC CANNOT BE deleted: '" + kubegresResource.Pvc.Name + "'")
			}

			if isPrimaryStatefulSetDeleted {
				nbreDeletedStafulSet++
			} else {
				log.Println("Primary StatefulSet CANNOT BE deleted: '" + kubegresResource.StatefulSet.Name + "'")
			}

			time.Sleep(5 * time.Second)
		}
	}

	Expect(nbreDeletedStafulSet).Should(Equal(1))
	Expect(nbreDeletedPvc).Should(Equal(1))
}

func (r *PrimaryFailureAndRecoveryTest) wait10Seconds() bool {

	toto := 0

	return Eventually(func() bool {
		toto++
		return toto == 13
	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *PrimaryFailureAndRecoveryTest) thenPodsStatesShouldBe(nbrePrimary, nbreReplicas int) bool {
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

func (r *PrimaryFailureAndRecoveryTest) GivenUserAddedInPrimaryDb() {
	Eventually(func() bool {
		return r.connectionPrimaryDb.InsertUser()
	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *PrimaryFailureAndRecoveryTest) ThenPrimaryDbContainsExpectedNbreUsers(expectedNbreUsers int) {
	Eventually(func() bool {

		users := r.connectionPrimaryDb.GetUsers()
		r.connectionPrimaryDb.Close()

		if len(users) != expectedNbreUsers ||
			r.connectionPrimaryDb.NbreInsertedUsers != expectedNbreUsers {
			log.Println("Primary DB does not contain the expected number of users. Expected: " + strconv.Itoa(expectedNbreUsers) + " Given: " + strconv.Itoa(len(users)))
			return false
		}

		return true

	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *PrimaryFailureAndRecoveryTest) ThenReplicaDbContainsExpectedNbreUsers(expectedNbreUsers int) {
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
