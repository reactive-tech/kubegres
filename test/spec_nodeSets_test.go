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
	"strconv"
	"time"
)

var _ = Describe("Setting Kubegres spec 'nodeSets'", func() {

	var test = SpecNodeSetsTest{}

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

	Context("GIVEN new Kubegres is created with spec 'replica' set to nil and 'nodeSets' set to nil", func() {

		It("THEN a replica validation error event should be logged", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'replica' set to nil and 'nodeSets' set to nil'")

			test.givenNewKubegresSpecIsSetToNil()

			test.whenKubegresIsCreated()

			test.thenReplicasErrorEventShouldBeLogged()

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'replica' set to nil and 'nodeSets' set to nil'")
		})
	})

	Context("GIVEN new Kubegres is created with spec 'replica' set to nil and 'nodeSets' set to an empty list", func() {

		It("THEN a node sets validation error event should be logged", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'replica' set to nil and 'nodeSets' set to an empty list'")

			test.givenNewKubegresSpecWithEmptyNodeSets()

			test.whenKubegresIsCreated()

			test.thenReplicasErrorEventShouldBeLogged()

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'replica' set to nil and 'nodeSets' set to an empty list'")
		})
	})

	Context("GIVEN new Kubegres is created with spec 'replica' set to a value and 'nodeSets' set to a value", func() {

		It("THEN a mutually exclusive validation error event should be logged", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'replica' set to a value and 'nodeSets' set to a value'")

			test.givenNewKubegresSpecWithReplicasAndNodeSetsOfDifferentSize()

			test.whenKubegresIsCreated()

			test.thenMutuallyExclusiveErrorEventShouldBeLogged()

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'replica' set to a value and 'nodeSets' set to a value'")
		})
	})

	Context("GIVEN new Kubegres is created with spec 'nodeSets' set to a value without 'Name'", func() {

		It("THEN a nodeSet[].Name validation error event should be logged", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'nodeSets' set to a value without 'Name''")

			test.givenNewKubegresSpecNodeSetsWithoutName()

			test.whenKubegresIsCreated()

			test.thenNodeSetsNameErrorEventShouldBeLogged()

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'nodeSets' set to a value without 'Name''")
		})
	})

	Context("GIVEN new Kubegres is created with spec 'nodeSets' set to 1 node", func() {

		It("THEN 1 primary and 0 replica should be created", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'nodeSets' set to 1 node'")

			test.givenNewKubegresSpecIsSetTo(1)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 0)

			test.thenDeployedKubegresSpecShouldBeSetTo(1)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'nodeSets' set to 1 node'")
		})

	})

	Context("GIVEN new Kubegres is created with spec 'nodeSets' set to 2 nodes", func() {

		It("THEN 1 primary and 1 replica should be created", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'nodeSets' set to 2 nodes'")

			test.givenNewKubegresSpecIsSetTo(2)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 1)

			test.thenDeployedKubegresSpecShouldBeSetTo(2)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'nodeSets' set to 2 nodes'")
		})

	})

	Context("GIVEN new Kubegres is created with spec 'nodeSets' set to 3 and then it is updated to different values", func() {

		It("GIVEN new Kubegres is created with spec 'nodeSets' set to 3 THEN 1 primary and 2 replica should be created", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'nodeSets' set to 3'")

			test.givenNewKubegresSpecIsSetTo(3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(3)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.keepCreatedResourcesForNextTest = true

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'nodeSets' set to 3'")
		})

		It("GIVEN existing Kubegres is updated with spec 'nodeSets' set from 3 to 4 THEN 1 more replica should be created", func() {

			log.Print("START OF: Test 'GIVEN existing Kubegres is updated with spec 'nodeSets' set from 3 to 4'")

			test.givenExistingKubegresSpecIsSetTo(4)

			test.whenKubernetesIsUpdated()

			test.thenPodsStatesShouldBe(1, 3)

			test.thenDeployedKubegresSpecShouldBeSetTo(4)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.keepCreatedResourcesForNextTest = true

			log.Print("END OF: Test 'GIVEN existing Kubegres is updated with spec 'nodeSets' set from 3 to 4'")
		})

		It("GIVEN existing Kubegres is updated with spec 'nodeSets' with one changed node THEN 1 more replica should be re-created", func() {

			log.Print("START OF: Test 'GIVEN existing Kubegres is updated with spec 'nodeSets' with one changed node THEN 1 more replica should be re-created'")

			test.givenExistingKubegresSpecReplicaNodeNameChanged()

			test.whenKubernetesIsUpdated()

			test.thenPodsStatesShouldBeHavingOnePodNamed(1, 3, "my-kubegres-new-node-1-0")

			test.thenDeployedKubegresSpecShouldBeSetTo(4)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.keepCreatedResourcesForNextTest = true

			log.Print("END OF: Test 'GIVEN existing Kubegres is updated with spec 'nodeSets' with one changed node THEN 1 more replica should be re-created'")
		})

		It("GIVEN existing Kubegres is updated with spec 'nodeSets' set from 4 to 3 THEN 1 replica should be deleted", func() {

			log.Print("START OF: Test 'GIVEN existing Kubegres is updated with spec 'nodeSets' set from 4 to 3'")

			test.givenExistingKubegresSpecIsSetTo(3)

			test.whenKubernetesIsUpdated()

			test.thenPodsStatesShouldBe(1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(3)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.keepCreatedResourcesForNextTest = true

			log.Print("END OF: Test 'GIVEN existing Kubegres is updated with spec 'nodeSets' set from 4 to 3'")
		})

		It("GIVEN existing Kubegres is updated with spec 'nodeSets' set from 3 to 1 THEN 2 replica should be deleted", func() {

			log.Print("START OF: Test 'GIVEN existing Kubegres is updated with spec 'nodeSets' set from 3 to 1'")

			test.givenExistingKubegresSpecIsSetTo(1)

			test.whenKubernetesIsUpdated()

			test.thenPodsStatesShouldBe(1, 0)

			test.thenDeployedKubegresSpecShouldBeSetTo(1)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()

			log.Print("END OF: Test 'GIVEN existing Kubegres is updated with spec 'nodeSets' set from 3 to 1'")
		})
	})

})

type SpecNodeSetsTest struct {
	keepCreatedResourcesForNextTest bool
	kubegresResource                *postgresv1.Kubegres
	dbQueryTestCases                testcases.DbQueryTestCases
	resourceCreator                 util.TestResourceCreator
	resourceRetriever               util.TestResourceRetriever
}

func (r *SpecNodeSetsTest) givenNewKubegresSpecIsSetToNil() {
	r.kubegresResource = resourceConfigs.LoadKubegresYaml()
	r.kubegresResource.Spec.Replicas = nil
	r.kubegresResource.Spec.NodeSets = nil
}

func (r *SpecNodeSetsTest) givenNewKubegresSpecWithEmptyNodeSets() {
	r.kubegresResource = resourceConfigs.LoadKubegresYaml()
	r.kubegresResource.Spec.Replicas = nil
	r.kubegresResource.Spec.NodeSets = []postgresv1.KubegresNodeSet{}
}

func (r *SpecNodeSetsTest) givenNewKubegresSpecWithReplicasAndNodeSetsOfDifferentSize() {
	replicas := int32(2)
	r.kubegresResource = resourceConfigs.LoadKubegresYaml()
	r.kubegresResource.Spec.Replicas = &replicas
	r.kubegresResource.Spec.NodeSets = []postgresv1.KubegresNodeSet{
		{
			Name: "node-1",
		},
	}
}

func (r *SpecNodeSetsTest) givenNewKubegresSpecNodeSetsWithoutName() {
	r.kubegresResource = resourceConfigs.LoadKubegresYaml()
	r.kubegresResource.Spec.Replicas = nil
	r.kubegresResource.Spec.NodeSets = []postgresv1.KubegresNodeSet{
		{},
	}
}

func (r *SpecNodeSetsTest) givenNewKubegresSpecIsSetTo(specNumberOfNodeSets int) {
	r.kubegresResource = resourceConfigs.LoadKubegresYaml()
	r.kubegresResource.Spec.Replicas = nil
	r.kubegresResource.Spec.NodeSets = make([]postgresv1.KubegresNodeSet, specNumberOfNodeSets)
	for i := 0; i < specNumberOfNodeSets; i++ {
		r.kubegresResource.Spec.NodeSets[i] = postgresv1.KubegresNodeSet{
			Name: "node-" + strconv.Itoa(i),
		}
	}
}

func (r *SpecNodeSetsTest) givenExistingKubegresSpecReplicaNodeNameChanged() {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	r.kubegresResource.Spec.NodeSets[1] = postgresv1.KubegresNodeSet{
		Name: "new-node-1",
	}
}

func (r *SpecNodeSetsTest) givenExistingKubegresSpecIsSetTo(numberOfNodeSets int) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	r.kubegresResource.Spec.NodeSets = make([]postgresv1.KubegresNodeSet, numberOfNodeSets)
	for i := 0; i < numberOfNodeSets; i++ {
		r.kubegresResource.Spec.NodeSets[i] = postgresv1.KubegresNodeSet{
			Name: "node-" + strconv.Itoa(i),
		}
	}
}

func (r *SpecNodeSetsTest) whenKubegresIsCreated() {
	r.resourceCreator.CreateKubegres(r.kubegresResource)
}

func (r *SpecNodeSetsTest) whenKubernetesIsUpdated() {
	r.resourceCreator.UpdateResource(r.kubegresResource, "Kubegres")
}

func (r *SpecNodeSetsTest) thenReplicasErrorEventShouldBeLogged() {
	expectedErrorEvent := util.EventRecord{
		Eventtype: v12.EventTypeWarning,
		Reason:    "SpecCheckErr",
		Message:   "In the Resources Spec the value of 'spec.replicas' is undefined. Please set a value otherwise this operator cannot work correctly.",
	}
	Eventually(func() bool {
		_, err := r.resourceRetriever.GetKubegres()
		if err != nil {
			return false
		}
		return eventRecorderTest.CheckEventExist(expectedErrorEvent)

	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *SpecNodeSetsTest) thenMutuallyExclusiveErrorEventShouldBeLogged() {
	expectedErrorEvent := util.EventRecord{
		Eventtype: v12.EventTypeWarning,
		Reason:    "SpecCheckErr",
		Message:   "In the Resources Spec the value of 'spec.replicas' and 'spec.nodeSets' are mutually exclusive. Please set only one of the value otherwise this operator cannot work correctly.",
	}
	Eventually(func() bool {
		_, err := r.resourceRetriever.GetKubegres()
		if err != nil {
			return false
		}
		return eventRecorderTest.CheckEventExist(expectedErrorEvent)

	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *SpecNodeSetsTest) thenNodeSetsNameErrorEventShouldBeLogged() {
	expectedErrorEvent := util.EventRecord{
		Eventtype: v12.EventTypeWarning,
		Reason:    "SpecCheckErr",
		Message:   "In the Resources Spec the value of 'spec.nodeSets[].Name' is undefined. Please set a value otherwise this operator cannot work correctly.",
	}
	Eventually(func() bool {
		_, err := r.resourceRetriever.GetKubegres()
		if err != nil {
			return false
		}
		return eventRecorderTest.CheckEventExist(expectedErrorEvent)

	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *SpecNodeSetsTest) thenPodsStatesShouldBe(numberOfPrimary, numberOfReplicas int) bool {
	return Eventually(func() bool {

		pods, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres pods")
			return false
		}

		if pods.AreAllReady &&
			pods.NbreDeployedPrimary == numberOfPrimary &&
			pods.NbreDeployedReplicas == numberOfReplicas {

			time.Sleep(resourceConfigs.TestRetryInterval)
			log.Println("Deployed and Ready StatefulSets check successful")
			return true
		}

		return false

	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *SpecNodeSetsTest) thenPodsStatesShouldBeHavingOnePodNamed(numberOfPrimary, numberOfReplicas int, newName string) bool {
	return Eventually(func() bool {

		pods, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres pods")
			return false
		}

		if pods.AreAllReady &&
			pods.NbreDeployedPrimary == numberOfPrimary &&
			pods.NbreDeployedReplicas == numberOfReplicas {
			for _, pod := range pods.Resources {
				if pod.Pod.Name == newName {
					time.Sleep(resourceConfigs.TestRetryInterval)
					log.Println("Deployed and Ready StatefulSets check successful")
					return true
				}
			}
		}

		return false

	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *SpecNodeSetsTest) thenDeployedKubegresSpecShouldBeSetTo(expectedNodeSetsSize int) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	nodeSets := r.kubegresResource.Spec.NodeSets
	Expect(len(nodeSets)).Should(Equal(expectedNodeSetsSize))
}
