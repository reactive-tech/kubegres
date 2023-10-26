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
	"strconv"
	"time"
)

var _ = Describe("Primary instances is not available, when the promotion of a PostgreSql instance is manually requested THEN promotion should be triggered.", func() {

	var test = SpecFailoverIsDisabledAndPromotePodAreSetTest{}

	BeforeEach(func() {
		Skip("Temporarily skipping test")

		namespace := resourceConfigs2.DefaultNamespace
		test.resourceRetriever = util2.CreateTestResourceRetriever(k8sClientTest, namespace)
		test.resourceCreator = util2.CreateTestResourceCreator(k8sClientTest, test.resourceRetriever, namespace)
		test.connectionPrimaryDb = util2.InitDbConnectionDbUtil(test.resourceCreator, resourceConfigs2.KubegresResourceName, resourceConfigs2.ServiceToSqlQueryPrimaryDbNodePort, true)
		test.connectionReplicaDb = util2.InitDbConnectionDbUtil(test.resourceCreator, resourceConfigs2.KubegresResourceName, resourceConfigs2.ServiceToSqlQueryReplicaDbNodePort, false)
	})

	AfterEach(func() {
		test.resourceCreator.DeleteAllTestResources()
	})

	Context("GIVEN Kubegres with 1 primary and 1 replica AND once deployed we update YAML with 'failover.promotePod' is set to a Pod name", func() {

		It("THEN the replica Pod set in spec 'failover.promotePod' should become the new primary AND a new replica should be created", func() {

			log.Print("START OF: Test 'GIVEN Kubegres with 1 primary and 1 replica AND once deployed we update YAML with 'failover.promotePod' is set to a Pod name'")

			test.givenNewKubegresSpecIsSetTo(2)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 1)

			expectedNbreUsers := 0

			test.givenUserAddedInPrimaryDb()
			expectedNbreUsers++

			test.givenUserAddedInPrimaryDb()
			expectedNbreUsers++

			replicaPodNameToPromote := test.getReplicaPodName()

			test.givenExistingKubegresSpecIsSetTo(replicaPodNameToPromote)

			test.whenKubernetesIsUpdated()

			time.Sleep(time.Second * 10)

			test.thenPodsStatesShouldBe(1, 1)

			test.thenPromotePodFieldInSpecShouldBeCleared()

			test.thenPrimaryPodNameMatches(replicaPodNameToPromote)

			test.thenPrimaryDbContainsExpectedNbreUsers(expectedNbreUsers)
			test.thenReplicaDbContainsExpectedNbreUsers(expectedNbreUsers)

			log.Print("END OF: Test 'GIVEN Kubegres with 1 primary and 1 replica AND once deployed we update YAML with 'failover.promotePod' is set to a Pod name'")
		})
	})

	Context("GIVEN Kubegres with 1 primary and 1 replica AND once deployed we update YAML with 'failover.isDisabled' is true AND with 'failover.promotePod' is set to a Pod name", func() {

		It("THEN the replica Pod set in spec 'failover.promotePod' should become the new primary AND a new replica should be created", func() {

			log.Print("START OF: Test 'GIVEN Kubegres with 1 primary and 1 replica AND once deployed we update YAML with 'failover.isDisabled' is true AND with 'failover.promotePod' is set to a Pod name'")

			test.givenNewKubegresSpecIsSetTo(2)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 1)

			expectedNbreUsers := 0

			test.givenUserAddedInPrimaryDb()
			expectedNbreUsers++

			test.givenUserAddedInPrimaryDb()
			expectedNbreUsers++

			replicaPodNameToPromote := test.getReplicaPodName()

			test.givenExistingKubegresSpecIsUpdatedTo(true, replicaPodNameToPromote)

			test.whenKubernetesIsUpdated()

			time.Sleep(time.Second * 10)

			test.thenPodsStatesShouldBe(1, 1)

			test.thenPromotePodFieldInSpecShouldBeCleared()

			test.thenPrimaryPodNameMatches(replicaPodNameToPromote)

			test.thenPrimaryDbContainsExpectedNbreUsers(expectedNbreUsers)
			test.thenReplicaDbContainsExpectedNbreUsers(expectedNbreUsers)

			log.Print("END OF: Test 'GIVEN Kubegres with 1 primary and 1 replica AND once deployed we update YAML with 'failover.isDisabled' is true AND with 'failover.promotePod' is set to a Pod name'")
		})
	})

	Context("GIVEN Kubegres with 1 primary and 1 replica AND once deployed we update YAML with 'failover.promotePod' is set to a Pod name which does NOT exist", func() {

		It("THEN nothing should happen AND an error message should be logged as event saying Pod does NOT exist", func() {

			log.Print("START OF: Test 'GIVEN Kubegres with 1 primary and 1 replica AND once deployed we update YAML with 'failover.promotePod' is set to a Pod name which does NOT exist'")

			test.givenNewKubegresSpecIsSetTo(2)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 1)

			expectedNbreUsers := 0

			test.givenUserAddedInPrimaryDb()
			expectedNbreUsers++

			test.givenUserAddedInPrimaryDb()
			expectedNbreUsers++

			primaryPodName, replicaPodName := test.getDeployedPodNames()

			replicaPodNameToPromote := "Pod_does_not_exist"

			test.givenExistingKubegresSpecIsSetTo(replicaPodNameToPromote)

			test.whenKubernetesIsUpdated()

			time.Sleep(time.Second * 10)

			test.thenPodsStatesShouldBe(1, 1)

			test.thenPromotePodFieldInSpecShouldBeCleared()

			test.thenErrorEventShouldBeLogged(replicaPodNameToPromote)

			test.thenDeployedPodNamesMatch(primaryPodName, replicaPodName)

			test.thenPrimaryDbContainsExpectedNbreUsers(expectedNbreUsers)
			test.thenReplicaDbContainsExpectedNbreUsers(expectedNbreUsers)

			log.Print("END OF: Test 'GIVEN Kubegres with 1 primary and 1 replica AND once deployed we update YAML with 'failover.promotePod' is set to a Pod name which does NOT exist'")
		})
	})

	Context("GIVEN Kubegres with 1 primary and 1 replica AND once deployed we update YAML with 'failover.promotePod' is set the Primary Pod name", func() {

		It("THEN nothing should happen", func() {

			log.Print("START OF: Test 'GIVEN Kubegres with 1 primary and 1 replica AND once deployed we update YAML with 'failover.promotePod' is set the Primary Pod name'")

			test.givenNewKubegresSpecIsSetTo(2)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 1)

			expectedNbreUsers := 0

			test.givenUserAddedInPrimaryDb()
			expectedNbreUsers++

			test.givenUserAddedInPrimaryDb()
			expectedNbreUsers++

			primaryPodName, replicaPodName := test.getDeployedPodNames()

			replicaPodNameToPromote := primaryPodName

			test.givenExistingKubegresSpecIsSetTo(replicaPodNameToPromote)

			test.whenKubernetesIsUpdated()

			time.Sleep(time.Second * 10)

			test.thenPodsStatesShouldBe(1, 1)

			test.thenPromotePodFieldInSpecShouldBeCleared()

			test.thenDeployedPodNamesMatch(primaryPodName, replicaPodName)

			test.thenPrimaryDbContainsExpectedNbreUsers(expectedNbreUsers)
			test.thenReplicaDbContainsExpectedNbreUsers(expectedNbreUsers)

			log.Print("END OF: Test 'GIVEN Kubegres with 1 primary and 1 replica AND once deployed we update YAML with 'failover.promotePod' is set the Primary Pod name'")
		})
	})
})

type SpecFailoverIsDisabledAndPromotePodAreSetTest struct {
	kubegresResource      *postgresv1.Kubegres
	connectionPrimaryDb   util2.DbConnectionDbUtil
	connectionReplicaDb   util2.DbConnectionDbUtil
	resourceCreator       util2.TestResourceCreator
	resourceRetriever     util2.TestResourceRetriever
	customEnvVariableName string
	customEnvVariableKey  string
}

func (r *SpecFailoverIsDisabledAndPromotePodAreSetTest) givenNewKubegresSpecIsSetTo(specNbreReplicas int32) {
	r.kubegresResource = resourceConfigs2.LoadKubegresYaml()
	r.kubegresResource.Spec.Replicas = &specNbreReplicas
}

func (r *SpecFailoverIsDisabledAndPromotePodAreSetTest) givenExistingKubegresSpecIsSetTo(promotePodName string) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	r.kubegresResource.Spec.Failover.PromotePod = promotePodName
}

func (r *SpecFailoverIsDisabledAndPromotePodAreSetTest) givenExistingKubegresSpecIsUpdatedTo(isFailoverDisabled bool, promotePodName string) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	r.kubegresResource.Spec.Failover.IsDisabled = isFailoverDisabled
	r.kubegresResource.Spec.Failover.PromotePod = promotePodName
}

func (r *SpecFailoverIsDisabledAndPromotePodAreSetTest) givenUserAddedInPrimaryDb() {
	Eventually(func() bool {
		return r.connectionPrimaryDb.InsertUser()
	}, resourceConfigs2.TestTimeout, resourceConfigs2.TestRetryInterval).Should(BeTrue())
}

func (r *SpecFailoverIsDisabledAndPromotePodAreSetTest) getDeployedPodNames() (primaryPodName, replicaPodName string) {

	kubegresResources, err := r.resourceRetriever.GetKubegresResources()
	if err != nil {
		Expect(err).Should(Succeed())
		return
	}

	for _, kubegresResource := range kubegresResources.Resources {
		if kubegresResource.IsPrimary {
			primaryPodName = kubegresResource.Pod.Name
		} else {
			replicaPodName = kubegresResource.Pod.Name
		}
	}

	return primaryPodName, replicaPodName
}

func (r *SpecFailoverIsDisabledAndPromotePodAreSetTest) whenKubegresIsCreated() {
	r.resourceCreator.CreateKubegres(r.kubegresResource)
}

func (r *SpecFailoverIsDisabledAndPromotePodAreSetTest) whenKubernetesIsUpdated() {
	r.resourceCreator.UpdateResource(r.kubegresResource, "Kubegres")
}

func (r *SpecFailoverIsDisabledAndPromotePodAreSetTest) thenPrimaryPodNameMatches(expectedPromotedPod string) {
	kubegresResources, err := r.resourceRetriever.GetKubegresResources()
	if err != nil {
		Expect(err).Should(Succeed())
		return
	}

	primaryPodName := ""
	for _, kubegresResource := range kubegresResources.Resources {
		if kubegresResource.IsPrimary {
			primaryPodName = kubegresResource.Pod.Name
			break
		}
	}

	Expect(primaryPodName).Should(Equal(expectedPromotedPod))
}

func (r *SpecFailoverIsDisabledAndPromotePodAreSetTest) thenPodsStatesShouldBe(nbrePrimary, nbreReplicas int) bool {

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

			time.Sleep(resourceConfigs2.TestRetryInterval)
			log.Println("Deployed and Ready StatefulSets check successful")
			return true
		}

		return false

	}, resourceConfigs2.TestTimeout, resourceConfigs2.TestRetryInterval).Should(BeTrue())
}

func (r *SpecFailoverIsDisabledAndPromotePodAreSetTest) thenErrorEventShouldBeLogged(podName string) {

	expectedErrorEvent := util2.EventRecord{
		Eventtype: v12.EventTypeWarning,
		Reason:    "ManualFailoverCannotHappenAsConfigErr",
		Message: "The value of the field 'failover.promotePod' is set to '" + podName + "'. " +
			"That value is either the name of a Primary Pod OR a Replica Pod which is not ready OR a Pod which does not exist. " +
			"Please set the name of a Replica Pod that you would like to promote as a Primary Pod.",
	}
	Eventually(func() bool {
		_, err := r.resourceRetriever.GetKubegres()
		if err != nil {
			return false
		}
		return eventRecorderTest.CheckEventExist(expectedErrorEvent)

	}, resourceConfigs2.TestTimeout, resourceConfigs2.TestRetryInterval).Should(BeTrue())
}

func (r *SpecFailoverIsDisabledAndPromotePodAreSetTest) getReplicaPodName() string {
	kubegresResources, _ := r.resourceRetriever.GetKubegresResources()
	for _, kubegresResource := range kubegresResources.Resources {
		if !kubegresResource.IsPrimary {
			return kubegresResource.Pod.Name
		}
	}
	return ""
}

func (r *SpecFailoverIsDisabledAndPromotePodAreSetTest) thenPrimaryDbContainsExpectedNbreUsers(expectedNbreUsers int) {
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

func (r *SpecFailoverIsDisabledAndPromotePodAreSetTest) thenReplicaDbContainsExpectedNbreUsers(expectedNbreUsers int) {
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

func (r *SpecFailoverIsDisabledAndPromotePodAreSetTest) thenDeployedPodNamesMatch(primaryPodName, replicaPodName string) {

	kubegresResources, err := r.resourceRetriever.GetKubegresResources()
	if err != nil {
		Expect(err).Should(Succeed())
		return
	}

	for _, kubegresResource := range kubegresResources.Resources {
		if kubegresResource.IsPrimary {
			if primaryPodName != kubegresResource.Pod.Name {
				Expect(err).Should(Succeed())
				return
			}
		} else {
			if replicaPodName != kubegresResource.Pod.Name {
				Expect(err).Should(Succeed())
				return
			}
		}
	}
}

func (r *SpecFailoverIsDisabledAndPromotePodAreSetTest) thenPromotePodFieldInSpecShouldBeCleared() {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	if r.kubegresResource.Spec.Failover.PromotePod != "" {
		Expect(err).Should(Succeed())
	}
}
