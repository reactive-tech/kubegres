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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"log"
	postgresv1 "reactive-tech.io/kubegres/api/v1"
	resourceConfigs2 "reactive-tech.io/kubegres/internal/test/resourceConfigs"
	util2 "reactive-tech.io/kubegres/internal/test/util"
	"strconv"
	"time"
)

var _ = Describe("We set the wal-level to 'logical' and simulate when Primary instance is not available then recovery/failover works", func() {

	var test = PostgresConfWalLevelLogicalTest{}

	BeforeEach(func() {
		//Skip("Temporarily skipping test")

		namespace := resourceConfigs2.DefaultNamespace
		test.resourceRetriever = util2.CreateTestResourceRetriever(k8sClientTest, namespace)
		test.resourceCreator = util2.CreateTestResourceCreator(k8sClientTest, test.resourceRetriever, namespace)
		test.connectionPrimaryDb = util2.InitDbConnectionDbUtil(test.resourceCreator, resourceConfigs2.KubegresResourceName, resourceConfigs2.ServiceToSqlQueryPrimaryDbNodePort, true)
		test.connectionReplicaDb = util2.InitDbConnectionDbUtil(test.resourceCreator, resourceConfigs2.KubegresResourceName, resourceConfigs2.ServiceToSqlQueryReplicaDbNodePort, false)
		test.resourceCreator.CreateConfigMapWithPostgresConfAndWalLevelSetToLogical()
	})

	AfterEach(func() {
		test.resourceCreator.DeleteAllTestResources()
	})

	Context("GIVEN Kubegres custom ConfigMap with wal-level set to 'logical' AND with 1 primary and 2 replicas AND primary is deleted", func() {

		It("THEN the failover should take place with a replica becoming primary AND a new replica created AND existing data available", func() {

			log.Print("START OF: Test 'GIVEN Kubegres custom ConfigMap with wal-level set to 'logical' AND with 1 primary and 2 replicas AND primary is deleted'")

			test.givenNewKubegresSpecIsSetTo(resourceConfigs2.CustomConfigMapWithPostgresConfAndWalLevelSetToLogicalResourceName, 3)

			test.whenKubegresIsCreated()

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

			log.Print("END OF: Test 'GIVEN Kubegres custom ConfigMap with wal-level set to 'logical' AND with 1 primary and 2 replicas AND primary is deleted'")
		})
	})

})

type PostgresConfWalLevelLogicalTest struct {
	kubegresResource      *postgresv1.Kubegres
	connectionPrimaryDb   util2.DbConnectionDbUtil
	connectionReplicaDb   util2.DbConnectionDbUtil
	resourceCreator       util2.TestResourceCreator
	resourceRetriever     util2.TestResourceRetriever
	customEnvVariableName string
	customEnvVariableKey  string
}

func (r *PostgresConfWalLevelLogicalTest) givenNewKubegresSpecIsSetTo(customConfig string, specNbreReplicas int32) {
	r.kubegresResource = resourceConfigs2.LoadKubegresYaml()
	r.kubegresResource.Spec.CustomConfig = customConfig
	r.kubegresResource.Spec.Replicas = &specNbreReplicas
}

func (r *PostgresConfWalLevelLogicalTest) whenKubegresIsCreated() {
	r.resourceCreator.CreateKubegres(r.kubegresResource)
}

func (r *PostgresConfWalLevelLogicalTest) whenPrimaryIsDeleted() {
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

func (r *PostgresConfWalLevelLogicalTest) whenPrimaryAndItsPVCAreDeleted() {
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

func (r *PostgresConfWalLevelLogicalTest) wait10Seconds() bool {

	waitingAttempts := 0

	return Eventually(func() bool {
		waitingAttempts++
		return waitingAttempts == 13
	}, resourceConfigs2.TestTimeout, resourceConfigs2.TestRetryInterval).Should(BeTrue())
}

func (r *PostgresConfWalLevelLogicalTest) thenPodsStatesShouldBe(nbrePrimary, nbreReplicas int) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
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

func (r *PostgresConfWalLevelLogicalTest) GivenUserAddedInPrimaryDb() {
	Eventually(func() bool {
		return r.connectionPrimaryDb.InsertUser()
	}, resourceConfigs2.TestTimeout, resourceConfigs2.TestRetryInterval).Should(BeTrue())
}

func (r *PostgresConfWalLevelLogicalTest) ThenPrimaryDbContainsExpectedNbreUsers(expectedNbreUsers int) {
	Eventually(func() bool {

		users := r.connectionPrimaryDb.GetUsers()
		r.connectionPrimaryDb.Close()

		if len(users) != expectedNbreUsers ||
			r.connectionPrimaryDb.NbreInsertedUsers != expectedNbreUsers {
			log.Println("Primary DB does not contain the expected number of users. Expected: " + strconv.Itoa(expectedNbreUsers) + " Given: " + strconv.Itoa(len(users)))
			return false
		}

		return true

	}, resourceConfigs2.TestTimeout, resourceConfigs2.TestRetryInterval).Should(BeTrue())
}

func (r *PostgresConfWalLevelLogicalTest) ThenReplicaDbContainsExpectedNbreUsers(expectedNbreUsers int) {
	Eventually(func() bool {

		users := r.connectionReplicaDb.GetUsers()
		r.connectionReplicaDb.Close()

		if len(users) != expectedNbreUsers ||
			r.connectionReplicaDb.NbreInsertedUsers != expectedNbreUsers {
			log.Println("Replica DB does not contain the expected number of users: " + strconv.Itoa(expectedNbreUsers))
			return false
		}

		return true

	}, resourceConfigs2.TestTimeout, resourceConfigs2.TestRetryInterval).Should(BeTrue())
}
