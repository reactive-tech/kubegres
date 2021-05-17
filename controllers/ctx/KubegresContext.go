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
	"reactive-tech.io/kubegres/api/v1"
	"reactive-tech.io/kubegres/controllers/ctx/log"
	"reactive-tech.io/kubegres/controllers/ctx/status"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

type KubegresContext struct {
	Kubegres *v1.Kubegres
	Status   *status.KubegresStatusWrapper
	Ctx      context.Context
	Log      log.LogWrapper
	Client   client.Client
}

const (
	PrimaryReplicationRole = "primary"
	KindKubegres           = "Kubegres"
	DeploymentOwnerKey     = ".metadata.controller"

	BaseConfigMapName                      = "base-kubegres-config"
	CronJobNamePrefix                      = "backup-"
	DefaultContainerVolumeMount            = "/var/lib/postgresql/data"
	DefaultDatabaseFolder                  = "pgdata"
	EnvVarNamePgData                       = "PGDATA"
	EnvVarNameOfPostgresSuperUserPsw       = "POSTGRES_PASSWORD"
	EnvVarNameOfPostgresReplicationUserPsw = "POSTGRES_REPLICATION_PASSWORD"
)

func (r *KubegresContext) GetServiceResourceName(isPrimary bool) string {
	if isPrimary {
		return r.Kubegres.Name
	}
	return r.Kubegres.Name + "-replica"
}

func (r *KubegresContext) GetStatefulSetResourceName(instanceIndex int32) string {
	return r.Kubegres.Name + "-" + strconv.Itoa(int(instanceIndex))
}
