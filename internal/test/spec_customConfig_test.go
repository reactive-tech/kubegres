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
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"log"
	postgresv1 "reactive-tech.io/kubegres/api/v1"
	"reactive-tech.io/kubegres/internal/controller/ctx"
	"reactive-tech.io/kubegres/internal/controller/states"
	resourceConfigs2 "reactive-tech.io/kubegres/internal/test/resourceConfigs"
	util2 "reactive-tech.io/kubegres/internal/test/util"
	"reactive-tech.io/kubegres/internal/test/util/testcases"
	"time"
)

var _ = Describe("Setting Kubegres specs 'customConfig'", func() {

	var test = SpecCustomConfigTest{}

	BeforeEach(func() {
		Skip("Temporarily skipping test")

		namespace := resourceConfigs2.DefaultNamespace
		test.resourceRetriever = util2.CreateTestResourceRetriever(k8sClientTest, namespace)
		test.resourceCreator = util2.CreateTestResourceCreator(k8sClientTest, test.resourceRetriever, namespace)
		test.dbQueryTestCases = testcases.InitDbQueryTestCases(test.resourceCreator, resourceConfigs2.KubegresResourceName)
		test.resourceCreator.CreateBackUpPvc()
		test.resourceCreator.CreateConfigMapEmpty()
		test.resourceCreator.CreateConfigMapWithAllConfigs()
		test.resourceCreator.CreateConfigMapWithBackupDatabaseScript()
		test.resourceCreator.CreateConfigMapWithPgHbaConf()
		test.resourceCreator.CreateConfigMapWithPostgresConf()
		test.resourceCreator.CreateConfigMapWithPrimaryInitScript()
	})

	AfterEach(func() {
		test.resourceCreator.DeleteAllTestResources(resourceConfigs2.BackUpPvcResourceName)
	})

	Context("GIVEN new Kubegres is created without spec 'customConfig' and spec 'replica' set to 3", func() {

		It("THEN the spec 'customConfig' is set to the value of the constant 'KubegresContext.BaseConfigMapName' AND 1 primary and 2 replicas are deployed", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created without spec 'customConfig''")

			test.givenNewKubegresSpecIsSetTo("", 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 2)

			test.thenPodsShouldNotContainsCustomConfig()

			test.thenPodsContainsConfigTypeAssociatedToFile(ctx.BaseConfigMapVolumeName, states.ConfigMapDataKeyPostgresConf, false)

			test.thenPodsContainsConfigTypeAssociatedToFile(ctx.BaseConfigMapVolumeName, states.ConfigMapDataKeyPrimaryInitScript, true)

			test.thenPodsContainsConfigTypeAssociatedToFile(ctx.BaseConfigMapVolumeName, states.ConfigMapDataKeyPgHbaConf, false)

			test.thenDeployedKubegresSpecShouldBeSetTo(ctx.BaseConfigMapName)

			log.Print("END OF: Test 'GIVEN new Kubegres is created without spec 'customConfig'")
		})
	})

	Context("GIVEN new Kubegres is created with spec 'customConfig' set to a non existent ConfigMap", func() {

		It("THEN an error event should be logged", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'customConfig' set to a non existent ConfigMap'")

			test.givenNewKubegresSpecIsSetTo("doesNotExistConfigMap", 3)

			test.whenKubegresIsCreated()

			test.thenErrorEventShouldBeLogged()

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'customConfig' set to a non existent ConfigMap'")
		})
	})

	Context("GIVEN new Kubegres is created with spec 'customConfig' set to a ConfigMap which is empty", func() {

		It("THEN the base-config should be used for 'postgres.conf', 'primary_init_script.sh' and 'pg_hba.conf''", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'customConfig' set to a ConfigMap which is empty'")

			test.givenNewKubegresSpecIsSetTo(resourceConfigs2.CustomConfigMapEmptyResourceName, 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 2)

			test.thenPodsContainsCustomConfigWithResourceName(resourceConfigs2.CustomConfigMapEmptyResourceName)

			test.thenPodsContainsConfigTypeAssociatedToFile(ctx.BaseConfigMapVolumeName, states.ConfigMapDataKeyPostgresConf, false)

			test.thenPodsContainsConfigTypeAssociatedToFile(ctx.BaseConfigMapVolumeName, states.ConfigMapDataKeyPrimaryInitScript, true)

			test.thenPodsContainsConfigTypeAssociatedToFile(ctx.BaseConfigMapVolumeName, states.ConfigMapDataKeyPgHbaConf, false)

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'customConfig' set to a ConfigMap which is empty'")
		})
	})

	Context("GIVEN new Kubegres is created with spec 'customConfig' set to a ConfigMap containing 'postgres.conf'", func() {

		It("THEN the custom-config should be used for 'postgres.conf' AND the base-config should be used for 'primary_init_script.sh' and 'pg_hba.conf'", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'customConfig' set to a ConfigMap containing 'postgres.conf''")

			test.givenNewKubegresSpecIsSetTo(resourceConfigs2.CustomConfigMapWithPostgresConfResourceName, 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 2)

			test.thenPodsContainsCustomConfigWithResourceName(resourceConfigs2.CustomConfigMapWithPostgresConfResourceName)

			test.thenPodsContainsConfigTypeAssociatedToFile(ctx.CustomConfigMapVolumeName, states.ConfigMapDataKeyPostgresConf, false)

			test.thenPodsContainsConfigTypeAssociatedToFile(ctx.BaseConfigMapVolumeName, states.ConfigMapDataKeyPrimaryInitScript, true)

			test.thenPodsContainsConfigTypeAssociatedToFile(ctx.BaseConfigMapVolumeName, states.ConfigMapDataKeyPgHbaConf, false)

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'customConfig' set to a ConfigMap containing 'postgres.conf'")
		})
	})

	Context("GIVEN new Kubegres is created with spec 'customConfig' set to a ConfigMap containing 'primary_init_script.sh'", func() {

		It("THEN the custom-config should be used for 'primary_init_script.sh' AND the base-config should be used for 'postgres.conf' and 'pg_hba.conf'", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'customConfig' set to a ConfigMap containing 'primary_init_script.sh''")

			test.givenNewKubegresSpecIsSetTo(resourceConfigs2.CustomConfigMapWithPrimaryInitScriptResourceName, 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 2)

			test.thenPodsContainsCustomConfigWithResourceName(resourceConfigs2.CustomConfigMapWithPrimaryInitScriptResourceName)

			test.thenPodsContainsConfigTypeAssociatedToFile(ctx.CustomConfigMapVolumeName, states.ConfigMapDataKeyPrimaryInitScript, true)

			test.thenPodsContainsConfigTypeAssociatedToFile(ctx.BaseConfigMapVolumeName, states.ConfigMapDataKeyPostgresConf, false)

			test.thenPodsContainsConfigTypeAssociatedToFile(ctx.BaseConfigMapVolumeName, states.ConfigMapDataKeyPgHbaConf, false)

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'customConfig' set to a ConfigMap containing 'primary_init_script.sh''")
		})
	})

	Context("GIVEN new Kubegres is created with spec 'customConfig' set to a ConfigMap containing 'pg_hba.conf'", func() {

		It("THEN the custom-config should be used for 'pg_hba.conf' AND the base-config should be used for 'postgres.conf' and 'primary_init_script.sh'", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'customConfig' set to a ConfigMap containing 'pg_hba.conf''")

			test.givenNewKubegresSpecIsSetTo(resourceConfigs2.CustomConfigMapWithPgHbaConfResourceName, 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 2)

			test.thenPodsContainsCustomConfigWithResourceName(resourceConfigs2.CustomConfigMapWithPgHbaConfResourceName)

			test.thenPodsContainsConfigTypeAssociatedToFile(ctx.CustomConfigMapVolumeName, states.ConfigMapDataKeyPgHbaConf, false)

			test.thenPodsContainsConfigTypeAssociatedToFile(ctx.BaseConfigMapVolumeName, states.ConfigMapDataKeyPrimaryInitScript, true)

			test.thenPodsContainsConfigTypeAssociatedToFile(ctx.BaseConfigMapVolumeName, states.ConfigMapDataKeyPostgresConf, false)

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'customConfig' set to a ConfigMap containing 'pg_hba.conf''")
		})
	})

	Context("GIVEN new Kubegres is created with spec 'customConfig' set to base-config AND later it is updated to a configMap containing data-key 'postgres.conf'", func() {

		It("THEN the custom-config should be used for 'postgres.conf'", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'customConfig' set to base-config AND later it is updated to a configMap containing data-key 'postgres.conf''")

			test.givenNewKubegresSpecIsSetTo(ctx.BaseConfigMapName, 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 2)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.givenExistingKubegresSpecIsSetTo(resourceConfigs2.CustomConfigMapWithPostgresConfResourceName)

			test.whenKubernetesIsUpdated()

			test.thenPodsContainsCustomConfigWithResourceName(resourceConfigs2.CustomConfigMapWithPostgresConfResourceName)

			test.thenPodsContainsConfigTypeAssociatedToFile(ctx.CustomConfigMapVolumeName, states.ConfigMapDataKeyPostgresConf, false)

			test.thenPodsContainsConfigTypeAssociatedToFile(ctx.BaseConfigMapVolumeName, states.ConfigMapDataKeyPrimaryInitScript, true)

			test.thenPodsContainsConfigTypeAssociatedToFile(ctx.BaseConfigMapVolumeName, states.ConfigMapDataKeyPgHbaConf, false)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'customConfig' set to base-config AND later it is updated to a configMap containing data-key 'postgres.conf''")
		})
	})

	Context("GIVEN new Kubegres is created with backUp enabled and spec 'customConfig' set to a ConfigMap containing 'backup_database.sh'", func() {

		It("THEN the custom-config should be used for 'backup_database.sh' AND the base-config should be used for 'postgres.conf', 'pg_hba.conf' and 'primary_init_script.sh'", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with backUp enabled and spec 'customConfig' set to a ConfigMap containing 'backup_database.sh''")

			test.givenNewKubegresSpecHasBackupEnabledWithCustomConfig(resourceConfigs2.CustomConfigMapWithBackupDatabaseScriptResourceName, 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 2)

			test.thenCronJobContainsConfigMap(resourceConfigs2.CustomConfigMapWithBackupDatabaseScriptResourceName)

			test.thenPodsContainsCustomConfigWithResourceName(resourceConfigs2.CustomConfigMapWithBackupDatabaseScriptResourceName)

			test.thenPodsContainsConfigTypeAssociatedToFile(ctx.BaseConfigMapVolumeName, states.ConfigMapDataKeyPgHbaConf, false)

			test.thenPodsContainsConfigTypeAssociatedToFile(ctx.BaseConfigMapVolumeName, states.ConfigMapDataKeyPrimaryInitScript, true)

			test.thenPodsContainsConfigTypeAssociatedToFile(ctx.BaseConfigMapVolumeName, states.ConfigMapDataKeyPostgresConf, false)

			log.Print("END OF: Test 'GIVEN new Kubegres is created with backUp enabled and spec 'customConfig' set to a ConfigMap containing 'backup_database.sh''")
		})
	})

	Context("GIVEN new Kubegres is created with backUp enabled and spec 'customConfig' set to base-config AND later it is updated to a configMap containing data-key 'backup_database.sh'", func() {

		It("THEN the custom-config should be used for 'postgres.conf' AND the base-config should be used for 'postgres.conf', 'pg_hba.conf' and 'primary_init_script.sh'", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with backUp enabled and spec 'customConfig' set to base-config AND later it is updated to a configMap containing data-key 'backup_database.sh''")

			test.givenNewKubegresSpecHasBackupEnabledWithCustomConfig(ctx.BaseConfigMapName, 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 2)

			test.thenCronJobContainsConfigMap(ctx.BaseConfigMapName)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			test.givenExistingKubegresSpecIsSetTo(resourceConfigs2.CustomConfigMapWithBackupDatabaseScriptResourceName)

			test.whenKubernetesIsUpdated()

			test.thenPodsContainsCustomConfigWithResourceName(resourceConfigs2.CustomConfigMapWithBackupDatabaseScriptResourceName)

			test.thenCronJobContainsConfigMap(resourceConfigs2.CustomConfigMapWithBackupDatabaseScriptResourceName)

			test.thenPodsContainsConfigTypeAssociatedToFile(ctx.BaseConfigMapVolumeName, states.ConfigMapDataKeyPgHbaConf, false)

			test.thenPodsContainsConfigTypeAssociatedToFile(ctx.BaseConfigMapVolumeName, states.ConfigMapDataKeyPrimaryInitScript, true)

			test.thenPodsContainsConfigTypeAssociatedToFile(ctx.BaseConfigMapVolumeName, states.ConfigMapDataKeyPostgresConf, false)

			test.dbQueryTestCases.ThenWeCanSqlQueryPrimaryDb()
			test.dbQueryTestCases.ThenWeCanSqlQueryReplicaDb()

			log.Print("END OF: Test 'GIVEN new Kubegres is created with backUp enabled and spec 'customConfig' set to base-config AND later it is updated to a configMap containing data-key 'backup_database.sh''")
		})
	})

})

type SpecCustomConfigTest struct {
	kubegresResource  *postgresv1.Kubegres
	dbQueryTestCases  testcases.DbQueryTestCases
	resourceCreator   util2.TestResourceCreator
	resourceRetriever util2.TestResourceRetriever
}

func (r *SpecCustomConfigTest) givenNewKubegresSpecIsSetTo(customConfig string, specNbreReplicas int32) {
	r.kubegresResource = resourceConfigs2.LoadKubegresYaml()
	r.kubegresResource.Spec.CustomConfig = customConfig
	r.kubegresResource.Spec.Replicas = &specNbreReplicas
}

func (r *SpecCustomConfigTest) givenExistingKubegresSpecIsSetTo(customConfig string) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	r.kubegresResource.Spec.CustomConfig = customConfig
}

func (r *SpecCustomConfigTest) givenNewKubegresSpecHasBackupEnabledWithCustomConfig(customConfig string, specNbreReplicas int32) {
	r.kubegresResource = resourceConfigs2.LoadKubegresYaml()
	r.kubegresResource.Spec.CustomConfig = customConfig
	r.kubegresResource.Spec.Replicas = &specNbreReplicas
	r.kubegresResource.Spec.Backup.Schedule = "*/1 * * * *"
	r.kubegresResource.Spec.Backup.PvcName = resourceConfigs2.BackUpPvcResourceName
	r.kubegresResource.Spec.Backup.VolumeMount = "/tmp/my-kubegres"
}

func (r *SpecCustomConfigTest) whenKubegresIsCreated() {
	if r.kubegresResource == nil {
		r.kubegresResource = resourceConfigs2.LoadKubegresYaml()
	}
	r.resourceCreator.CreateKubegres(r.kubegresResource)
}

func (r *SpecCustomConfigTest) whenKubernetesIsUpdated() {
	r.resourceCreator.UpdateResource(r.kubegresResource, "Kubegres")
}

func (r *SpecCustomConfigTest) thenErrorEventShouldBeLogged() {
	expectedErrorEvent := util2.EventRecord{
		Eventtype: v12.EventTypeWarning,
		Reason:    "SpecCheckErr",
		Message:   "In the Resources Spec the value of 'spec.customConfig' has a configMap name which is not deployed. Please deploy this configMap otherwise this operator cannot work correctly.",
	}
	Eventually(func() bool {
		_, err := r.resourceRetriever.GetKubegres()
		if err != nil {
			return false
		}
		return eventRecorderTest.CheckEventExist(expectedErrorEvent)

	}, resourceConfigs2.TestTimeout, resourceConfigs2.TestRetryInterval).Should(BeTrue())
}

func (r *SpecCustomConfigTest) thenPodsStatesShouldBe(nbrePrimary, nbreReplicas int) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
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

func (r *SpecCustomConfigTest) thenPodsShouldNotContainsCustomConfig() bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		for _, resource := range kubegresResources.Resources {
			for _, volume := range resource.StatefulSet.Spec.Template.Spec.Volumes {
				if volume.Name == ctx.CustomConfigMapVolumeName {
					log.Println("Pod '" + resource.Pod.Name + "' has the customConfig type: '" + ctx.CustomConfigMapVolumeName + "'. Waiting...")
					return false
				}
			}

		}

		return true

	}, time.Second*10, time.Second*5).Should(BeTrue())
}

func (r *SpecCustomConfigTest) thenPodsContainsCustomConfigWithResourceName(expectedCustomConfigResourceName string) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		for _, resource := range kubegresResources.Resources {
			if !r.hasCustomConfigWithResourceName(resource.StatefulSet.Spec, expectedCustomConfigResourceName) {
				log.Println("Pod '" + resource.Pod.Name + "' has NOT the expected customConfig: '" + expectedCustomConfigResourceName + "'. Waiting...")
				return false
			}
		}

		return true

	}, resourceConfigs2.TestTimeout, resourceConfigs2.TestRetryInterval).Should(BeTrue())

}

func (r *SpecCustomConfigTest) hasCustomConfigWithResourceName(statefulSetSpec v1.StatefulSetSpec, expectedCustomConfigResourceName string) bool {
	for _, volume := range statefulSetSpec.Template.Spec.Volumes {
		if volume.Name == ctx.CustomConfigMapVolumeName && volume.ConfigMap.Name == expectedCustomConfigResourceName {
			return true
		}
	}
	return false
}

func (r *SpecCustomConfigTest) thenPodsContainsConfigTypeAssociatedToFile(expectedVolumeNameForConfigType, expectedConfigFile string, isOnlyPrimaryStatefulSet bool) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		for _, resource := range kubegresResources.Resources {

			if isOnlyPrimaryStatefulSet && !resource.IsPrimary {
				continue
			}

			if !r.hasConfigTypeAssociatedToFile(resource.StatefulSet.Spec, expectedVolumeNameForConfigType, expectedConfigFile) {
				log.Println("Pod '" + resource.Pod.Name + "' doesn't have the expected config type: '" + expectedVolumeNameForConfigType + "' and file: '" + expectedConfigFile + "'. Waiting...")
				return false
			}
		}

		return true

	}, time.Second*10, time.Second*5).Should(BeTrue())

}

func (r *SpecCustomConfigTest) thenCronJobContainsConfigMap(expectedConfigMapName string) bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		if err != nil && !apierrors.IsNotFound(err) {
			log.Println("ERROR while retrieving Kubegres kubegresResources")
			return false
		}

		backUpCronJob := kubegresResources.BackUpCronJob
		if backUpCronJob.Name == "" {
			return false
		}

		cronJobConfigMapName := backUpCronJob.Spec.JobTemplate.Spec.Template.Spec.Volumes[1].ConfigMap.Name

		if expectedConfigMapName != cronJobConfigMapName {
			log.Println("CronJob '" + backUpCronJob.Name + "' doesn't have the expected configMap name: '" + expectedConfigMapName + "'. Waiting...")
			return false
		}

		return true

	}, time.Second*10, time.Second*5).Should(BeTrue())

}

func (r *SpecCustomConfigTest) hasConfigTypeAssociatedToFile(statefulSetSpec v1.StatefulSetSpec, expectedVolumeNameForConfigType, expectedConfigFile string) bool {
	for _, volumeMount := range statefulSetSpec.Template.Spec.Containers[0].VolumeMounts {
		if volumeMount.Name == expectedVolumeNameForConfigType && volumeMount.SubPath == expectedConfigFile {
			return true
		}
	}
	return false
}

func (r *SpecCustomConfigTest) thenDeployedKubegresSpecShouldBeSetTo(customConfig string) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}

	Expect(r.kubegresResource.Spec.CustomConfig).Should(Equal(customConfig))
}
