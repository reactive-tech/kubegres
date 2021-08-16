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
	"reactive-tech.io/kubegres/controllers/ctx"
	"reactive-tech.io/kubegres/test/resourceConfigs"
	"reactive-tech.io/kubegres/test/util"
	"reactive-tech.io/kubegres/test/util/testcases"
	"strconv"
	"time"
)

var _ = Describe("Setting Kubegres spec 'port'", func() {

	var test = SpecPortTest{}

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

	Context("GIVEN new Kubegres is created without spec 'port' and with spec 'replica' set to 3", func() {

		It("THEN 1 primary and 2 replica should be created with port '5432' and a normal event should be logged", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created without spec 'port' and with spec 'replica' set to 3'")

			test.givenNewKubegresSpecIsSetTo(0, 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(ctx.DefaultContainerPortNumber, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(ctx.DefaultContainerPortNumber)

			test.thenPortOfKubegresPrimaryAndReplicaServicesShouldBeSetTo(ctx.DefaultContainerPortNumber)

			test.thenEventShouldBeLoggedSayingPortIsSetToDefaultValue()

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN new Kubegres is created without spec 'port' and with spec 'replica' set to 3'")
		})
	})

	Context("GIVEN new Kubegres is created with spec 'port' set to '5433' and spec 'replica' set to 3 and later 'port' is updated to '5434'", func() {

		It("GIVEN new Kubegres is created with spec 'port' set to '5433' and spec 'replica' set to 3 THEN 1 primary and 2 replica should be created with spec 'port' set to '5433", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'port' set to '5433' and spec 'replica' set to 3")

			test.givenNewKubegresSpecIsSetTo(5433, 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(5433, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(5433)

			test.thenPortOfKubegresPrimaryAndReplicaServicesShouldBeSetTo(5433)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.keepCreatedResourcesForNextTest = true

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'port' set to '5433' and spec 'replica' set to 3'")
		})

		It("GIVEN existing Kubegres is updated with spec 'port' set from '5433' to '5434' THEN 1 primary and 2 replica should be re-deployed with spec 'port' set to '5434'", func() {

			log.Print("START OF: Test 'GIVEN existing Kubegres is updated with spec 'port' set from '5433' to '5434'")

			test.givenExistingKubegresSpecIsSetTo(5434)

			test.whenKubernetesIsUpdated()

			test.thenPodsStatesShouldBe(5434, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(5434)

			test.thenPortOfKubegresPrimaryAndReplicaServicesShouldBeSetTo(5434)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.updateTestServicesPort(resourceConfigs.DbPort)

			log.Print("END OF: Test 'GIVEN existing Kubegres is updated with spec 'port' set from '5433' to '5434'")
		})
	})

})

type SpecPortTest struct {
	keepCreatedResourcesForNextTest bool
	kubegresResource                *postgresv1.Kubegres
	dbQueryTestCases                testcases.DbQueryTestCases
	resourceCreator                 util.TestResourceCreator
	resourceRetriever               util.TestResourceRetriever
}

func (r *SpecPortTest) givenNewKubegresSpecIsSetTo(port int32, specNbreReplicas int32) {
	r.kubegresResource = resourceConfigs.LoadKubegresYaml()
	r.kubegresResource.Spec.Port = port
	r.kubegresResource.Spec.Replicas = &specNbreReplicas

	if port > 0 {
		r.updateTestServicesPort(port)
	}
}

func (r *SpecPortTest) givenExistingKubegresSpecIsSetTo(port int32) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	r.kubegresResource.Spec.Port = port

	if port > 0 {
		r.updateTestServicesPort(port)
	}
}

func (r *SpecPortTest) whenKubegresIsCreated() {
	r.resourceCreator.CreateKubegres(r.kubegresResource)
}

func (r *SpecPortTest) whenKubernetesIsUpdated() {
	r.resourceCreator.UpdateResource(r.kubegresResource, "Kubegres")
}

func (r *SpecPortTest) thenEventShouldBeLoggedSayingPortIsSetToDefaultValue() {

	expectedErrorEvent := util.EventRecord{
		Eventtype: v12.EventTypeNormal,
		Reason:    "DefaultSpecValue",
		Message:   "A default value was set for a field in Kubegres YAML spec. 'spec.port': New value: " + strconv.Itoa(ctx.DefaultContainerPortNumber),
	}
	Eventually(func() bool {
		_, err := r.resourceRetriever.GetKubegres()
		if err != nil {
			return false
		}
		return eventRecorderTest.CheckEventExist(expectedErrorEvent)

	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *SpecPortTest) thenPodsStatesShouldBe(port int32, nbrePrimary, nbreReplicas int) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		for _, resource := range kubegresResources.Resources {
			currentPort := resource.Pod.Spec.Containers[0].Ports[0].ContainerPort
			if currentPort != port {
				log.Println("Pod '" + resource.Pod.Name + "' doesn't have the expected port: '" + strconv.Itoa(int(port)) + "'. " +
					"Current value: '" + strconv.Itoa(int(currentPort)) + "'. Waiting...")
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

func (r *SpecPortTest) thenDeployedKubegresSpecShouldBeSetTo(port int32) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	Expect(r.kubegresResource.Spec.Port).Should(Equal(port))
}

func (r *SpecPortTest) thenPortOfKubegresPrimaryAndReplicaServicesShouldBeSetTo(port int32) {

	Eventually(func() bool {

		primaryService, err := r.resourceRetriever.GetService(resourceConfigs.KubegresResourceName)
		if err != nil {
			log.Println("Error while getting Kubegres Primary Service resource : ", err)
			return false
		}

		replicaService, err := r.resourceRetriever.GetService(resourceConfigs.KubegresResourceName + "-replica")
		if err != nil {
			log.Println("Error while getting Kubegres Replica Service resource : ", err)
			return false
		}

		return primaryService.Spec.Ports[0].Port == port && replicaService.Spec.Ports[0].Port == port

	}, time.Second*20, time.Second*5).Should(BeTrue())
}

func (r *SpecPortTest) updateTestServicesPort(port int32) {

	serviceNameAllowingToSqlQueryPrimaryDb := r.resourceRetriever.GetServiceNameAllowingToSqlQueryDb(resourceConfigs.KubegresResourceName, true)
	primaryTestService, err := r.resourceRetriever.GetService(serviceNameAllowingToSqlQueryPrimaryDb)
	if err != nil {
		log.Println("Error while getting Primary Test service resource : ", err)
	}

	primaryTestService.Spec.Ports[0].Port = port
	r.resourceCreator.UpdateResource(primaryTestService, "Primary Test Service")

	serviceNameAllowingToSqlQueryReplicaDb := r.resourceRetriever.GetServiceNameAllowingToSqlQueryDb(resourceConfigs.KubegresResourceName, false)
	replicaTestService, err := r.resourceRetriever.GetService(serviceNameAllowingToSqlQueryReplicaDb)
	if err != nil {
		log.Println("Error while getting Replica Test service resource : ", err)
	}

	replicaTestService.Spec.Ports[0].Port = port
	r.resourceCreator.UpdateResource(replicaTestService, "Replica Test Service")
}
