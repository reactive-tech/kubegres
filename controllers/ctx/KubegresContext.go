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

package ctx

import (
	"context"
	apps "k8s.io/api/apps/v1"
	"reactive-tech.io/kubegres/api/v1"
	"reactive-tech.io/kubegres/controllers/ctx/log"
	"reactive-tech.io/kubegres/controllers/ctx/status"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
)

type KubegresContext struct {
	Kubegres *v1.Kubegres
	Status   *status.KubegresStatusWrapper
	Ctx      context.Context
	Log      log.LogWrapper
	Client   client.Client
}

const (
	PrimaryRoleName                        = "primary"
	KindKubegres                           = "Kubegres"
	DeploymentOwnerKey                     = ".metadata.controller"
	DatabaseVolumeName                     = "postgres-db"
	BaseConfigMapVolumeName                = "base-config"
	CustomConfigMapVolumeName              = "custom-config"
	BaseConfigMapName                      = "base-kubegres-config"
	CronJobNamePrefix                      = "backup-"
	DefaultContainerPortNumber             = 5432
	DefaultDatabaseVolumeMount             = "/var/lib/postgresql/data"
	DefaultDatabaseFolder                  = "pgdata"
	EnvVarNamePgData                       = "PGDATA"
	EnvVarNameOfPostgresSuperUserPsw       = "POSTGRES_PASSWORD"
	EnvVarNameOfPostgresReplicationUserPsw = "POSTGRES_REPLICATION_PASSWORD"
	NameLabelKey                           = "app.kubernetes.io/name"
	InstanceLabelKey                       = "app.kubernetes.io/instance"
	ReplicationRoleLabelKey                = "app.kubegres.io/replication-role"
)

func (r *KubegresContext) GetServiceResourceName(isPrimary bool) string {
	if isPrimary {
		return r.Kubegres.Name
	}
	return r.Kubegres.Name + "-replica"
}

func (r *KubegresContext) GetStatefulSetResourceName(instance string) string {
	return r.Kubegres.Name + "-" + instance
}

func (r *KubegresContext) IsReservedVolumeName(volumeName string) bool {
	return volumeName == DatabaseVolumeName ||
		volumeName == BaseConfigMapVolumeName ||
		volumeName == CustomConfigMapVolumeName ||
		strings.Contains(volumeName, "kube-api")
}

func (r *KubegresContext) ReplicasCount() int32 {
	if r.Kubegres.Spec.NodeSets == nil {
		return *r.Kubegres.Spec.Replicas
	}
	return int32(len(r.Kubegres.Spec.NodeSets))
}

func (r *KubegresContext) GetNodeSetsFromSpec() []v1.KubegresNodeSet {
	if r.Kubegres.Spec.NodeSets == nil {
		nodeSets := make([]v1.KubegresNodeSet, *r.Kubegres.Spec.Replicas)
		for i := int32(0); i < *r.Kubegres.Spec.Replicas; i += 1 {
			nodeSets[i] = v1.KubegresNodeSet{
				Name: strconv.Itoa(int(i)),
			}
		}
		return nodeSets
	}
	return r.Kubegres.Spec.NodeSets
}

func (r *KubegresContext) GetInstanceFromStatefulSet(statefulSet apps.StatefulSet) string {
	return statefulSet.Labels[InstanceLabelKey]
}

func (r *KubegresContext) GetNodeSetSpecFromInstance(instance string) *v1.KubegresNodeSet {
	for _, nodeSet := range r.Kubegres.Spec.NodeSets {
		if nodeSet.Name == instance {
			return &nodeSet
		}
	}
	return &v1.KubegresNodeSet{
		Name: instance,
	}
}
