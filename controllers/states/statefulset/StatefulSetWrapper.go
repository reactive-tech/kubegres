package statefulset

import (
	"errors"
	"k8s.io/api/apps/v1"
	"reactive-tech.io/kubegres/controllers/ctx"
	"sort"
)

type StatefulSetWrapper struct {
	IsDeployed  bool
	IsReady     bool
	StatefulSet v1.StatefulSet
	Pod         PodWrapper
}

func (r *StatefulSetWrapper) Instance() string {
	return r.StatefulSet.Labels[ctx.InstanceLabelKey]
}

type StatefulSetWrappers struct {
	statefulSetsSortedByInstance        []StatefulSetWrapper
	statefulSetsReverseSortedByInstance []StatefulSetWrapper
}

func (r *StatefulSetWrappers) Add(statefulSetWrapper StatefulSetWrapper) {
	r.statefulSetsSortedByInstance = append(r.statefulSetsSortedByInstance, statefulSetWrapper)
	sort.Sort(SortByInstance(r.statefulSetsSortedByInstance))

	r.statefulSetsReverseSortedByInstance = append(r.statefulSetsReverseSortedByInstance, statefulSetWrapper)
	sort.Sort(ReverseSortByInstance(r.statefulSetsReverseSortedByInstance))
}

func (r *StatefulSetWrappers) GetAllSortedByInstance() []StatefulSetWrapper {
	return r.copy(r.statefulSetsSortedByInstance)
}

func (r *StatefulSetWrappers) GetAllReverseSortedByInstance() []StatefulSetWrapper {
	return r.copy(r.statefulSetsReverseSortedByInstance)
}

func (r *StatefulSetWrappers) GetByInstance(instance string) (StatefulSetWrapper, error) {
	for _, statefulSet := range r.statefulSetsSortedByInstance {
		if instance == statefulSet.Instance() {
			return statefulSet, nil
		}
	}

	return StatefulSetWrapper{}, errors.New("Given StatefulSet's instance '" + instance + "' does not exist.")
}

func (r *StatefulSetWrappers) GetByName(statefulSetName string) (StatefulSetWrapper, error) {
	for _, statefulSet := range r.statefulSetsSortedByInstance {
		if statefulSetName == statefulSet.StatefulSet.Name {
			return statefulSet, nil
		}
	}

	return StatefulSetWrapper{}, errors.New("Given StatefulSet's name '" + statefulSetName + "' does not exist.")
}

func (r *StatefulSetWrappers) copy(toCopy []StatefulSetWrapper) []StatefulSetWrapper {
	var copyAll []StatefulSetWrapper
	copy(toCopy, copyAll)
	return toCopy
}
