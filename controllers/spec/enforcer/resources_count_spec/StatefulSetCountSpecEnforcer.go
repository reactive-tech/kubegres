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

package resources_count_spec

import (
	"reactive-tech.io/kubegres/controllers/spec/enforcer/resources_count_spec/statefulset"
)

type StatefulSetCountSpecEnforcer struct {
	primaryDbCountSpecEnforcer statefulset.PrimaryDbCountSpecEnforcer
	replicaDbCountSpecEnforcer statefulset.ReplicaDbCountSpecEnforcer
}

func CreateStatefulSetCountSpecEnforcer(primaryDbCountSpecEnforcer statefulset.PrimaryDbCountSpecEnforcer,
	replicaDbCountSpecEnforcer statefulset.ReplicaDbCountSpecEnforcer) StatefulSetCountSpecEnforcer {

	return StatefulSetCountSpecEnforcer{
		primaryDbCountSpecEnforcer: primaryDbCountSpecEnforcer,
		replicaDbCountSpecEnforcer: replicaDbCountSpecEnforcer,
	}
}

func (r *StatefulSetCountSpecEnforcer) EnforceSpec() error {
	if err := r.enforcePrimaryDbInstance(); err != nil {
		return err
	}
	return r.enforceReplicaDbInstances()
}

func (r *StatefulSetCountSpecEnforcer) enforcePrimaryDbInstance() error {
	return r.primaryDbCountSpecEnforcer.Enforce()
}

func (r *StatefulSetCountSpecEnforcer) enforceReplicaDbInstances() error {
	return r.replicaDbCountSpecEnforcer.Enforce()
}
