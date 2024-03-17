package test

import (
	"log"
	"reflect"
	"time"

	resourceConfigs2 "reactive-tech.io/kubegres/internal/test/resourceConfigs"
	util2 "reactive-tech.io/kubegres/internal/test/util"
	"reactive-tech.io/kubegres/internal/test/util/testcases"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v12 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/pointer"

	v1 "reactive-tech.io/kubegres/api/v1"
)

var _ = Describe("Setting Kubegres spec 'containerSecurityContext'", func() {
	var test = SpeccontainerSecurityContextTest{}
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

	Context("GIVEN new Kubegres is created without spec 'containerSecurityContext' and with spec 'replica' set to 3", func() {

		It("THEN 1 primary and 2 replica should be created with 'containerSecurityContext' set to nil", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created without spec 'containerSecurityContext' and with spec 'replica' set to 3'")

			test.givenNewKubegresSpecIsWithoutcontainerSecurityContext(3)

			test.whenKubegresIsCreated()

			test.thenStatefulSetStatesShouldBeWithoutcontainerSecurityContext(1, 2)

			test.thenDeployedKubegresSpecShouldWithoutcontainerSecurityContext()

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN new Kubegres is created without spec 'containerSecurityContext' and with spec 'replica' set to 3'")
		})
	})

	Context("GIVEN new Kubegres is created with spec 'containerSecurityContext' set to a value and spec 'replica' set to 3 and later 'containerSecurityContext' is updated to a new value", func() {

		It("GIVEN new Kubegres is created with spec 'containerSecurityContext' set to a value and spec 'replica' set to 3 THEN 1 primary and 2 replica should be created with spec 'containerSecurityContext' set the value", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'containerSecurityContext' set to a value and spec 'replica' set to 3")

			securityContext := test.givencontainerSecurityContext1()

			test.givenNewKubegresSpecIsSetTo(securityContext, 3)

			test.whenKubegresIsCreated()

			test.thenStatefulSetStatesShouldBe(securityContext, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(securityContext)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.keepCreatedResourcesForNextTest = true

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'containerSecurityContext' set to a value and spec 'replica' set to 3'")
		})

		It("GIVEN existing Kubegres is updated with spec 'containerSecurityContext' set to a new value THEN 1 primary and 2 replica should be re-deployed with spec 'containerSecurityContext' set the new value", func() {

			log.Print("START OF: Test 'GIVEN existing Kubegres is updated with spec 'containerSecurityContext' set to a new value")

			newsecurityContext := test.givencontainerSecurityContext2()

			test.givenExistingKubegresSpecIsSetTo(newsecurityContext)

			test.whenKubernetesIsUpdated()

			test.thenStatefulSetStatesShouldBe(newsecurityContext, 1, 2)

			test.thenDeployedKubegresSpecShouldBeSetTo(newsecurityContext)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN existing Kubegres is updated with spec 'containerSecurityContext' set to a new value")
		})
	})

})

type SpeccontainerSecurityContextTest struct {
	keepCreatedResourcesForNextTest bool
	kubegresResource                *v1.Kubegres
	dbQueryTestCases                testcases.DbQueryTestCases
	resourceCreator                 util2.TestResourceCreator
	resourceRetriever               util2.TestResourceRetriever
}

func (r *SpeccontainerSecurityContextTest) givencontainerSecurityContext1() *v12.SecurityContext {
	return &v12.SecurityContext{
		RunAsNonRoot:             pointer.BoolPtr(true),
		AllowPrivilegeEscalation: pointer.BoolPtr(false),
		Capabilities:             &v12.Capabilities{Drop: []v12.Capability{"ALL"}},
		SeccompProfile:           &v12.SeccompProfile{Type: v12.SeccompProfileTypeRuntimeDefault},
		ReadOnlyRootFilesystem:   pointer.BoolPtr(true),
		Privileged:               pointer.BoolPtr(false),
	}
}

func (r *SpeccontainerSecurityContextTest) givencontainerSecurityContext2() *v12.SecurityContext {
	return &v12.SecurityContext{
		RunAsNonRoot:             pointer.BoolPtr(true),
		AllowPrivilegeEscalation: pointer.BoolPtr(true),
		Capabilities:             &v12.Capabilities{Add: []v12.Capability{"ALL"}}, // Inverse of drop
		ReadOnlyRootFilesystem:   pointer.BoolPtr(false),
		Privileged:               pointer.BoolPtr(true),
	}
}

func (r *SpeccontainerSecurityContextTest) givenNewKubegresSpecIsWithoutcontainerSecurityContext(replicaCount int32) {
	r.kubegresResource = resourceConfigs2.LoadKubegresYaml()
	r.kubegresResource.Spec.ContainerSecurityContext = &v12.SecurityContext{}
	r.kubegresResource.Spec.Replicas = &replicaCount
}

func (r *SpeccontainerSecurityContextTest) givenNewKubegresSpecIsSetTo(securityContext *v12.SecurityContext, specNbreReplicas int32) {
	r.kubegresResource = resourceConfigs2.LoadKubegresYaml()
	r.kubegresResource.Spec.ContainerSecurityContext = securityContext
	r.kubegresResource.Spec.Replicas = &specNbreReplicas
}

func (r *SpeccontainerSecurityContextTest) givenExistingKubegresSpecIsSetTo(containerSecurityContext *v12.SecurityContext) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	r.kubegresResource.Spec.ContainerSecurityContext = containerSecurityContext
}

func (r *SpeccontainerSecurityContextTest) whenKubegresIsCreated() {
	r.resourceCreator.CreateKubegres(r.kubegresResource)
}

func (r *SpeccontainerSecurityContextTest) whenKubernetesIsUpdated() {
	r.resourceCreator.UpdateResource(r.kubegresResource, "Kubegres")
}

func (r *SpeccontainerSecurityContextTest) thenStatefulSetStatesShouldBeWithoutcontainerSecurityContext(nbrePrimary, nbreReplicas int) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		for _, resource := range kubegresResources.Resources {
			currentContainerSecurityContext := resource.StatefulSet.Spec.Template.Spec.Containers[0].SecurityContext
			emptyResources := &v12.SecurityContext{}
			if !reflect.DeepEqual(currentContainerSecurityContext, emptyResources) {
				log.Println("StatefulSet '" + resource.StatefulSet.Name + "' has not the expected containerSecurityContext which should be nil" +
					"Current valye: '" + currentContainerSecurityContext.String() + "'. Waiting...")
				return false
			}

			// Check whether InitContainer has the same security context
			if len(resource.StatefulSet.Spec.Template.Spec.InitContainers) > 0 {
				currentInitContainerSecurityContext := resource.StatefulSet.Spec.Template.Spec.InitContainers[0].SecurityContext
				if !reflect.DeepEqual(currentInitContainerSecurityContext, emptyResources) {
					log.Println("StatefulSet '" + resource.StatefulSet.Name + "' has not the expected initContainerSecurityContext which should be nil" +
						"Current valye: '" + currentInitContainerSecurityContext.String() + "'. Waiting...")
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

func (r *SpeccontainerSecurityContextTest) thenStatefulSetStatesShouldBe(expectedContainerSecurityContext *v12.SecurityContext, nbrePrimary, nbreReplicas int) bool {
	return Eventually(func() bool {
		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		for _, resource := range kubegresResources.Resources {
			currentContainerSecurityContext := resource.StatefulSet.Spec.Template.Spec.Containers[0].SecurityContext

			if !reflect.DeepEqual(currentContainerSecurityContext, expectedContainerSecurityContext) {
				{
					log.Println("StatefulSet '" + resource.StatefulSet.Name + "' doesn't have the expected spec securityContext: " + expectedContainerSecurityContext.String() + " " +
						"Current value: '" + currentContainerSecurityContext.String() + "'. Waiting...")
					return false
				}
			}

			// Check whether InitContainer has the same security context
			if len(resource.StatefulSet.Spec.Template.Spec.InitContainers) > 0 {
				currentInitContainerSecurityContext := resource.StatefulSet.Spec.Template.Spec.InitContainers[0].SecurityContext
				if !reflect.DeepEqual(currentInitContainerSecurityContext, expectedContainerSecurityContext) {
					log.Println("StatefulSet '" + resource.StatefulSet.Name + "' has not the expected initContainerSecurityContext which should be nil" +
						"Current valye: '" + currentInitContainerSecurityContext.String() + "'. Waiting...")
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

func (r *SpeccontainerSecurityContextTest) thenDeployedKubegresSpecShouldWithoutcontainerSecurityContext() {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()
	emptyResources := &v12.SecurityContext{}
	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	Expect(r.kubegresResource.Spec.SecurityContext).Should(Equal(emptyResources))
}

func (r *SpeccontainerSecurityContextTest) thenDeployedKubegresSpecShouldBeSetTo(expectedcontainerSecurityContext *v12.SecurityContext) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	Expect(r.kubegresResource.Spec.ContainerSecurityContext).Should(Equal(expectedcontainerSecurityContext))
}
