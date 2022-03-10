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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	postgresv1 "reactive-tech.io/kubegres/api/v1"
	"reactive-tech.io/kubegres/test/resourceConfigs"
	"reactive-tech.io/kubegres/test/util"
	"reactive-tech.io/kubegres/test/util/testcases"
	"reflect"
	"strconv"
	"time"
)

var _ = Describe("Setting Kubegres spec 'nodeSets[].affinity'", func() {

	var test = SpecNodeSetsAffinityTest{}

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

	Context("GIVEN new Kubegres is created with spec 'nodeSets[].affinity' set to a value and spec 'replica' set to 3 and later 'nodeSets[].affinity' is updated to a new value", func() {

		It("GIVEN new Kubegres is created with spec 'nodeSets[].affinity' set to a value and spec 'replica' set to 3 THEN 1 primary and 2 replica should be created with spec 'nodeSets[].affinity' set the value", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'nodeSets[].affinity' set to a value and spec 'replica' set to 3")

			nodeSets := test.givenNodeSetsWithAffinity1()

			test.givenNewKubegresSpecIsSetTo(nodeSets)

			test.whenKubegresIsCreated()

			test.thenStatefulSetStatesShouldBe(nodeSets, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(nodeSets)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.keepCreatedResourcesForNextTest = true

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'nodeSets[].affinity' set to a value and spec 'replica' set to 3'")
		})

		It("GIVEN existing Kubegres is updated with spec 'nodeSets[].affinity' set to a new value THEN 1 primary and 2 replica should be re-deployed with spec 'nodeSets[].affinity' set the new value", func() {

			log.Print("START OF: Test 'GIVEN existing Kubegres is updated with spec 'scheduler.affinity' set to a new value")

			newNodeSets := test.givenNodeSetsWithAffinity2()

			test.givenExistingKubegresSpecIsSetTo(newNodeSets)

			test.whenKubernetesIsUpdated()

			test.thenStatefulSetStatesShouldBe(newNodeSets, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(newNodeSets)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN existing Kubegres is updated with spec 'scheduler.affinity' set to a new value")
		})
	})
})

type SpecNodeSetsAffinityTest struct {
	keepCreatedResourcesForNextTest bool
	kubegresResource                *postgresv1.Kubegres
	dbQueryTestCases                testcases.DbQueryTestCases
	resourceCreator                 util.TestResourceCreator
	resourceRetriever               util.TestResourceRetriever
}

func (r *SpecNodeSetsAffinityTest) givenNewKubegresSpecIsSetToNil() {
	r.kubegresResource = resourceConfigs.LoadKubegresYaml()
	r.kubegresResource.Spec.Replicas = nil
	r.kubegresResource.Spec.NodeSets = nil
}

func (r *SpecNodeSetsAffinityTest) givenNewKubegresSpecWithReplicasAndNodeSets() {
	replicas := int32(2)
	r.kubegresResource = resourceConfigs.LoadKubegresYaml()
	r.kubegresResource.Spec.Replicas = &replicas
	r.kubegresResource.Spec.NodeSets = []postgresv1.KubegresNodeSet{
		{
			Name: "no-affinity-a",
		},
	}
}

func (r *SpecNodeSetsAffinityTest) givenNewKubegresSpecWithEmptyNodeSets() {
	r.kubegresResource = resourceConfigs.LoadKubegresYaml()
	r.kubegresResource.Spec.Replicas = nil
	r.kubegresResource.Spec.NodeSets = []postgresv1.KubegresNodeSet{}
}

func (r *SpecNodeSetsAffinityTest) givenDefaultAffinity() *v12.Affinity {

	resourceName := resourceConfigs.KubegresResourceName

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

func (r *SpecNodeSetsAffinityTest) givenNodeSetsWithAffinity1() []postgresv1.KubegresNodeSet {

	resourceName := resourceConfigs.KubegresResourceName

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

	return []postgresv1.KubegresNodeSet{
		{
			Name: "affinity-a",
		},
		{
			Name: "affinity-b",
		},
		{
			Name: "affinity-c",
			Affinity: &v12.Affinity{
				PodAntiAffinity: &v12.PodAntiAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []v12.WeightedPodAffinityTerm{weightedPodAffinityTerm},
				},
			},
		},
	}
}

func (r *SpecNodeSetsAffinityTest) givenNodeSetsWithAffinity2() []postgresv1.KubegresNodeSet {

	resourceName := resourceConfigs.KubegresResourceName

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

	return []postgresv1.KubegresNodeSet{
		{
			Name: "affinity-a",
		},
		{
			Name: "affinity-b",
			Affinity: &v12.Affinity{
				PodAntiAffinity: &v12.PodAntiAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []v12.WeightedPodAffinityTerm{weightedPodAffinityTerm},
				},
			},
		},
		{
			Name: "affinity-c",
			Affinity: &v12.Affinity{
				PodAntiAffinity: &v12.PodAntiAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []v12.WeightedPodAffinityTerm{weightedPodAffinityTerm},
				},
			},
		},
	}
}

func (r *SpecNodeSetsAffinityTest) givenNewKubegresSpecIsSetTo(nodeSets []postgresv1.KubegresNodeSet) {
	r.kubegresResource = resourceConfigs.LoadKubegresYaml()
	r.kubegresResource.Spec.NodeSets = nodeSets
	r.kubegresResource.Spec.Replicas = nil
}

func (r *SpecNodeSetsAffinityTest) givenExistingKubegresSpecIsSetTo(nodeSets []postgresv1.KubegresNodeSet) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	r.kubegresResource.Spec.NodeSets = nodeSets
}

func (r *SpecNodeSetsAffinityTest) whenKubegresIsCreated() {
	r.resourceCreator.CreateKubegres(r.kubegresResource)
}

func (r *SpecNodeSetsAffinityTest) whenKubernetesIsUpdated() {
	r.resourceCreator.UpdateResource(r.kubegresResource, "Kubegres")
}

func (r *SpecNodeSetsAffinityTest) thenReplicasErrorEventShouldBeLogged() {
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

func (r *SpecNodeSetsAffinityTest) thenNodeSetsErrorEventShouldBeLogged() {
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

func (r *SpecNodeSetsAffinityTest) thenMutuallyExclusiveErrorEventShouldBeLogged() {
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

func (r *SpecNodeSetsAffinityTest) thenStatefulSetStatesShouldBe(expectedNodeSets []postgresv1.KubegresNodeSet, nbrePrimary, nbreReplicas int) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		for _, resource := range kubegresResources.Resources {
			currentAffinity := resource.StatefulSet.Spec.Template.Spec.Affinity
			statefulSetIndex, _ := strconv.ParseInt(resource.Pod.Metadata.Labels["index"], 10, 32)
			expectedAffinity := expectedNodeSets[statefulSetIndex-1].Affinity
			if expectedAffinity == nil {
				expectedAffinity = r.givenDefaultAffinity()
			}
			if !reflect.DeepEqual(currentAffinity, expectedAffinity) {
				log.Println("StatefulSet '" + resource.StatefulSet.Name + "' doesn't have the expected spec 'affinity': " + expectedAffinity.String() + " " +
					"Current value: '" + currentAffinity.String() + "'. Waiting...")
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

func (r *SpecNodeSetsAffinityTest) thenDeployedKubegresSpecShouldBeSetTo(expectedNodeSets []postgresv1.KubegresNodeSet) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	nodeSets := r.kubegresResource.Spec.NodeSets
	Expect(nodeSets).Should(Equal(expectedNodeSets))
}
