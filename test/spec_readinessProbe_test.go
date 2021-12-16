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
	postgresv1 "reactive-tech.io/kubegres/api/v1"
	"reactive-tech.io/kubegres/test/resourceConfigs"
	"reactive-tech.io/kubegres/test/util"
	"reactive-tech.io/kubegres/test/util/testcases"
)

var _ = Describe("Setting Kubegres spec 'readinessProbe'", func() {

	var test = SpecReadinessProbeTest{}

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

	Context("GIVEN new Kubegres is created with spec 'readinessProbe' set to a value and spec 'replica' set to 3 and later 'readinessProbe' is updated to a new value", func() {

		It("GIVEN new Kubegres is created with spec 'readinessProbe' set to a value and spec 'replica' set to 3 THEN 1 primary and 2 replica should be created with spec 'readinessProbe' set the value", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'readinessProbe' set to a value and spec 'replica' set to 3")

			readinessProbe := test.givenReadinessProbe1()

			test.givenNewKubegresSpecIsSetTo(readinessProbe, 3)

			test.whenKubegresIsCreated()

			test.thenStatefulSetStatesShouldBe(readinessProbe, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(readinessProbe)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.keepCreatedResourcesForNextTest = true

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'readinessProbe' set to a value and spec 'replica' set to 3'")
		})

		It("GIVEN existing Kubegres is updated with spec 'readinessProbe' set to a new value THEN 1 primary and 2 replica should be re-deployed with spec 'readinessProbe' set the new value", func() {

			log.Print("START OF: Test 'GIVEN existing Kubegres is updated with spec 'readinessProbe' set to a new value")

			newReadinessProbe := test.givenReadinessProbe2()

			test.givenExistingKubegresSpecIsSetTo(newReadinessProbe)

			test.whenKubernetesIsUpdated()

			test.thenStatefulSetStatesShouldBe(newReadinessProbe, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(newReadinessProbe)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN existing Kubegres is updated with spec 'readinessProbe' set to a new value")
		})
	})

})

type SpecReadinessProbeTest struct {
	keepCreatedResourcesForNextTest bool
	kubegresResource                *postgresv1.Kubegres
	dbQueryTestCases                testcases.DbQueryTestCases
	resourceCreator                 util.TestResourceCreator
	resourceRetriever               util.TestResourceRetriever
}

func (r *SpecReadinessProbeTest) whenKubernetesIsUpdated() {
	r.resourceCreator.UpdateResource(r.kubegresResource, "Kubegres")
}

func (r *SpecReadinessProbeTest) givenReadinessProbe1() *v12.Probe {
	command := []string{"sh", "-c", "exec pg_isready -U postgres -h $POD_IP"}
	execAction := &v12.ExecAction{Command: command}
	handler := v12.Handler{Exec: execAction}
	return &v12.Probe{
		Handler:             handler,
		InitialDelaySeconds: int32(6),
		TimeoutSeconds:      int32(5),
		PeriodSeconds:       int32(10),
		SuccessThreshold:    int32(1),
		FailureThreshold:    int32(6),
	}
}

func (r *SpecReadinessProbeTest) givenReadinessProbe2() *v12.Probe {
	command := []string{"sh", "-c", "exec pg_isready -U postgres -h $POD_IP"}
	execAction := &v12.ExecAction{Command: command}
	handler := v12.Handler{Exec: execAction}
	return &v12.Probe{
		Handler:             handler,
		InitialDelaySeconds: int32(7),
		TimeoutSeconds:      int32(10),
		PeriodSeconds:       int32(20),
		SuccessThreshold:    int32(1),
		FailureThreshold:    int32(6),
	}
}

func (r *SpecReadinessProbeTest) givenNewKubegresSpecIsSetTo(readinessProbe *v12.Probe, specNbreReplicas int32) {
	r.kubegresResource = resourceConfigs.LoadKubegresYaml()
	r.kubegresResource.Spec.ReadinessProbe = readinessProbe
	r.kubegresResource.Spec.Replicas = &specNbreReplicas
}

func (r *SpecReadinessProbeTest) givenExistingKubegresSpecIsSetTo(readinessProbe *v12.Probe) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	r.kubegresResource.Spec.ReadinessProbe = readinessProbe
}

func (r *SpecReadinessProbeTest) whenKubegresIsCreated() {
	r.resourceCreator.CreateKubegres(r.kubegresResource)
}

func (r *SpecReadinessProbeTest) thenStatefulSetStatesShouldBe(expectedProbe *v12.Probe, nbrePrimary, nbreReplicas int) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		for _, resource := range kubegresResources.Resources {
			currentProbe := resource.StatefulSet.Spec.Template.Spec.Containers[0].ReadinessProbe

			if !reflect.DeepEqual(currentProbe, expectedProbe) {
				log.Println("StatefulSet '" + resource.StatefulSet.Name + "' doesn't have the expected spec 'readinessProbe': " + expectedProbe.String() + " " +
					"Current value: '" + currentProbe.String() + "'. Waiting...")
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

func (r *SpecReadinessProbeTest) thenDeployedKubegresSpecShouldBeSetTo(expectedProbe *v12.Probe) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	currentProbe := r.kubegresResource.Spec.ReadinessProbe
	Expect(currentProbe).Should(Equal(expectedProbe))
}
