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
	resourceConfigs2 "reactive-tech.io/kubegres/internal/test/resourceConfigs"
	util2 "reactive-tech.io/kubegres/internal/test/util"
	"reactive-tech.io/kubegres/internal/test/util/testcases"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v12 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/pointer"

	postgresv1 "reactive-tech.io/kubegres/api/v1"
)

var _ = Describe("Setting Kubegres spec 'securityContext'", func() {

	var test = SpecsecurityContextTest{}

	BeforeEach(func() {
		Skip("Temporarily skipping test")

		namespace := resourceConfigs2.DefaultNamespace
		test.resourceRetriever = util2.CreateTestResourceRetriever(k8sClientTest, namespace)
		test.resourceCreator = util2.CreateTestResourceCreator(k8sClientTest, test.resourceRetriever, namespace)
		test.dbQueryTestCases = testcases.InitDbQueryTestCases(test.resourceCreator, resourceConfigs2.KubegresResourceName)
	})

	AfterEach(func() {
		if !test.keepCreatedResourcesForNextTest {
			test.resourceCreator.DeleteAllTestResources()
		} else {
			test.keepCreatedResourcesForNextTest = false
		}
	})

	Context("GIVEN new Kubegres is created without spec 'securityContext' and with spec 'replica' set to 3", func() {

		It("THEN 1 primary and 2 replica should be created with 'securityContext' set to nil", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created without spec 'securityContext' and with spec 'replica' set to 3'")

			test.givenNewKubegresSpecIsWithoutsecurityContext(3)

			test.whenKubegresIsCreated()

			test.thenStatefulSetStatesShouldBeWithoutsecurityContext(1, 2)

			test.thenDeployedKubegresSpecShouldWithoutsecurityContext()

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN new Kubegres is created without spec 'securityContext' and with spec 'replica' set to 3'")
		})
	})

	Context("GIVEN new Kubegres is created with spec 'securityContext' set to a value and spec 'replica' set to 3 and later 'securityContext' is updated to a new value", func() {

		It("GIVEN new Kubegres is created with spec 'securityContext' set to a value and spec 'replica' set to 3 THEN 1 primary and 2 replica should be created with spec 'securityContext' set the value", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'securityContext' set to a value and spec 'replica' set to 3")

			securityContext := test.givensecurityContext1()

			test.givenNewKubegresSpecIsSetTo(securityContext, 3)

			test.whenKubegresIsCreated()

			test.thenStatefulSetStatesShouldBe(securityContext, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(securityContext)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.keepCreatedResourcesForNextTest = true

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'securityContext' set to a value and spec 'replica' set to 3'")
		})

		It("GIVEN existing Kubegres is updated with spec 'securityContext' set to a new value THEN 1 primary and 2 replica should be re-deployed with spec 'securityContext' set the new value", func() {

			log.Print("START OF: Test 'GIVEN existing Kubegres is updated with spec 'securityContext' set to a new value")

			newsecurityContext := test.givensecurityContext2()

			test.givenExistingKubegresSpecIsSetTo(newsecurityContext)

			test.whenKubernetesIsUpdated()

			test.thenStatefulSetStatesShouldBe(newsecurityContext, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(newsecurityContext)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN existing Kubegres is updated with spec 'securityContext' set to a new value")
		})
	})

})

type SpecsecurityContextTest struct {
	keepCreatedResourcesForNextTest bool
	kubegresResource                *postgresv1.Kubegres
	dbQueryTestCases                testcases.DbQueryTestCases
	resourceCreator                 util2.TestResourceCreator
	resourceRetriever               util2.TestResourceRetriever
}

func (r *SpecsecurityContextTest) givensecurityContext1() *v12.PodSecurityContext {
	return &v12.PodSecurityContext{
		RunAsUser:    pointer.Int64Ptr(1000),
		RunAsGroup:   pointer.Int64Ptr(3000),
		RunAsNonRoot: pointer.BoolPtr(true),
		FSGroup:      pointer.Int64Ptr(2000),
	}
}

func (r *SpecsecurityContextTest) givensecurityContext2() *v12.PodSecurityContext {
	return &v12.PodSecurityContext{
		RunAsUser:    pointer.Int64Ptr(1000),
		RunAsNonRoot: pointer.BoolPtr(true),
	}
}

func (r *SpecsecurityContextTest) givenNewKubegresSpecIsWithoutsecurityContext(specNbreReplicas int32) {
	r.kubegresResource = resourceConfigs2.LoadKubegresYaml()
	r.kubegresResource.Spec.SecurityContext = &v12.PodSecurityContext{}
	r.kubegresResource.Spec.Replicas = &specNbreReplicas
}

func (r *SpecsecurityContextTest) givenNewKubegresSpecIsSetTo(securityContext *v12.PodSecurityContext, specNbreReplicas int32) {
	r.kubegresResource = resourceConfigs2.LoadKubegresYaml()
	r.kubegresResource.Spec.SecurityContext = securityContext
	r.kubegresResource.Spec.Replicas = &specNbreReplicas
}

func (r *SpecsecurityContextTest) givenExistingKubegresSpecIsSetTo(securityContext *v12.PodSecurityContext) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	r.kubegresResource.Spec.SecurityContext = securityContext
}

func (r *SpecsecurityContextTest) whenKubegresIsCreated() {
	r.resourceCreator.CreateKubegres(r.kubegresResource)
}

func (r *SpecsecurityContextTest) whenKubernetesIsUpdated() {
	r.resourceCreator.UpdateResource(r.kubegresResource, "Kubegres")
}

func (r *SpecsecurityContextTest) thenStatefulSetStatesShouldBeWithoutsecurityContext(nbrePrimary, nbreReplicas int) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		for _, resource := range kubegresResources.Resources {
			currentSecurityContext := resource.StatefulSet.Spec.Template.Spec.SecurityContext
			emptyResources := &v12.PodSecurityContext{}
			if !reflect.DeepEqual(currentSecurityContext, emptyResources) {

				log.Println("StatefulSet '" + resource.StatefulSet.Name + "' doesn't have the expected spec securityContext' which should be nil. " +
					"Current value: '" + currentSecurityContext.String() + "'. Waiting...")
				return false
			}
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

func (r *SpecsecurityContextTest) thenStatefulSetStatesShouldBe(expectedSecurityContext *v12.PodSecurityContext, nbrePrimary, nbreReplicas int) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		for _, resource := range kubegresResources.Resources {
			currentSecurityContext := resource.StatefulSet.Spec.Template.Spec.SecurityContext

			if !reflect.DeepEqual(currentSecurityContext, expectedSecurityContext) {
				{
					log.Println("StatefulSet '" + resource.StatefulSet.Name + "' doesn't have the expected spec securityContext: " + expectedSecurityContext.String() + " " +
						"Current value: '" + currentSecurityContext.String() + "'. Waiting...")
					return false
				}
			}
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

func (r *SpecsecurityContextTest) thenDeployedKubegresSpecShouldWithoutsecurityContext() {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()
	emptyResources := &v12.PodSecurityContext{}
	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	Expect(r.kubegresResource.Spec.SecurityContext).Should(Equal(emptyResources))
}

func (r *SpecsecurityContextTest) thenDeployedKubegresSpecShouldBeSetTo(expectedsecurityContext *v12.PodSecurityContext) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	Expect(r.kubegresResource.Spec.SecurityContext).Should(Equal(expectedsecurityContext))
}
