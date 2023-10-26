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
	"time"
)

const customAnnotationKey = "linkerd.io/inject"
const customAnnotationValue = "enabled"
const customEnvVarName = "MY_CUSTOM_ENV_VAR"
const customEnvVarValue = "postgreSqlPower"
const scheduleBackupEveryMin = "*/1 * * * *"
const scheduleBackupEvery2Mins = "*/2 * * * *"

var _ = Describe("Setting Kubegres specs 'backup.*'", func() {

	var test = SpecBackUpTest{}

	BeforeEach(func() {
		//Skip("Temporarily skipping test")

		namespace := resourceConfigs2.DefaultNamespace
		test.resourceRetriever = util2.CreateTestResourceRetriever(k8sClientTest, namespace)
		test.resourceCreator = util2.CreateTestResourceCreator(k8sClientTest, test.resourceRetriever, namespace)
		test.dbQueryTestCases = testcases.InitDbQueryTestCases(test.resourceCreator, resourceConfigs2.KubegresResourceName)
		test.resourceCreator.CreateConfigMapWithBackupDatabaseScript()
		test.resourceCreator.CreateConfigMapWithPgHbaConf()
		test.resourceCreator.CreateBackUpPvc()
		test.resourceCreator.CreateBackUpPvc2()
	})

	AfterEach(func() {
		if !test.keepCreatedResourcesForNextTest {
			test.resourceCreator.DeleteAllTestResources(resourceConfigs2.BackUpPvcResourceName, resourceConfigs2.BackUpPvcResourceName2)
		} else {
			test.keepCreatedResourcesForNextTest = false
		}
	})

	Context("GIVEN new Kubegres is created without spec 'backup' AND with spec 'replica' set to 3", func() {

		It("THEN backup CronJob is NOT created AND 1 primary and 2 replicas are deployed", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created without spec 'backup''")

			test.givenNewKubegresSpecIsSetTo(ctx.BaseConfigMapName, "", "", "", 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 2)

			test.thenBackupCronJobDoesNOTExist()

			log.Print("END OF: Test 'GIVEN new Kubegres is created without spec 'customConfig'")
		})
	})

	Context("GIVEN new Kubegres is created without spec 'backup.schedule' BUT with spec 'backup.volumeMount' AND with spec 'backup.pvcName' AND with spec 'replica' set to 3", func() {

		It("THEN backup CronJob is NOT created AND 1 primary and 2 replicas are deployed", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created without spec 'backup.schedule' BUT with spec 'backup.volumeMount' AND with spec 'backup.pvcName'")

			test.givenNewKubegresSpecIsSetTo(ctx.BaseConfigMapName, "", resourceConfigs2.BackUpPvcResourceName, "/tmp/my-kubegres", 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 2)

			test.thenBackupCronJobDoesNOTExist()

			log.Print("END OF: Test 'GIVEN new Kubegres is created without spec 'backup.schedule' BUT with spec 'backup.volumeMount' AND with spec 'backup.pvcName'")
		})
	})

	Context("GIVEN new Kubegres is created with spec 'backup.schedule' BUT WITHOUT spec 'backup.volumeMount''", func() {

		It("THEN an error event should be logged", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'backup.schedule' BUT WITHOUT spec 'backup.volumeMount''")

			test.givenNewKubegresSpecIsSetTo(ctx.BaseConfigMapName, scheduleBackupEveryMin, resourceConfigs2.BackUpPvcResourceName, "", 3)

			test.whenKubegresIsCreated()

			test.thenErrorEventShouldBeLogged("spec.Backup.VolumeMount")

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'backup.schedule' BUT WITHOUT spec 'backup.volumeMount''")
		})
	})

	Context("GIVEN new Kubegres is created with spec 'backup.schedule' AND 'backup.volumeMount' BUT WITHOUT spec 'backup.pvcName''", func() {

		It("THEN an error event should be logged", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'backup.schedule' AND 'backup.volumeMount' BUT WITHOUT spec 'backup.pvcName''")

			test.givenNewKubegresSpecIsSetTo(ctx.BaseConfigMapName, scheduleBackupEveryMin, "", "/tmp/my-kubegres", 3)

			test.whenKubegresIsCreated()

			test.thenErrorEventShouldBeLogged("spec.Backup.PvcName")

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'backup.schedule' AND 'backup.volumeMount' BUT WITHOUT spec 'backup.pvcName''")
		})
	})

	Context("GIVEN new Kubegres is created with spec 'backup.schedule' AND 'backup.volumeMount' AND 'backup.pvcName' BUT the given PVC is NOT deployed'", func() {

		It("THEN an error event should be logged saying PVC is NOT deployed", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'backup.schedule' AND 'backup.volumeMount' AND 'backup.pvcName' BUT the given PVC is NOT deployed'")

			test.givenNewKubegresSpecIsSetTo(ctx.BaseConfigMapName, scheduleBackupEveryMin, "PvcDoesNotExists", "/tmp/my-kubegres", 3)

			test.whenKubegresIsCreated()

			test.thenErrorEventSayingPvcIsNotDeployed()

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'backup.schedule' AND 'backup.volumeMount' AND 'backup.pvcName' BUT the given PVC is NOT deployed'")
		})
	})

	Context("GIVEN new Kubegres is created with spec 'backup.schedule' AND 'backup.volumeMount' AND 'backup.pvcName' and the given PVC is deployed", func() {

		It("THEN backup CronJob is created AND 1 primary and 2 replicas are deployed", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with spec 'backup.schedule' AND 'backup.volumeMount' AND 'backup.pvcName' and the given PVC is deployed")

			test.givenNewKubegresSpecIsSetTo(ctx.BaseConfigMapName, scheduleBackupEveryMin, resourceConfigs2.BackUpPvcResourceName, "/tmp/my-kubegres", 3)
			test.givenKubegresEnvVarIsSetTo(customEnvVarName, customEnvVarValue)
			test.givenKubegresAnnotationIsSetTo(customAnnotationKey, customAnnotationValue)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 2)

			test.thenCronJobExistsWithSpec(ctx.BaseConfigMapName, scheduleBackupEveryMin, resourceConfigs2.BackUpPvcResourceName, "/tmp/my-kubegres")
			test.thenCronJobSpecShouldHaveEnvVar(customEnvVarName, customEnvVarValue)
			test.thenCronJobSpecShouldHaveAnnotation(customAnnotationKey, customAnnotationValue)

			log.Print("END OF: Test 'GIVEN new Kubegres is created with spec 'backup.schedule' AND 'backup.volumeMount' AND 'backup.pvcName' and the given PVC is deployed")
		})
	})

	Context("GIVEN new Kubegres is created with backup specs set AND later Kubegres is updated with new values for backup specs", func() {

		It("THEN backup CronJob is updated with the new backup specs", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with backup specs set AND later Kubegres is updated with new values for backup specs'")

			test.givenNewKubegresSpecIsSetTo(ctx.BaseConfigMapName, scheduleBackupEveryMin, resourceConfigs2.BackUpPvcResourceName, "/tmp/my-kubegres", 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 2)

			test.givenExistingKubegresSpecIsSetTo(ctx.BaseConfigMapName, scheduleBackupEvery2Mins, resourceConfigs2.BackUpPvcResourceName2, "/tmp/my-kubegres-2")

			test.whenKubernetesIsUpdated()

			test.thenCronJobExistsWithSpec(ctx.BaseConfigMapName, scheduleBackupEvery2Mins, resourceConfigs2.BackUpPvcResourceName2, "/tmp/my-kubegres-2")

			log.Print("END OF: Test 'GIVEN new Kubegres is created with backup specs set AND later Kubegres is updated with new values for backup specs'")
		})
	})

	Context("GIVEN new Kubegres is created with backup specs set AND later the Kubernetes field 'spec.customConfig' is changed", func() {

		It("the Kubernetes field 'spec.customConfig' is changed to a configMap which does NOT contain 'backup_database.sh' "+
			"THEN backup CronJob should NOT be updated and should remain with configMap set to the base configMap", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with backup specs AND later the Kubernetes field " +
				"'spec.customConfig' is changed to a configMap which does NOT contain 'backup_database.sh'")

			test.givenNewKubegresSpecIsSetTo(ctx.BaseConfigMapName, scheduleBackupEveryMin, resourceConfigs2.BackUpPvcResourceName, "/tmp/my-kubegres", 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 2)

			test.givenExistingKubegresSpecIsSetTo(resourceConfigs2.CustomConfigMapWithPgHbaConfResourceName, scheduleBackupEveryMin, resourceConfigs2.BackUpPvcResourceName, "/tmp/my-kubegres")

			test.whenKubernetesIsUpdated()

			test.thenCronJobExistsWithSpec(ctx.BaseConfigMapName, scheduleBackupEveryMin, resourceConfigs2.BackUpPvcResourceName, "/tmp/my-kubegres")

			test.keepCreatedResourcesForNextTest = true

			log.Print("END OF: Test 'GIVEN new Kubegres is created with backup specs AND later the Kubernetes field " +
				"'spec.customConfig' is changed to a configMap which does NOT contain 'backup_database.sh'")
		})

		It("the Kubernetes field 'spec.customConfig' is changed to a configMap which contain 'backup_database.sh' "+
			"THEN backup CronJob should be updated with the configMap name set to the new custom configMap", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with backup specs AND later the Kubernetes field " +
				"'spec.customConfig' is changed to a configMap which contain 'backup_database.sh'")

			test.givenExistingKubegresSpecIsSetTo(resourceConfigs2.CustomConfigMapWithBackupDatabaseScriptResourceName, scheduleBackupEveryMin, resourceConfigs2.BackUpPvcResourceName, "/tmp/my-kubegres")

			test.whenKubernetesIsUpdated()

			test.thenCronJobExistsWithSpec(resourceConfigs2.CustomConfigMapWithBackupDatabaseScriptResourceName, scheduleBackupEveryMin, resourceConfigs2.BackUpPvcResourceName, "/tmp/my-kubegres")

			log.Print("END OF: Test 'GIVEN new Kubegres is created with backup specs AND later the Kubernetes field " +
				"'spec.customConfig' is changed to a configMap which contain 'backup_database.sh'")
		})
	})

	Context("GIVEN new Kubegres is created with backup specs set AND later Kubegres is updated with backup disabled", func() {

		It("THEN existing backup CronJob is deleted", func() {

			log.Print("START OF: Test 'GIVEN new Kubegres is created with backup specs set AND later Kubegres is updated with backup disabled'")

			test.givenNewKubegresSpecIsSetTo(ctx.BaseConfigMapName, scheduleBackupEveryMin, resourceConfigs2.BackUpPvcResourceName, "/tmp/my-kubegres", 3)

			test.whenKubegresIsCreated()

			test.thenPodsStatesShouldBe(1, 2)

			test.givenExistingKubegresSpecIsSetTo(ctx.BaseConfigMapName, "", "", "")

			test.whenKubernetesIsUpdated()

			test.thenPodsStatesShouldBe(1, 2)

			test.thenBackupCronJobDoesNOTExist()

			log.Print("END OF: Test 'GIVEN new Kubegres is created with backup specs set AND later Kubegres is updated with backup disabled'")
		})
	})

})

type SpecBackUpTest struct {
	keepCreatedResourcesForNextTest bool
	kubegresResource                *postgresv1.Kubegres
	dbQueryTestCases                testcases.DbQueryTestCases
	resourceCreator                 util2.TestResourceCreator
	resourceRetriever               util2.TestResourceRetriever
	resourceModifier                util2.TestResourceModifier
}

func (r *SpecBackUpTest) givenNewKubegresSpecIsSetTo(customConfig, backupSchedule, backupPvcName, backupVolumeMount string, specNbreReplicas int32) {
	r.kubegresResource = resourceConfigs2.LoadKubegresYaml()
	r.kubegresResource.Spec.CustomConfig = customConfig
	r.kubegresResource.Spec.Replicas = &specNbreReplicas

	if backupSchedule != "" {
		r.kubegresResource.Spec.Backup.Schedule = backupSchedule
		r.kubegresResource.Spec.Backup.PvcName = backupPvcName
		r.kubegresResource.Spec.Backup.VolumeMount = backupVolumeMount
	}
}

func (r *SpecBackUpTest) givenKubegresEnvVarIsSetTo(envVarName, envVarVal string) {
	r.resourceModifier.AppendEnvVar(envVarName, envVarVal, r.kubegresResource)
}

func (r *SpecBackUpTest) givenKubegresAnnotationIsSetTo(annotationKey, annotationValue string) {
	r.resourceModifier.AppendAnnotation(annotationKey, annotationValue, r.kubegresResource)
}

func (r *SpecBackUpTest) givenExistingKubegresSpecIsSetTo(customConfig, backupSchedule, backupPvcName, backupVolumeMount string) {
	var err error
	r.kubegresResource, err = r.resourceRetriever.GetKubegres()
	r.kubegresResource.Spec.Backup.Schedule = backupSchedule
	r.kubegresResource.Spec.Backup.PvcName = backupPvcName
	r.kubegresResource.Spec.Backup.VolumeMount = backupVolumeMount
	r.kubegresResource.Spec.CustomConfig = customConfig

	if err != nil {
		log.Println("Error while getting Kubegres resource : ", err)
		Expect(err).Should(Succeed())
		return
	}
}

func (r *SpecBackUpTest) whenKubegresIsCreated() {
	if r.kubegresResource == nil {
		r.kubegresResource = resourceConfigs2.LoadKubegresYaml()
	}
	r.resourceCreator.CreateKubegres(r.kubegresResource)
}

func (r *SpecBackUpTest) whenKubernetesIsUpdated() {
	r.resourceCreator.UpdateResource(r.kubegresResource, "Kubegres")
}

func (r *SpecBackUpTest) thenErrorEventShouldBeLogged(specName string) {
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

func (r *SpecBackUpTest) thenErrorEventSayingPvcIsNotDeployed() {
	expectedErrorEvent := util2.EventRecord{
		Eventtype: v12.EventTypeWarning,
		Reason:    "SpecCheckErr",
		Message:   "In the Resources Spec the value of 'spec.Backup.PvcName' has a PersistentVolumeClaim name which is not deployed. Please deploy this PersistentVolumeClaim, otherwise this operator cannot work correctly.",
	}
	Eventually(func() bool {
		_, err := r.resourceRetriever.GetKubegres()
		if err != nil {
			return false
		}
		return eventRecorderTest.CheckEventExist(expectedErrorEvent)

	}, resourceConfigs2.TestTimeout, resourceConfigs2.TestRetryInterval).Should(BeTrue())
}

func (r *SpecBackUpTest) thenPodsStatesShouldBe(nbrePrimary, nbreReplicas int) bool {
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

func (r *SpecBackUpTest) thenBackupCronJobDoesNOTExist() bool {
	return Eventually(func() bool {

		kubegresResources, err := r.resourceRetriever.GetKubegresResources()
		Expect(err).ToNot(HaveOccurred())

		if kubegresResources.BackUpCronJob.Name != "" {
			log.Println("CronJob '" + kubegresResources.BackUpCronJob.Name + "' should be deleted. Waiting...")
			return false
		}
		return true

	}, time.Second*10, time.Second*5).Should(BeTrue())
}

func (r *SpecBackUpTest) thenCronJobExistsWithSpec(expectedConfigMapName,
	expectedBackupSchedule,
	expectedBackupPvcName,
	expectedBackupVolumeMount string) bool {

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

		cronJobSchedule := backUpCronJob.Spec.Schedule
		if expectedBackupSchedule != cronJobSchedule {
			log.Println("CronJob '" + backUpCronJob.Name + "' doesn't have the expected schedule: '" + expectedBackupSchedule + "'. Waiting...")
			return false
		}

		cronJobPvcName := backUpCronJob.Spec.JobTemplate.Spec.Template.Spec.Volumes[0].PersistentVolumeClaim.ClaimName
		if expectedBackupPvcName != cronJobPvcName {
			log.Println("CronJob '" + backUpCronJob.Name + "' doesn't have the expected PVC with name: '" + expectedBackupPvcName + "'. Waiting...")
			return false
		}

		cronJobVolumeMount := backUpCronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath
		if expectedBackupVolumeMount != cronJobVolumeMount {
			log.Println("CronJob '" + backUpCronJob.Name + "' doesn't have the expected volume mount: '" + expectedBackupVolumeMount + "'. Waiting...")
			return false
		}

		return true

	}, time.Second*10, time.Second*5).Should(BeTrue())
}

func (r *SpecBackUpTest) thenCronJobSpecShouldHaveAnnotation(annotationKey, annotationValue string) bool {

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

		if backUpCronJob.Spec.JobTemplate.Spec.Template.Annotations[annotationKey] == annotationValue {
			log.Println("The container of CronJob'" + backUpCronJob.Name + "' has the expected annotation with " +
				"key: '" + annotationKey + "' and value: '" + annotationValue + "' in its metadata.")
			return true
		}

		log.Println("The container of CronJob'" + backUpCronJob.Name + "' does NOT have the expected annotation with " +
			"key: '" + annotationKey + "' and value: '" + annotationValue + "' in its metadata. Waiting...")
		return false

	}, time.Second*10, time.Second*5).Should(BeTrue())
}

func (r *SpecBackUpTest) thenCronJobSpecShouldHaveEnvVar(envVarName, envVarVal string) bool {

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

		for _, envVar := range backUpCronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Env {
			if envVar.Name == envVarName && envVar.Value == envVarVal {
				log.Println("The container of CronJob'" + backUpCronJob.Name + "' has the expected environment variable with " +
					"name: '" + envVarName + "' and value: '" + envVarVal + "' in its Spec.")
				return true
			}
		}

		log.Println("The container of CronJob'" + backUpCronJob.Name + "' does NOT have the expected environment variable with " +
			"name: '" + envVarName + "' and value: '" + envVarVal + "' in its Spec.")
		return false

	}, time.Second*10, time.Second*5).Should(BeTrue())
}
