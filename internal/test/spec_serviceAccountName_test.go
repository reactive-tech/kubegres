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
	"log"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	postgresv1 "reactive-tech.io/kubegres/api/v1"
	"reactive-tech.io/kubegres/internal/controller/ctx"
	"reactive-tech.io/kubegres/internal/test/resourceConfigs"
	"reactive-tech.io/kubegres/internal/test/util"
	"reactive-tech.io/kubegres/internal/test/util/testcases"
)

var _ = Describe("Setting Kubegres spec 'serviceAccountName'", func() {

	var test = SpecServiceAccountNameTest{}

	BeforeEach(func() {
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

	Context("GIVEN new Kubegres is created without spec 'serviceAccountName' and with spec 'replica' set to 3", func() {

		It("THEN 1 primary and 2 replica should be created with serviceAccountName 'default' and a normal event should be logged", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created without spec 'serviceAccountName' and with spec 'replica' set to 3'")

			test.givenNewKubegresSpecIsSetTo("", 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(ctx.DefaultPodServiceAccountName, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo("")

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN new Kubegres is created without spec 'serviceAccountName' and with spec 'replica' set to 3'")
		})
	})

	Context("GIVEN new Kubegres is created with spec 'serviceAccountName' set to 'my-kubegres' and spec 'replica' set to 3 and later 'serviceAccountName' is updated to 'changed-kubegres'", func() {

		It("GIVEN new Kubegres is created with spec 'serviceAccountName' set to 'my-kubegres' and spec 'replica' set to 3 THEN 1 primary and 2 replica should be created with spec 'serviceAccountName' set to 'kubegres", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'serviceAccountName' set to 'my-kubegres' and spec 'replica' set to 3")

			test.whenServiceAccountIsCreated("my-kubegres")

			test.givenNewKubegresSpecIsSetTo("my-kubegres", 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe("my-kubegres", 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo("my-kubegres")

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.keepCreatedResourcesForNextTest = true

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'serviceAccountName' set to 'my-kubegres' and spec 'replica' set to 3'")
		})

		It("GIVEN existing Kubegres is updated with spec 'serviceAccountName' set from 'my-kubegres' to 'changed-kubegres' THEN 1 primary and 2 replica should be re-deployed with spec 'serviceAccountName' set to 'changed-kubegres'", func() {

			log.Print("START OF: Test 'GIVEN existing Kubegres is updated with spec 'serviceAccountName' set from 'my-kubegres' to 'changed-kubegres'")

			test.whenServiceAccountIsCreated("changed-kubegres")

			test.givenExistingKubegresSpecIsSetTo("changed-kubegres")

			test.whenKubernetesIsUpdated()

			test.thenPodsStatesShouldBe("changed-kubegres", 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo("changed-kubegres")

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN existing Kubegres is updated with spec 'serviceAccountName' set from 'my-kubegres' to 'changed-kubegres'")
		})
	})

})

type SpecServiceAccountNameTest struct {
	keepCreatedResourcesForNextTest bool
	kubegresResource                *postgresv1.Kubegres
	dbQueryTestCases                testcases.DbQueryTestCases
	resourceCreator                 util.TestResourceCreator
	resourceRetriever               util.TestResourceRetriever
}

func (r *SpecServiceAccountNameTest) givenNewKubegresSpecIsSetTo(serviceAccountName string, specNbreReplicas int32) {
	r.kubegresResource = resourceConfigs.LoadKubegresYaml()
	r.kubegresResource.Spec.Replicas = &specNbreReplicas
	if serviceAccountName != "" {
		r.kubegresResource.Spec.ServiceAccountName = serviceAccountName
	}
}

func (r *SpecServiceAccountNameTest) givenExistingKubegresSpecIsSetTo(serviceAccountName string) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	r.kubegresResource.Spec.ServiceAccountName = serviceAccountName
}

func (r *SpecServiceAccountNameTest) whenKubegresIsCreated() {
	r.resourceCreator.CreateKubegres(r.kubegresResource)
}

func (r *SpecServiceAccountNameTest) whenKubernetesIsUpdated() {
	r.resourceCreator.UpdateResource(r.kubegresResource, "Kubegres")
}

func (r *SpecServiceAccountNameTest) thenPodsStatesShouldBe(serviceAccountName string, nbrePrimary, nbreReplicas int) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		for _, resource := range kubegresResources.Resources {
			currentServiceAccountName := resource.Pod.Spec.ServiceAccountName
			if currentServiceAccountName != serviceAccountName {
				log.Println("Pod '" + resource.Pod.Name + "' doesn't have the expected serviceAccountName: '" + serviceAccountName + "'. " +
					"Current value: '" + currentServiceAccountName + "'. Waiting...")
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

func (r *SpecServiceAccountNameTest) thenDeployedKubegresSpecShouldBeSetTo(serviceAccountName string) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	Expect(r.kubegresResource.Spec.ServiceAccountName).Should(Equal(serviceAccountName))
}

func (r *SpecServiceAccountNameTest) whenServiceAccountIsCreated(serviceAccountName string) {
	r.resourceCreator.CreateServiceAccount(serviceAccountName)
}
