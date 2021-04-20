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

type KubegresSpec struct {
	Replicas         *int32                    `json:"replicas,omitempty"`
	Image            string                    `json:"image,omitempty"`
	Port             int32                     `json:"port,omitempty"`
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	CustomConfig string           `json:"customConfig,omitempty"`
	Database     KubegresDatabase `json:"database,omitempty"`
	Backup       KubegresBackUp   `json:"backup,omitempty"`
	Env          []v1.EnvVar      `json:"env,omitempty"`
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
}

// ----------------------- RESOURCE ---------------------------------------

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// Kubegres is the Schema for the kubegres API
type Kubegres struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubegresSpec   `json:"spec,omitempty"`
	Status KubegresStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KubegresList contains a list of Kubegres
type KubegresList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kubegres `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Kubegres{}, &KubegresList{})
}
