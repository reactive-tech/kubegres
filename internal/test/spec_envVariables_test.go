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
	"log"
	postgresv1 "reactive-tech.io/kubegres/api/v1"
	"reactive-tech.io/kubegres/internal/controller/ctx"
	resourceConfigs2 "reactive-tech.io/kubegres/internal/test/resourceConfigs"
	util2 "reactive-tech.io/kubegres/internal/test/util"
	"reactive-tech.io/kubegres/internal/test/util/testcases"
)

var _ = Describe("Setting Kubegres spec 'env.*'", func() {

	var test = SpecEnVariablesTest{}

	BeforeEach(func() {
		//Skip("Temporarily skipping test")

		namespace := resourceConfigs2.DefaultNamespace
		test.resourceRetriever = util2.CreateTestResourceRetriever(k8sClientTest, namespace)
		test.resourceCreator = util2.CreateTestResourceCreator(k8sClientTest, test.resourceRetriever, namespace)
		test.dbQueryTestCases = testcases.InitDbQueryTestCases(test.resourceCreator, resourceConfigs2.KubegresResourceName)
	})

	AfterEach(func() {
		test.resourceCreator.DeleteAllTestResources()
	})

	Context("GIVEN new Kubegres is created without environment variable of postgres super-user password", func() {

		It("THEN an error event should be logged saying the super-user password is missing", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created without environment variable of postgres super-user password'")

			test.givenNewKubegresWithoutEnvVarOfPostgresSuperUserPassword()

			test.whenKubegresIsCreated()

			test.thenErrorEventShouldBeLogged("spec.env.POSTGRES_PASSWORD")

			log.Print("END OF: Test 'GIVEN new Kubegres is created without environment variable of postgres super-user password'")
		})
	})

	Context("GIVEN new Kubegres is created without environment variable of postgres replication-user password", func() {

		It("THEN an error event should be logged saying the replication-user password is missing", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created without environment variable of postgres replication-user password'")

			test.givenNewKubegresWithoutEnvVarOfPostgresReplicationUserPassword()

			test.whenKubegresIsCreated()

			test.thenErrorEventShouldBeLogged("spec.env.POSTGRES_REPLICATION_PASSWORD")

			log.Print("END OF: Test 'GIVEN new Kubegres is created without environment variable of postgres replication-user password'")
		})
	})

	Context("GIVEN new Kubegres is created with all environment variables AND spec 'replica' set to 3", func() {

		It("THEN 1 primary and 2 replica should be created with all environment variables", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with all environment variables'")

			test.givenNewKubegresWithAllEnvVarsSet(3)

			test.whenKubegresIsCreated()

			test.thenPodsShouldContainAllEnvVariables(1, 2)

			test.thenDeployedKubegresSpecShouldHaveAllEnvVars()

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN new Kubegres is created with all environment variables'")
		})
	})

	Context("GIVEN new Kubegres is created with all environment variables and a custom one AND spec 'replica' set to 3", func() {

		It("THEN 1 primary and 2 replica should be created with all environment variables including the custom one too", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with all environment variables and a custom one'")

			test.givenNewKubegresWithAllEnvVarsSetAndACustomOne(3)

			test.whenKubegresIsCreated()

			test.thenPodsShouldContainAllEnvVariables(1, 2)

			test.thenDeployedKubegresSpecShouldHaveAllEnvVars()

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN new Kubegres is created with all environment variables and a custom one'")
		})
	})
})

type SpecEnVariablesTest struct {
	kubegresResource      *postgresv1.Kubegres
	dbQueryTestCases      testcases.DbQueryTestCases
	resourceCreator       util2.TestResourceCreator
	resourceRetriever     util2.TestResourceRetriever
	resourceModifier      util2.TestResourceModifier
	customEnvVariableName string
	customEnvVariableKey  string
}

func (r *SpecEnVariablesTest) givenNewKubegresWithoutEnvVarOfPostgresSuperUserPassword() {
	r.kubegresResource = resourceConfigs2.LoadKubegresYaml()
	r.kubegresResource.Spec.Env = []v12.EnvVar{}
	r.resourceModifier.AppendEnvVarFromSecretKey(ctx.EnvVarNameOfPostgresReplicationUserPsw, "replicationUserPassword", r.kubegresResource)
}

func (r *SpecEnVariablesTest) givenNewKubegresWithoutEnvVarOfPostgresReplicationUserPassword() {
	r.kubegresResource = resourceConfigs2.LoadKubegresYaml()
	r.kubegresResource.Spec.Env = []v12.EnvVar{}
	r.resourceModifier.AppendEnvVarFromSecretKey(ctx.EnvVarNameOfPostgresSuperUserPsw, "superUserPassword", r.kubegresResource)
}

func (r *SpecEnVariablesTest) givenNewKubegresWithAllEnvVarsSet(specNbreReplicas int32) {

	r.kubegresResource = resourceConfigs2.LoadKubegresYaml()
	r.kubegresResource.Spec.Replicas = &specNbreReplicas

	r.kubegresResource.Spec.Env = []v12.EnvVar{}
	r.resourceModifier.AppendEnvVarFromSecretKey(ctx.EnvVarNameOfPostgresReplicationUserPsw, "replicationUserPassword", r.kubegresResource)
	r.resourceModifier.AppendEnvVarFromSecretKey(ctx.EnvVarNameOfPostgresSuperUserPsw, "superUserPassword", r.kubegresResource)
}

func (r *SpecEnVariablesTest) givenNewKubegresWithAllEnvVarsSetAndACustomOne(specNbreReplicas int32) {

	r.givenNewKubegresWithAllEnvVarsSet(specNbreReplicas)

	r.customEnvVariableName = "POSTGRES_CUSTOM_ENV_VAR"
	r.customEnvVariableKey = "myAppUserPassword"
	r.resourceModifier.AppendEnvVarFromSecretKey(r.customEnvVariableName, r.customEnvVariableKey, r.kubegresResource)
}

func (r *SpecEnVariablesTest) whenKubegresIsCreated() {
	r.resourceCreator.CreateKubegres(r.kubegresResource)
}

func (r *SpecEnVariablesTest) thenPodsShouldContainAllEnvVariables(nbrePrimary, nbreReplicas int) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		for _, resource := range kubegresResources.Resources {
			envVars := resource.Pod.Spec.Containers[0].Env

			if !r.doesEnvVarExist(ctx.EnvVarNameOfPostgresSuperUserPsw, envVars) {
				log.Println("Pod '" + resource.Pod.Name + "' doesn't have the expected env variable: '" + ctx.EnvVarNameOfPostgresSuperUserPsw + "'. Waiting...")
				return false
			}

			if !r.doesEnvVarExist(ctx.EnvVarNameOfPostgresReplicationUserPsw, envVars) {
				log.Println("Pod '" + resource.Pod.Name + "' doesn't have the expected env variable: '" + ctx.EnvVarNameOfPostgresReplicationUserPsw + "'. Waiting...")
				return false
			}

			if r.customEnvVariableName != "" {
				if !r.doesEnvVarExist(r.customEnvVariableName, envVars) {
					log.Println("Pod '" + resource.Pod.Name + "' doesn't have the expected env variable: '" + r.customEnvVariableName + "'. Waiting...")
					return false
				}
			}
		}

		return kubegresResources.AreAllReady &&
			kubegresResources.NbreDeployedPrimary == nbrePrimary &&
			kubegresResources.NbreDeployedReplicas == nbreReplicas

	}, resourceConfigs2.TestTimeout, resourceConfigs2.TestRetryInterval).Should(BeTrue())
}

func (r *SpecEnVariablesTest) thenDeployedKubegresSpecShouldHaveAllEnvVars() {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	Expect(r.doesEnvVarExist(ctx.EnvVarNameOfPostgresSuperUserPsw, r.kubegresResource.Spec.Env)).Should(Equal(true))
	Expect(r.doesEnvVarExist(ctx.EnvVarNameOfPostgresReplicationUserPsw, r.kubegresResource.Spec.Env)).Should(Equal(true))
}

func (r *SpecEnVariablesTest) doesEnvVarExist(envVarName string, envVars []v12.EnvVar) bool {
	for _, env := range envVars {
		if env.Name == envVarName {
			return true
		}
	}
	return false
}

func (r *SpecEnVariablesTest) thenErrorEventShouldBeLogged(specName string) {
	expectedErrorEvent := util2.EventRecord{
		Eventtype: v12.EventTypeWarning,
		Reason:    "SpecCheckErr",
		Message:   "In the Resources Spec the value of '" + specName + "' is undefined. Please set a value otherwise this operator cannot work correctly.",
	}
	Eventually(func() bool {
		_, err := r.resourceRetriever.GetKubegres()
		if err != nil {
			return false
		}
		return eventRecorderTest.CheckEventExist(expectedErrorEvent)

	}, resourceConfigs2.TestTimeout, resourceConfigs2.TestRetryInterval).Should(BeTrue())
}
