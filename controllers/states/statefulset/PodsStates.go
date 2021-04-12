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

package statefulset

import (
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"reactive-tech.io/kubegres/controllers/ctx"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

type PodStates struct {
	pods            []PodWrapper
	kubegresContext ctx.KubegresContext
}

type PodWrapper struct {
	IsDeployed    bool
	IsReady       bool
	IsStuck       bool
	InstanceIndex int32
	Pod           core.Pod
}

func loadPodsStates(kubegresContext ctx.KubegresContext) (PodStates, error) {
	podStates := PodStates{kubegresContext: kubegresContext}
	err := podStates.loadStates()
	return podStates, err
}

func (r *PodStates) loadStates() (err error) {

	deployedPods, err := r.getDeployedPods()
	if err != nil {
		return err
	}

	for _, pod := range deployedPods.Items {

		isPodReady := r.isPodReady(pod)
		isPodStuck := r.isPodStuck(pod)

		podWrapper := PodWrapper{
			IsDeployed:    true,
			IsReady:       isPodReady && !isPodStuck,
			IsStuck:       isPodStuck,
			InstanceIndex: r.getInstanceIndex(pod),
			Pod:           pod,
		}

		r.pods = append(r.pods, podWrapper)
	}

	return nil
}

func (r *PodStates) getDeployedPods() (*core.PodList, error) {

	list := &core.PodList{}
	opts := []client.ListOption{
		client.InNamespace(r.kubegresContext.Kubegres.Namespace),
		client.MatchingLabels{"app": r.kubegresContext.Kubegres.Name},
	}
	err := r.kubegresContext.Client.List(r.kubegresContext.Ctx, list, opts...)

	if err != nil {
		if apierrors.IsNotFound(err) {
			r.kubegresContext.Log.Info("There is not any deployed Pods yet", "Kubegres name", r.kubegresContext.Kubegres.Name)
			err = nil
		} else {
			r.kubegresContext.Log.ErrorEvent("PodLoadingErr", err, "Unable to load any deployed Pods.", "Kubegres name", r.kubegresContext.Kubegres.Name)
		}
	}

	return list, err
}

func (r *PodStates) isPodReady(pod core.Pod) bool {

	if len(pod.Status.ContainerStatuses) == 0 {
		return false
	}

	return pod.Status.ContainerStatuses[0].Ready
}

func (r *PodStates) isPodStuck(pod core.Pod) bool {

	if len(pod.Status.ContainerStatuses) == 0 ||
		pod.Status.ContainerStatuses[0].State.Waiting == nil {
		return false
	}

	waitingReason := pod.Status.ContainerStatuses[0].State.Waiting.Reason
	if waitingReason == "CrashLoopBackOff" || waitingReason == "Error" {
		r.kubegresContext.Log.Info("POD is waiting", "Reason", waitingReason)
		return true
	}

	return false
}

func (r *PodStates) getInstanceIndex(pod core.Pod) int32 {
	instanceIndex, _ := strconv.ParseInt(pod.Labels["index"], 10, 32)
	return int32(instanceIndex)
}
