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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"log"
	postgresv1 "reactive-tech.io/kubegres/api/v1"
	"reactive-tech.io/kubegres/controllers/spec/template"
	"reactive-tech.io/kubegres/test/resourceConfigs"
	"reactive-tech.io/kubegres/test/util"
	"reactive-tech.io/kubegres/test/util/testcases"
	"time"
)

const customAnnotation1Key = "linkerd.io/inject"
const customAnnotation1Value = "enabled"
const customAnnotation2Key = "toto.io/test"
const customAnnotation2Value = "disabled"

var _ = Describe("Creating Kubegres with custom annotations", func() {

	var test = CustomAnnotationTest{}

	BeforeEach(func() {
		//Skip("Temporarily skipping test")

		namespace := resourceConfigs.DefaultNamespace
		test.resourceRetriever = util.CreateTestResourceRetriever(k8sClientTest, namespace)
		test.resourceCreator = util.CreateTestResourceCreator(k8sClientTest, test.resourceRetriever, namespace)
		test.dbQueryTestCases = testcases.InitDbQueryTestCases(test.resourceCreator, resourceConfigs.KubegresResourceName)
		test.kubegresResource = resourceConfigs.LoadKubegresYaml()
	})

	AfterEach(func() {
		if !test.keepCreatedResourcesForNextTest {
			test.resourceCreator.DeleteAllTestResources()
		} else {
			test.keepCreatedResourcesForNextTest = false
		}
	})

	Context("GIVEN new Kubegres is created with custom annotations and with spec 'replica' set to 3", func() {

		// GIVEN new Kubegres is created with custom annotations and spec 'replica' set to 3 then

		It("GIVEN new Kubegres is created with with custom annotations and with spec 'replica' set to 3 THEN it should be deployed with StatefulSets and Pods containing the custom annotations AND 1 primary and 2 replica should be created", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with custom annotations and with spec 'replica' set to 3'")

			test.givenKubegresAnnotationIsSetTo(customAnnotation1Key, customAnnotation1Value)
			test.givenKubegresAnnotationIsSetTo(customAnnotation2Key, customAnnotation2Value)

			test.givenKubegresSpecIsSetTo(3)

			test.whenKubernetesIsCreated()

			test.thenPodsStatesShouldBe(1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(3)
			test.thenPodsAndStatefulSetsShouldHaveAnnotation(customAnnotation1Key, customAnnotation1Value)
			test.thenPodsAndStatefulSetsShouldHaveAnnotation(customAnnotation2Key, customAnnotation2Value)
			test.thenPodsAndStatefulSetsShouldNOTHaveAnnotationKey(template.KubegresInternalAnnotationKey)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.keepCreatedResourcesForNextTest = true

			log.Print("END OF: Test 'GIVEN new Kubegres is created with custom annotations and with spec 'replica' set to 3'")
		})

	})

})

type CustomAnnotationTest struct {
	keepCreatedResourcesForNextTest bool
	kubegresResource                *postgresv1.Kubegres
	dbQueryTestCases                testcases.DbQueryTestCases
	resourceCreator                 util.TestResourceCreator
	resourceRetriever               util.TestResourceRetriever
	resourceModifier                util.TestResourceModifier
}

func (r *CustomAnnotationTest) givenKubegresAnnotationIsSetTo(annotationKey string, annotationValue string) {
	r.resourceModifier.AppendAnnotation(annotationKey, annotationValue, r.kubegresResource)
}

func (r *CustomAnnotationTest) givenKubegresSpecIsSetTo(specNbreReplicas int32) {
	r.kubegresResource.Spec.Replicas = &specNbreReplicas
}

func (r *CustomAnnotationTest) whenKubernetesIsCreated() {
	r.resourceCreator.CreateKubegres(r.kubegresResource)
}

func (r *CustomAnnotationTest) thenPodsAndStatefulSetsShouldHaveAnnotation(annotationKey, annotationValue string) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		for _, kubegresResource := range kubegresResources.Resources {
			if kubegresResource.Pod.Metadata.Annotations[annotationKey] != annotationValue {
				log.Println("Pods do NOT contain the annotation '" + annotationKey + ":" + annotationValue + "'")
				return false
			}
			if kubegresResource.StatefulSet.Metadata.Annotations[annotationKey] != annotationValue {
				log.Println("StatefulSets do NOT contain the annotation '" + annotationKey + ":" + annotationValue + "'")
				return false
			}
		}

		log.Println("Pods and StatefulSets do contain the annotation '" + annotationKey + ":" + annotationValue + "'")
		return true

	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *CustomAnnotationTest) thenPodsAndStatefulSetsShouldNOTHaveAnnotationKey(annotationKey string) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		for _, kubegresResource := range kubegresResources.Resources {

			_, existsInPod := kubegresResource.Pod.Metadata.Annotations[annotationKey]
			if existsInPod {
				log.Println("Pods do contain the unexpected annotation key '" + annotationKey + "'")
				return false
			}

			_, existsInStatefulSet := kubegresResource.StatefulSet.Metadata.Annotations[annotationKey]
			if existsInStatefulSet {
				log.Println("StatefulSets do contain the unexpected annotation key '" + annotationKey + "'")
				return false
			}
		}

		log.Println("As expected, Pods and StatefulSets do NOT contain the annotation key '" + annotationKey + "'")
		return true

	}, resourceConfigs.TestTimeout, resourceConfigs.TestRetryInterval).Should(BeTrue())
}

func (r *CustomAnnotationTest) thenPodsStatesShouldBe(nbrePrimary, nbreReplicas int) bool {
	return Eventually(func() bool {

		pods, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres pods")
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

func (r *CustomAnnotationTest) thenDeployedKubegresSpecShouldBeSetTo(specNbreReplicas int32) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	Expect(*r.kubegresResource.Spec.Replicas).Should(Equal(specNbreReplicas))
}
