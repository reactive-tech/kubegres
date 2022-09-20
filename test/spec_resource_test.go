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
	"log"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v12 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	postgresv1 "reactive-tech.io/kubegres/api/v1"
	"reactive-tech.io/kubegres/test/resourceConfigs"
	"reactive-tech.io/kubegres/test/util"
	"reactive-tech.io/kubegres/test/util/testcases"
)

var _ = Describe("Setting Kubegres spec 'resource'", func() {

	var test = SpecResourceTest{}

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

	Context("GIVEN new Kubegres is created without spec 'resources' and with spec 'replica' set to 3", func() {

		It("THEN 1 primary and 2 replica should be created without 'resources' values ", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created without spec 'resources' and with spec 'replica' set to 3'")

			test.givenNewKubegresSpecIsWithoutResources(3)

			test.whenKubegresIsCreated()

			test.thenStatefulSetStatesShouldBeWithoutResources(1, 2)

			test.thenDeployedKubegresSpecShouldWithoutResource()

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN new Kubegres is created without spec 'resources' and with spec 'replica' set to 3'")
		})
	})

	Context("GIVEN new Kubegres is created with spec 'resources' set to a value and spec 'replica' set to 3 and later 'resources' is updated to a new value", func() {

		It("GIVEN new Kubegres is created with spec 'resources' set to a value and spec 'replica' set to 3 THEN 1 primary and 2 replica should be created with spec 'resources' set the value", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'resources' set to a value and spec 'replica' set to 3")

			resources := test.givenResources("2", "2Gi", "1", "1Gi")

			test.givenNewKubegresSpecIsSetTo(resources, 3)

			test.whenKubegresIsCreated()

			test.thenStatefulSetStatesShouldBe(resources, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(resources)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.keepCreatedResourcesForNextTest = true

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'resources' set to a value and spec 'replica' set to 3'")
		})

		It("GIVEN existing Kubegres is updated with spec 'resources' set to a new value THEN 1 primary and 2 replica should be re-deployed with spec 'resources' set the new value", func() {

			log.Print("START OF: Test 'GIVEN existing Kubegres is updated with spec 'resources' set to a new value")

			newResources := test.givenResources("2", "4Gi", "500m", "512Mi")

			test.givenExistingKubegresSpecIsSetTo(newResources)

			test.whenKubernetesIsUpdated()

			test.thenStatefulSetStatesShouldBe(newResources, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(newResources)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN existing Kubegres is updated with spec 'resources' set to a new value")
		})
	})

})

type SpecResourceTest struct {
	keepCreatedResourcesForNextTest bool
	kubegresResource                *postgresv1.Kubegres
	dbQueryTestCases                testcases.DbQueryTestCases
	resourceCreator                 util.TestResourceCreator
	resourceRetriever               util.TestResourceRetriever
}

func (r *SpecResourceTest) whenKubernetesIsUpdated() {
	r.resourceCreator.UpdateResource(r.kubegresResource, "Kubegres")
}

func (r *SpecResourceTest) givenResources(cpuLimit, memLimit, cpuReq, memReq string) v12.ResourceRequirements {
	return v12.ResourceRequirements{
		Limits: v12.ResourceList{
			"cpu":    resource.MustParse(cpuLimit),
			"memory": resource.MustParse(memLimit),
		},
		Requests: v12.ResourceList{
			"cpu":    resource.MustParse(cpuReq),
			"memory": resource.MustParse(memReq),
		},
	}
}

func (r *SpecResourceTest) givenNewKubegresSpecIsWithoutResources(specNbreReplicas int32) {
	r.kubegresResource = resourceConfigs.LoadKubegresYaml()
	r.kubegresResource.Spec.Resources = v12.ResourceRequirements{}
	r.kubegresResource.Spec.Replicas = &specNbreReplicas
}

func (r *SpecResourceTest) givenNewKubegresSpecIsSetTo(resources v12.ResourceRequirements, specNbreReplicas int32) {
	r.kubegresResource = resourceConfigs.LoadKubegresYaml()
	r.kubegresResource.Spec.Resources = resources
	r.kubegresResource.Spec.Replicas = &specNbreReplicas
}

func (r *SpecResourceTest) givenExistingKubegresSpecIsSetTo(resources v12.ResourceRequirements) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	r.kubegresResource.Spec.Resources = resources
}

func (r *SpecResourceTest) whenKubegresIsCreated() {
	r.resourceCreator.CreateKubegres(r.kubegresResource)
}

func (r *SpecResourceTest) thenStatefulSetStatesShouldBeWithoutResources(nbrePrimary, nbreReplicas int) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		for _, resource := range kubegresResources.Resources {
			currentResources := resource.StatefulSet.Spec.Template.Spec.Containers[0].Resources
			emptyResources := v12.ResourceRequirements{}

			if !reflect.DeepEqual(currentResources, emptyResources) {
				log.Println("StatefulSet '" + resource.StatefulSet.Name + emptyResources.String() + "  ' doesn't have the expected spec 'resources' which should be the default one. " +
					"Current value: '" + currentResources.String() + "'. Waiting...")
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

func (r *SpecResourceTest) thenStatefulSetStatesShouldBe(expectedResources v12.ResourceRequirements, nbrePrimary, nbreReplicas int) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		for _, resource := range kubegresResources.Resources {
			currentResources := resource.StatefulSet.Spec.Template.Spec.Containers[0].Resources

			if !reflect.DeepEqual(currentResources, expectedResources) {
				log.Println("StatefulSet '" + resource.StatefulSet.Name + "' doesn't have the expected spec 'resources': " + expectedResources.String() + " " +
					"Current value: '" + currentResources.String() + "'. Waiting...")
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

func (r *SpecResourceTest) thenDeployedKubegresSpecShouldWithoutResource() {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}
	currentResources := r.kubegresResource.Spec.Resources
	emptyResources := v12.ResourceRequirements{}
	Expect(currentResources).Should(Equal(emptyResources))
}

func (r *SpecResourceTest) thenDeployedKubegresSpecShouldBeSetTo(expectedResources v12.ResourceRequirements) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	currentResources := r.kubegresResource.Spec.Resources
	Expect(currentResources).Should(Equal(expectedResources))
}
