package statefulset_spec

import (
	"reflect"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"reactive-tech.io/kubegres/internal/controller/ctx"
)

type ContainerSecurityContextSpecEnforcer struct {
	KubegresContext ctx.KubegresContext
}

func CreateContainerSecurityContextSpecEnforcer(kubegresContext ctx.KubegresContext) ContainerSecurityContextSpecEnforcer {
	return ContainerSecurityContextSpecEnforcer{KubegresContext: kubegresContext}
}

func (r *ContainerSecurityContextSpecEnforcer) GetSpecName() string {
	return "ContainerSecurityContext"
}

func (r *ContainerSecurityContextSpecEnforcer) CheckForSpecDifference(statefulSet *apps.StatefulSet) StatefulSetSpecDifference {
	current := statefulSet.Spec.Template.Spec.Containers[0].SecurityContext
	expected := r.KubegresContext.Kubegres.Spec.ContainerSecurityContext
	emptyContainerSecurityContext := &v1.SecurityContext{}

	if expected == nil && reflect.DeepEqual(current, emptyContainerSecurityContext) {
		return StatefulSetSpecDifference{}
	}

	if !reflect.DeepEqual(current, expected) {
		return StatefulSetSpecDifference{
			SpecName: r.GetSpecName(),
			Current:  current.String(),
			Expected: expected.String(),
		}
	}

	return StatefulSetSpecDifference{}
}

func (r *ContainerSecurityContextSpecEnforcer) EnforceSpec(statefulSet *apps.StatefulSet) (wasSpecUpdated bool, err error) {

	statefulSet.Spec.Template.Spec.Containers[0].SecurityContext = r.KubegresContext.Kubegres.Spec.ContainerSecurityContext

	if len(statefulSet.Spec.Template.Spec.InitContainers) > 0 {
		statefulSet.Spec.Template.Spec.InitContainers[0].SecurityContext = r.KubegresContext.Kubegres.Spec.ContainerSecurityContext
	}

	return true, nil
}

func (r *ContainerSecurityContextSpecEnforcer) OnSpecEnforcedSuccessfully(statefulSet *apps.StatefulSet) error {
	return nil
}
