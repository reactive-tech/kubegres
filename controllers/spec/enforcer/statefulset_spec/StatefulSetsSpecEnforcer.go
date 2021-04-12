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

package statefulset_spec

import (
	apps "k8s.io/api/apps/v1"
	"reactive-tech.io/kubegres/controllers/ctx"
)

type StatefulSetSpecEnforcer interface {
	GetSpecName() string
	CheckForSpecDifference(statefulSet *apps.StatefulSet) StatefulSetSpecDifference
	EnforceSpec(statefulSet *apps.StatefulSet) (wasSpecUpdated bool, err error)
	OnSpecEnforcedSuccessfully(statefulSet *apps.StatefulSet) error
}

type StatefulSetsSpecsEnforcer struct {
	kubegresContext ctx.KubegresContext
	registry        []StatefulSetSpecEnforcer
}

func CreateStatefulSetsSpecsEnforcer(kubegresContext ctx.KubegresContext) StatefulSetsSpecsEnforcer {
	return StatefulSetsSpecsEnforcer{kubegresContext: kubegresContext}
}

func (r *StatefulSetsSpecsEnforcer) AddSpecEnforcer(specEnforcer StatefulSetSpecEnforcer) {
	r.registry = append(r.registry, specEnforcer)
}

func (r *StatefulSetsSpecsEnforcer) CheckForSpecDifferences(statefulSet *apps.StatefulSet) StatefulSetSpecDifferences {

	var specDifferences []StatefulSetSpecDifference

	for _, specEnforcer := range r.registry {
		specDifference := specEnforcer.CheckForSpecDifference(statefulSet)
		if specDifference.IsThereDifference() {
			specDifferences = append(specDifferences, specDifference)
		}
	}

	r.logSpecDifferences(specDifferences, statefulSet)

	return StatefulSetSpecDifferences{
		Differences: specDifferences,
	}
}

func (r *StatefulSetsSpecsEnforcer) EnforceSpec(statefulSet *apps.StatefulSet) error {

	var updatedSpecDifferences []StatefulSetSpecDifference

	for _, specEnforcer := range r.registry {

		specDifference := specEnforcer.CheckForSpecDifference(statefulSet)
		if !specDifference.IsThereDifference() {
			continue
		}

		wasSpecUpdated, err := specEnforcer.EnforceSpec(statefulSet)
		if err != nil {
			return err

		} else if wasSpecUpdated {
			updatedSpecDifferences = append(updatedSpecDifferences, specDifference)
		}
	}

	if len(updatedSpecDifferences) > 0 {
		r.kubegresContext.Log.Info("Updating Spec of a StatefulSet", "StatefulSet name", statefulSet.Name)
		err := r.kubegresContext.Client.Update(r.kubegresContext.Ctx, statefulSet)
		if err != nil {
			return err
		}

		r.logUpdatedSpecDifferences(updatedSpecDifferences, statefulSet)
	}

	return nil
}

func (r *StatefulSetsSpecsEnforcer) OnSpecUpdatedSuccessfully(statefulSet *apps.StatefulSet) error {
	for _, specEnforcer := range r.registry {
		if err := specEnforcer.OnSpecEnforcedSuccessfully(statefulSet); err != nil {
			return err
		}
	}
	return nil
}

func (r *StatefulSetsSpecsEnforcer) logSpecDifferences(specDifferences []StatefulSetSpecDifference, statefulSet *apps.StatefulSet) {
	for _, specDifference := range specDifferences {
		r.kubegresContext.Log.InfoEvent("StatefulSetOperation", "The Spec is NOT up-to-date for a StatefulSet.", "StatefulSet name", statefulSet.Name, "SpecName", specDifference.SpecName, "Expected", specDifference.Expected, "Current", specDifference.Current)
	}
}

func (r *StatefulSetsSpecsEnforcer) logUpdatedSpecDifferences(specDiffBeforeUpdate []StatefulSetSpecDifference, statefulSet *apps.StatefulSet) {
	for _, specDifference := range specDiffBeforeUpdate {
		r.kubegresContext.Log.InfoEvent("StatefulSetOperation", "Updated a StatefulSet with up-to-date Spec.", "StatefulSet name", statefulSet.Name, "SpecName", specDifference.SpecName, "New", specDifference.Expected)
	}
}
