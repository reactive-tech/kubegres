/*
Copyright 2023.

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

package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ----------------------- SPEC -------------------------------------------

type KubegresDatabase struct {
	Size             string  `json:"size,omitempty"`
	VolumeMount      string  `json:"volumeMount,omitempty"`
	StorageClassName *string `json:"storageClassName,omitempty"`
}

type KubegresBackUp struct {
	Schedule    string `json:"schedule,omitempty"`
	VolumeMount string `json:"volumeMount,omitempty"`
	PvcName     string `json:"pvcName,omitempty"`
}

type KubegresFailover struct {
	IsDisabled bool   `json:"isDisabled,omitempty"`
	PromotePod string `json:"promotePod,omitempty"`
}

type KubegresScheduler struct {
	Affinity    *v1.Affinity    `json:"affinity,omitempty"`
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
}

type VolumeClaimTemplate struct {
	Name string                       `json:"name,omitempty"`
	Spec v1.PersistentVolumeClaimSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

type Volume struct {
	VolumeMounts         []v1.VolumeMount      `json:"volumeMounts,omitempty"`
	Volumes              []v1.Volume           `json:"volumes,omitempty"`
	VolumeClaimTemplates []VolumeClaimTemplate `json:"volumeClaimTemplates,omitempty"`
}

type Probe struct {
	LivenessProbe  *v1.Probe `json:"livenessProbe,omitempty"`
	ReadinessProbe *v1.Probe `json:"readinessProbe,omitempty"`
}

type KubegresSpec struct {
	Replicas           *int32                    `json:"replicas,omitempty"`
	Image              string                    `json:"image,omitempty"`
	Port               int32                     `json:"port,omitempty"`
	ImagePullSecrets   []v1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	CustomConfig       string                    `json:"customConfig,omitempty"`
	Database           KubegresDatabase          `json:"database,omitempty"`
	Failover           KubegresFailover          `json:"failover,omitempty"`
	Backup             KubegresBackUp            `json:"backup,omitempty"`
	Env                []v1.EnvVar               `json:"env,omitempty"`
	Scheduler          KubegresScheduler         `json:"scheduler,omitempty"`
	Resources          v1.ResourceRequirements   `json:"resources,omitempty"`
	Volume             Volume                    `json:"volume,omitempty"`
	SecurityContext    *v1.PodSecurityContext    `json:"securityContext,omitempty"`
	Probe              Probe                     `json:"probe,omitempty"`
	ServiceAccountName string                    `json:"serviceAccountName,omitempty"`
}

// ----------------------- STATUS -----------------------------------------

type KubegresStatefulSetOperation struct {
	InstanceIndex int32  `json:"instanceIndex,omitempty"`
	Name          string `json:"name,omitempty"`
}

type KubegresStatefulSetSpecUpdateOperation struct {
	SpecDifferences string `json:"specDifferences,omitempty"`
}

type KubegresBlockingOperation struct {
	OperationId          string `json:"operationId,omitempty"`
	StepId               string `json:"stepId,omitempty"`
	TimeOutEpocInSeconds int64  `json:"timeOutEpocInSeconds,omitempty"`
	HasTimedOut          bool   `json:"hasTimedOut,omitempty"`

	// Custom operation fields
	StatefulSetOperation           KubegresStatefulSetOperation           `json:"statefulSetOperation,omitempty"`
	StatefulSetSpecUpdateOperation KubegresStatefulSetSpecUpdateOperation `json:"statefulSetSpecUpdateOperation,omitempty"`
}

type KubegresStatus struct {
	LastCreatedInstanceIndex  int32                     `json:"lastCreatedInstanceIndex,omitempty"`
	BlockingOperation         KubegresBlockingOperation `json:"blockingOperation,omitempty"`
	PreviousBlockingOperation KubegresBlockingOperation `json:"previousBlockingOperation,omitempty"`
	EnforcedReplicas          int32                     `json:"enforcedReplicas,omitempty"`
}

// ----------------------- RESOURCE ---------------------------------------

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Kubegres is the Schema for the kubegres API
type Kubegres struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubegresSpec   `json:"spec,omitempty"`
	Status KubegresStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KubegresList contains a list of Kubegres
type KubegresList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kubegres `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Kubegres{}, &KubegresList{})
}
