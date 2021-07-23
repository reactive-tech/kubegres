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

const (
	kubegresOne = "kubegres-one"
	kubegresTwo = "kubegres-two"
)

var _ = Describe("Testing when there are 2 different Kubegres instances running in same namespace", func() {

	var test = ManyDifferentKubegresInstancesTest{}

	BeforeEach(func() {
		//Skip("Temporarily skipping test")

		namespace := resourceConfigs.DefaultNamespace
		test.resourceRetriever = util.CreateTestResourceRetriever(k8sClientTest, namespace)
		test.resourceCreator = util.CreateTestResourceCreator(k8sClientTest, test.resourceRetriever, namespace)
		test.kubegresOneDbQueryTestCases = testcases.InitDbQueryTestCasesWithNodePorts(test.resourceCreator, kubegresOne, resourceConfigs.ServiceToSqlQueryPrimaryDbNodePort, resourceConfigs.ServiceToSqlQueryReplicaDbNodePort)
		test.kubegresTwoDbQueryTestCases = testcases.InitDbQueryTestCasesWithNodePorts(test.resourceCreator, kubegresTwo, resourceConfigs.ServiceToSqlQueryPrimaryDbNodePort+2, resourceConfigs.ServiceToSqlQueryReplicaDbNodePort+2)
	})

	AfterEach(func() {
		test.resourceCreator.DeleteAllTestResources()
	})

	Context("GIVEN new Kubegres is created with spec 'replica' set to 2", func() {

		It("THEN 1 primary and 2 replica should be created", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'replica' set to 2'")

			kubegresOneResource := test.givenNewKubegresSpecIsSetTo(kubegresOne, 2)
			kubegresTwoResource := test.givenNewKubegresSpecIsSetTo(kubegresTwo, 3)

			test.whenKubegresIsCreated(kubegresOneResource)
			test.whenKubegresIsCreated(kubegresTwoResource)

			test.thenPodsStatesShouldBe(kubegresOne, 1, 1)
			test.thenPodsStatesShouldBe(kubegresTwo, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(kubegresOne, 2)
			test.thenDeployedKubegresSpecShouldBeSetTo(kubegresTwo, 3)

			test.kubegresOneDbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.kubegresOneDbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.kubegresTwoDbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.kubegresTwoDbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'replica' set to 2'")
		})

	})

})

type ManyDifferentKubegresInstancesTest struct {
	kubegresOneDbQueryTestCases testcases.DbQueryTestCases
	kubegresTwoDbQueryTestCases testcases.DbQueryTestCases
	resourceCreator             util.TestResourceCreator
	resourceRetriever           util.TestResourceRetriever
}

func (r *ManyDifferentKubegresInstancesTest) givenNewKubegresSpecIsSetTo(kubegresName string, specNbreReplicas int32) *postgresv1.Kubegres {
	kubegresResource := resourceConfigs.LoadKubegresYaml()
	kubegresResource.Name = kubegresName
	kubegresResource.Spec.Replicas = &specNbreReplicas
	return kubegresResource
}

func (r *ManyDifferentKubegresInstancesTest) givenExistingKubegresSpecIsSetTo(kubegresName string, specNbreReplicas int32) {
	var err error
	kubegresResource, err := r.resourceRetriever.GetKubegresByName(kubegresName)

	if err != nil {
		log.Println("Error while getting Kubegres resource with name '"+kubegresName+"' : ", err)
		Expect(err).Should(Succeed())
		return
	}

	kubegresResource.Spec.Replicas = &specNbreReplicas
}

func (r *ManyDifferentKubegresInstancesTest) whenKubegresIsCreated(kubegresResource *postgresv1.Kubegres) {
	r.resourceCreator.CreateKubegres(kubegresResource)
}

func (r *ManyDifferentKubegresInstancesTest) whenKubernetesIsUpdated(kubegresResource *postgresv1.Kubegres) {
	r.resourceCreator.UpdateResource(kubegresResource, "Kubegres")
}

func (r *ManyDifferentKubegresInstancesTest) thenErrorEventShouldBeLogged() {
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

func (r *ManyDifferentKubegresInstancesTest) thenPodsStatesShouldBe(kubegresName string, nbrePrimary, nbreReplicas int) bool {
	return Eventually(func() bool {

		pods, err := r.resourceRetriever.GetKubegresResourcesByName(kubegresName)
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving pods of Kubegres '" + kubegresName + "'")
			return false
		}

		if pods.AreAllReady &&
			pods.NbreDeployedPrimary == nbrePrimary &&
			pods.NbreDeployedReplicas == nbreReplicas {

			time.Sleep(resourceConfigs.TestRetryInterval)
			log.Println("Deployed and Ready StatefulSets check successful")
			return true
		}

		return false

	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *ManyDifferentKubegresInstancesTest) thenDeployedKubegresSpecShouldBeSetTo(kubegresName string, specNbreReplicas int32) {
	var err error
	kubegresResource, err := r.resourceRetriever.GetKubegresByName(kubegresName)

	if err != nil {
		log.Println("Error while getting Kubegres resource with name '"+kubegresName+"' : ", err)
		Expect(err).Should(Succeed())
		return
	}

	Expect(*kubegresResource.Spec.Replicas).Should(Equal(specNbreReplicas))
}
