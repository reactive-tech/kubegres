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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	postgresv1 "reactive-tech.io/kubegres/api/v1"
	resourceConfigs2 "reactive-tech.io/kubegres/internal/test/resourceConfigs"
	util2 "reactive-tech.io/kubegres/internal/test/util"
	"reactive-tech.io/kubegres/internal/test/util/testcases"
	"reflect"
	"time"
)

var _ = Describe("Setting Kubegres spec 'scheduler.affinity'", func() {

	var test = SpecAffinityTest{}

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

	Context("GIVEN new Kubegres is created without spec 'scheduler.affinity' and with spec 'replica' set to 3", func() {

		It("THEN 1 primary and 2 replica should be created with 'scheduler.affinity' set to default value", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created without spec 'scheduler.affinity' and with spec 'replica' set to 3'")

			test.givenNewKubegresSpecIsWithoutAffinity(3)

			test.whenKubegresIsCreated()

			test.thenStatefulSetStatesShouldBeWithDefaultAffinity(1, 2)

			test.thenDeployedKubegresSpecShouldWithDefaultAffinity()

			test.thenEventShouldBeLoggedSayingAffinityIsSetToDefaultValue()

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN new Kubegres is created without spec 'scheduler.affinity' and with spec 'replica' set to 3'")
		})
	})

	Context("GIVEN new Kubegres is created with spec 'scheduler.affinity' set to a value and spec 'replica' set to 3 and later 'scheduler.affinity' is updated to a new value", func() {

		It("GIVEN new Kubegres is created with spec 'scheduler.affinity' set to a value and spec 'replica' set to 3 THEN 1 primary and 2 replica should be created with spec 'scheduler.affinity' set the value", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'scheduler.affinity' set to a value and spec 'replica' set to 3")

			affinity := test.givenAffinity1()

			test.givenNewKubegresSpecIsSetTo(affinity, 3)

			test.whenKubegresIsCreated()

			test.thenStatefulSetStatesShouldBe(affinity, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(affinity)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.keepCreatedResourcesForNextTest = true

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'scheduler.affinity' set to a value and spec 'replica' set to 3'")
		})

		It("GIVEN existing Kubegres is updated with spec 'scheduler.affinity' set to a new value THEN 1 primary and 2 replica should be re-deployed with spec 'scheduler.affinity' set the new value", func() {

			log.Print("START OF: Test 'GIVEN existing Kubegres is updated with spec 'scheduler.affinity' set to a new value")

			newAffinity := test.givenAffinity2()

			test.givenExistingKubegresSpecIsSetTo(newAffinity)

			test.whenKubernetesIsUpdated()

			test.thenStatefulSetStatesShouldBe(newAffinity, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(newAffinity)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN existing Kubegres is updated with spec 'scheduler.affinity' set to a new value")
		})
	})
})

type SpecAffinityTest struct {
	keepCreatedResourcesForNextTest bool
	kubegresResource                *postgresv1.Kubegres
	dbQueryTestCases                testcases.DbQueryTestCases
	resourceCreator                 util2.TestResourceCreator
	resourceRetriever               util2.TestResourceRetriever
}

func (r *SpecAffinityTest) givenDefaultAffinity() *v12.Affinity {

	resourceName := resourceConfigs2.KubegresResourceName

	weightedPodAffinityTerm := v12.WeightedPodAffinityTerm{
		Weight: 100,
		PodAffinityTerm: v12.PodAffinityTerm{
			LabelSelector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "app",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{resourceName},
					},
				},
			},
			TopologyKey: "kubernetes.io/hostname",
		},
	}

	return &v12.Affinity{
		PodAntiAffinity: &v12.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []v12.WeightedPodAffinityTerm{weightedPodAffinityTerm},
		},
	}
}

func (r *SpecAffinityTest) givenAffinity1() *v12.Affinity {

	resourceName := resourceConfigs2.KubegresResourceName

	weightedPodAffinityTerm := v12.WeightedPodAffinityTerm{
		Weight: 90,
		PodAffinityTerm: v12.PodAffinityTerm{
			LabelSelector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "app",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{resourceName},
					},
				},
			},
			TopologyKey: "kubernetes.io/hostname",
		},
	}

	return &v12.Affinity{
		PodAntiAffinity: &v12.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []v12.WeightedPodAffinityTerm{weightedPodAffinityTerm},
		},
	}
}

func (r *SpecAffinityTest) givenAffinity2() *v12.Affinity {

	resourceName := resourceConfigs2.KubegresResourceName

	weightedPodAffinityTerm := v12.WeightedPodAffinityTerm{
		Weight: 80,
		PodAffinityTerm: v12.PodAffinityTerm{
			LabelSelector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "app",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{resourceName},
					},
				},
			},
			TopologyKey: "kubernetes.io/hostname",
		},
	}

	return &v12.Affinity{
		PodAntiAffinity: &v12.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []v12.WeightedPodAffinityTerm{weightedPodAffinityTerm},
		},
	}
}

func (r *SpecAffinityTest) givenNewKubegresSpecIsWithoutAffinity(specNbreReplicas int32) {
	r.kubegresResource = resourceConfigs2.LoadKubegresYaml()
	r.kubegresResource.Spec.Scheduler.Affinity = nil
	r.kubegresResource.Spec.Replicas = &specNbreReplicas
}

func (r *SpecAffinityTest) givenNewKubegresSpecIsSetTo(affinity *v12.Affinity, specNbreReplicas int32) {
	r.kubegresResource = resourceConfigs2.LoadKubegresYaml()
	r.kubegresResource.Spec.Scheduler.Affinity = affinity
	r.kubegresResource.Spec.Replicas = &specNbreReplicas
}

func (r *SpecAffinityTest) givenExistingKubegresSpecIsSetTo(affinity *v12.Affinity) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	r.kubegresResource.Spec.Scheduler.Affinity = affinity
}

func (r *SpecAffinityTest) whenKubegresIsCreated() {
	r.resourceCreator.CreateKubegres(r.kubegresResource)
}

func (r *SpecAffinityTest) whenKubernetesIsUpdated() {
	r.resourceCreator.UpdateResource(r.kubegresResource, "Kubegres")
}

func (r *SpecAffinityTest) thenEventShouldBeLoggedSayingAffinityIsSetToDefaultValue() {

	expectedErrorEvent := util2.EventRecord{
		Eventtype: v12.EventTypeNormal,
		Reason:    "DefaultSpecValue",
		Message:   "A default value was set for a field in Kubegres YAML spec. 'spec.Affinity': New value: " + r.givenDefaultAffinity().String(),
	}
	Eventually(func() bool {
		_, err := r.resourceRetriever.GetKubegres()
		if err != nil {
			return false
		}
		return eventRecorderTest.CheckEventExist(expectedErrorEvent)

	}, resourceConfigs2.TestTimeout, resourceConfigs2.TestRetryInterval).Should(BeTrue())
}

func (r *SpecAffinityTest) thenStatefulSetStatesShouldBeWithDefaultAffinity(nbrePrimary, nbreReplicas int) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		for _, resource := range kubegresResources.Resources {
			affinity := resource.StatefulSet.Spec.Template.Spec.Affinity
			defaultAffinity := r.givenDefaultAffinity()

			if !reflect.DeepEqual(affinity, defaultAffinity) {
				log.Println("StatefulSet '" + resource.StatefulSet.Name + "' doesn't have the expected spec 'scheduler.affinity' which should be the default one. " +
					"Current value: '" + affinity.String() + "'. Waiting...")
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

func (r *SpecAffinityTest) thenStatefulSetStatesShouldBe(expectedAffinity *v12.Affinity, nbrePrimary, nbreReplicas int) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		for _, resource := range kubegresResources.Resources {
			currentAffinity := resource.StatefulSet.Spec.Template.Spec.Affinity

			if !reflect.DeepEqual(currentAffinity, expectedAffinity) {
				log.Println("StatefulSet '" + resource.StatefulSet.Name + "' doesn't have the expected spec 'scheduler.affinity': " + expectedAffinity.String() + " " +
					"Current value: '" + currentAffinity.String() + "'. Waiting...")
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

func (r *SpecAffinityTest) thenDeployedKubegresSpecShouldWithDefaultAffinity() {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	affinity := r.kubegresResource.Spec.Scheduler.Affinity
	defaultAffinity := r.givenDefaultAffinity()
	Expect(affinity).Should(Equal(defaultAffinity))
}

func (r *SpecAffinityTest) thenDeployedKubegresSpecShouldBeSetTo(expectedAffinity *v12.Affinity) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	affinity := r.kubegresResource.Spec.Scheduler.Affinity
	Expect(affinity).Should(Equal(expectedAffinity))
}
