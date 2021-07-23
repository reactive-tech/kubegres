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

var _ = Describe("Primary instances is not available, when failover is disabled checking NO failover should be triggered", func() {

	var test = SpecFailoverIsDisabledTest{}

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

	Context("GIVEN Kubegres is creating a cluster with changing values for the number of deployed replicas AND in YAML the flag 'failover.isDisabled' is true", func() {

		It("THEN a primary and a replica instances should be created as normal", func() {

			log.Print("START OF: Test 'GIVEN Kubegres is creating a cluster with changing values for the number of deployed replicas AND in YAML the flag 'failover.isDisabled' is true'")

			test.givenNewKubegresSpecIsSetTo(true, 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 2)

			test.givenExistingKubegresSpecIsSetTo(true, 4)

			test.whenKubernetesIsUpdated()

			test.thenPodsStatesShouldBe(1, 3)

			test.givenExistingKubegresSpecIsSetTo(true, 2)

			test.whenKubernetesIsUpdated()

			test.thenPodsStatesShouldBe(1, 1)

			log.Print("END OF: Test 'GIVEN Kubegres is creating a cluster with changing values for the number of deployed replicas AND in YAML the flag 'failover.isDisabled' is true'")
		})
	})

	Context("GIVEN Kubegres with 1 primary and 2 replicas AND once deployed we update YAML with 'failover.isDisabled' is true AND we delete primary", func() {

		It("THEN the primary failover should NOT take place AND the deleted primary should NOT be replaced", func() {

			log.Print("START OF: Test 'GIVEN Kubegres with 1 primary and 2 replicas AND once deployed we update YAML with 'failover.isDisabled' is true AND we delete primary'")

			test.givenNewKubegresSpecIsSetTo(false, 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 2)

			expectedNbreUsers := 0

			test.givenUserAddedInPrimaryDb()
			expectedNbreUsers++

			test.givenUserAddedInPrimaryDb()
			expectedNbreUsers++

			test.givenExistingKubegresSpecIsSetTo(true, 3)

			test.whenKubernetesIsUpdated()

			time.Sleep(time.Second * 10)

			test.thenPodsStatesShouldBe(1, 2)

			test.whenPrimaryIsDeleted()

			time.Sleep(time.Second * 10)

			test.thenPodsStatesShouldBe(0, 2)

			test.thenReplicaDbContainsExpectedNbreUsers(expectedNbreUsers)

			log.Print("END OF: Test 'GIVEN Kubegres with 1 primary and 2 replicas AND once deployed we update YAML with 'failover.isDisabled' is true AND we delete primary'")
		})
	})

	Context("GIVEN Kubegres with 1 primary and 2 replicas AND once deployed we update YAML with 'failover.isDisabled' is true AND we delete a replica", func() {

		It("THEN the replica failover should NOT take place AND the deleted replica should NOT be replaced", func() {

			log.Print("START OF: Test 'GIVEN Kubegres with 1 primary and 2 replicas AND once deployed we update YAML with 'failover.isDisabled' is true AND we delete a replica'")

			test.givenNewKubegresSpecIsSetTo(false, 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 2)

			expectedNbreUsers := 0

			test.givenUserAddedInPrimaryDb()
			expectedNbreUsers++

			test.givenUserAddedInPrimaryDb()
			expectedNbreUsers++

			test.givenExistingKubegresSpecIsSetTo(true, 3)

			test.whenKubernetesIsUpdated()

			time.Sleep(time.Second * 10)

			test.thenPodsStatesShouldBe(1, 2)

			test.whenOneReplicaIsDeleted()

			time.Sleep(time.Second * 10)

			test.thenPodsStatesShouldBe(1, 1)

			test.thenReplicaDbContainsExpectedNbreUsers(expectedNbreUsers)

			log.Print("END OF: Test 'GIVEN Kubegres with 1 primary and 2 replicas AND once deployed we update YAML with 'failover.isDisabled' is true AND we delete a replica'")
		})
	})

})

type SpecFailoverIsDisabledTest struct {
	kubegresResource      *postgresv1.Kubegres
	connectionPrimaryDb   util.DbConnectionDbUtil
	connectionReplicaDb   util.DbConnectionDbUtil
	resourceCreator       util.TestResourceCreator
	resourceRetriever     util.TestResourceRetriever
	customEnvVariableName string
	customEnvVariableKey  string
}

func (r *SpecFailoverIsDisabledTest) givenNewKubegresSpecIsSetTo(isFailoverDisabled bool, specNbreReplicas int32) {
	r.kubegresResource = resourceConfigs.LoadKubegresYaml()
	r.kubegresResource.Spec.Replicas = &specNbreReplicas
	r.kubegresResource.Spec.Failover.IsDisabled = isFailoverDisabled
}

func (r *SpecFailoverIsDisabledTest) givenExistingKubegresSpecIsSetTo(isFailoverDisabled bool, specNbreReplicas int32) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	r.kubegresResource.Spec.Replicas = &specNbreReplicas
	r.kubegresResource.Spec.Failover.IsDisabled = isFailoverDisabled
}

func (r *SpecFailoverIsDisabledTest) givenUserAddedInPrimaryDb() {
	Eventually(func() bool {
		return r.connectionPrimaryDb.InsertUser()
	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *SpecFailoverIsDisabledTest) whenKubegresIsCreated() {
	r.resourceCreator.CreateKubegres(r.kubegresResource)
}

func (r *SpecFailoverIsDisabledTest) whenKubernetesIsUpdated() {
	r.resourceCreator.UpdateResource(r.kubegresResource, "Kubegres")
}

func (r *SpecFailoverIsDisabledTest) whenPrimaryIsDeleted() {
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
				log.Println("StatefulSet CANNOT BE deleted: '" + kubegresResource.StatefulSet.Name + "'")
			} else {
				nbreDeleted++
				time.Sleep(5 * time.Second)
			}
		}
	}

	Expect(nbreDeleted).Should(Equal(1))
}

func (r *SpecFailoverIsDisabledTest) whenOneReplicaIsDeleted() {
	kubegresResources, err := r.resourceRetriever.GetKubegresResources()
	if err != nil {
		Expect(err).Should(Succeed())
		return
	}

	nbreDeleted := 0
	for _, kubegresResource := range kubegresResources.Resources {

		if nbreDeleted == 1 {
			break
		}

		if !kubegresResource.IsPrimary {
			log.Println("Attempting to delete StatefulSet: '" + kubegresResource.StatefulSet.Name + "'")
			if !r.resourceCreator.DeleteResource(kubegresResource.StatefulSet.Resource, kubegresResource.StatefulSet.Name) {
				log.Println("StatefulSet CANNOT BE deleted: '" + kubegresResource.StatefulSet.Name + "'")
			} else {
				nbreDeleted++
				time.Sleep(5 * time.Second)
			}
		}
	}

	Expect(nbreDeleted).Should(Equal(1))
}

func (r *SpecFailoverIsDisabledTest) thenPodsStatesShouldBe(nbrePrimary, nbreReplicas int) bool {

	nreInstancesInSpec := int(*r.kubegresResource.Spec.Replicas)
	nbreInstancesWanted := nbrePrimary + nbreReplicas

	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		if (nreInstancesInSpec != nbreInstancesWanted || kubegresResources.AreAllReady) &&
			kubegresResources.NbreDeployedPrimary == nbrePrimary &&
			kubegresResources.NbreDeployedReplicas == nbreReplicas {

			time.Sleep(resourceConfigs.TestRetryInterval)
			log.Println("Deployed and Ready StatefulSets check successful")
			return true
		}

		return false

	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *SpecFailoverIsDisabledTest) thenReplicaDbContainsExpectedNbreUsers(expectedNbreUsers int) {
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
