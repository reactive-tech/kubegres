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

var _ = Describe("Setting Kubegres spec 'scheduler.tolerations'", func() {

	var test = SpecTolerationsTest{}

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

	Context("GIVEN new Kubegres is created without spec 'scheduler.tolerations' and with spec 'replica' set to 3", func() {

		It("THEN 1 primary and 2 replica should be created with 'scheduler.tolerations' set to nil", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created without spec 'scheduler.tolerations' and with spec 'replica' set to 3'")

			test.givenNewKubegresSpecIsWithoutTolerations(3)

			test.whenKubegresIsCreated()

			test.thenStatefulSetStatesShouldBeWithoutTolerations(1, 2)

			test.thenDeployedKubegresSpecShouldWithoutTolerations()

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN new Kubegres is created without spec 'scheduler.tolerations' and with spec 'replica' set to 3'")
		})
	})

	Context("GIVEN new Kubegres is created with spec 'scheduler.tolerations' set to a value and spec 'replica' set to 3 and later 'scheduler.tolerations' is updated to a new value", func() {

		It("GIVEN new Kubegres is created with spec 'scheduler.tolerations' set to a value and spec 'replica' set to 3 THEN 1 primary and 2 replica should be created with spec 'scheduler.tolerations' set the value", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'scheduler.tolerations' set to a value and spec 'replica' set to 3")

			toleration := test.givenToleration1()

			test.givenNewKubegresSpecIsSetTo(toleration, 3)

			test.whenKubegresIsCreated()

			test.thenStatefulSetStatesShouldBe(toleration, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(toleration)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.keepCreatedResourcesForNextTest = true

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'scheduler.tolerations' set to a value and spec 'replica' set to 3'")
		})

		It("GIVEN existing Kubegres is updated with spec 'scheduler.tolerations' set to a new value THEN 1 primary and 2 replica should be re-deployed with spec 'scheduler.tolerations' set the new value", func() {

			log.Print("START OF: Test 'GIVEN existing Kubegres is updated with spec 'scheduler.tolerations' set to a new value")

			newToleration := test.givenToleration2()

			test.givenExistingKubegresSpecIsSetTo(newToleration)

			test.whenKubernetesIsUpdated()

			test.thenStatefulSetStatesShouldBe(newToleration, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(newToleration)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN existing Kubegres is updated with spec 'scheduler.tolerations' set to a new value")
		})
	})

})

type SpecTolerationsTest struct {
	keepCreatedResourcesForNextTest bool
	kubegresResource                *postgresv1.Kubegres
	dbQueryTestCases                testcases.DbQueryTestCases
	resourceCreator                 util.TestResourceCreator
	resourceRetriever               util.TestResourceRetriever
}

func (r *SpecTolerationsTest) givenToleration1() v12.Toleration {
	return v12.Toleration{
		Key:      "group",
		Operator: v12.TolerationOpEqual,
		Value:    "critical",
	}
}

func (r *SpecTolerationsTest) givenToleration2() v12.Toleration {
	return v12.Toleration{
		Key:      "group",
		Operator: v12.TolerationOpEqual,
		Value:    "nonCritical",
	}
}

func (r *SpecTolerationsTest) givenNewKubegresSpecIsWithoutTolerations(specNbreReplicas int32) {
	r.kubegresResource = resourceConfigs.LoadKubegresYaml()
	r.kubegresResource.Spec.Scheduler.Tolerations = nil
	r.kubegresResource.Spec.Replicas = &specNbreReplicas
}

func (r *SpecTolerationsTest) givenNewKubegresSpecIsSetTo(toleration v12.Toleration, specNbreReplicas int32) {
	r.kubegresResource = resourceConfigs.LoadKubegresYaml()
	r.kubegresResource.Spec.Scheduler.Tolerations = []v12.Toleration{toleration}
	r.kubegresResource.Spec.Replicas = &specNbreReplicas
}

func (r *SpecTolerationsTest) givenExistingKubegresSpecIsSetTo(toleration v12.Toleration) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	r.kubegresResource.Spec.Scheduler.Tolerations = []v12.Toleration{toleration}
}

func (r *SpecTolerationsTest) whenKubegresIsCreated() {
	r.resourceCreator.CreateKubegres(r.kubegresResource)
}

func (r *SpecTolerationsTest) whenKubernetesIsUpdated() {
	r.resourceCreator.UpdateResource(r.kubegresResource, "Kubegres")
}

func (r *SpecTolerationsTest) thenStatefulSetStatesShouldBeWithoutTolerations(nbrePrimary, nbreReplicas int) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		for _, resource := range kubegresResources.Resources {
			tolerations := resource.StatefulSet.Spec.Template.Spec.Tolerations

			if len(tolerations) > 0 {
				log.Println("StatefulSet '" + resource.StatefulSet.Name + "' doesn't have the expected spec 'scheduler.toleration' which should be nil. " +
					"Current value: '" + tolerations[0].String() + "'. Waiting...")
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

func (r *SpecTolerationsTest) thenStatefulSetStatesShouldBe(expectedToleration v12.Toleration, nbrePrimary, nbreReplicas int) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		for _, resource := range kubegresResources.Resources {
			tolerations := resource.StatefulSet.Spec.Template.Spec.Tolerations

			if len(tolerations) != 1 {
				log.Println("StatefulSet '" + resource.StatefulSet.Name + "' doesn't have the expected spec 'scheduler.toleration' which is nil when it should have a value. " +
					"Current value: '" + tolerations[0].String() + "'. Waiting...")
				return false

			} else if tolerations[0] != expectedToleration {
				log.Println("StatefulSet '" + resource.StatefulSet.Name + "' doesn't have the expected spec scheduler.toleration: " + expectedToleration.String() + " " +
					"Current value: '" + tolerations[0].String() + "'. Waiting...")
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

func (r *SpecTolerationsTest) thenDeployedKubegresSpecShouldWithoutTolerations() {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	Expect(len(r.kubegresResource.Spec.Scheduler.Tolerations)).Should(Equal(0))
}

func (r *SpecTolerationsTest) thenDeployedKubegresSpecShouldBeSetTo(expectedToleration v12.Toleration) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	Expect(len(r.kubegresResource.Spec.Scheduler.Tolerations)).Should(Equal(1))
	Expect(r.kubegresResource.Spec.Scheduler.Tolerations[0]).Should(Equal(expectedToleration))
}
