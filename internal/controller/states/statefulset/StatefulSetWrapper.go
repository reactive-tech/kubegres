package statefulset

import (
	"errors"
	"k8s.io/api/apps/v1"
	"sort"
	"strconv"
)

type StatefulSetWrapper struct {
	IsDeployed    bool
	IsReady       bool
	InstanceIndex int32
	StatefulSet   v1.StatefulSet
	Pod           PodWrapper
}

type StatefulSetWrappers struct {
	statefulSetsSortedByInstanceIndex        []StatefulSetWrapper
	statefulSetsReverseSortedByInstanceIndex []StatefulSetWrapper
}

func (r *StatefulSetWrappers) Add(statefulSetWrapper StatefulSetWrapper) {

	r.statefulSetsSortedByInstanceIndex = append(r.statefulSetsSortedByInstanceIndex, statefulSetWrapper)
	sort.Sort(SortByInstanceIndex(r.statefulSetsSortedByInstanceIndex))

	r.statefulSetsReverseSortedByInstanceIndex = append(r.statefulSetsReverseSortedByInstanceIndex, statefulSetWrapper)
	sort.Sort(ReverseSortByInstanceIndex(r.statefulSetsReverseSortedByInstanceIndex))
}

func (r *StatefulSetWrappers) GetAllSortedByInstanceIndex() []StatefulSetWrapper {
	return r.copy(r.statefulSetsSortedByInstanceIndex)
}

func (r *StatefulSetWrappers) GetAllReverseSortedByInstanceIndex() []StatefulSetWrapper {
	return r.copy(r.statefulSetsReverseSortedByInstanceIndex)
}

func (r *StatefulSetWrappers) GetByInstanceIndex(instanceIndex int32) (StatefulSetWrapper, error) {

	for _, statefulSet := range r.statefulSetsSortedByInstanceIndex {
		if instanceIndex == statefulSet.InstanceIndex {
			return statefulSet, nil
		}
	}

	return StatefulSetWrapper{}, errors.New("Given StatefulSet's instanceIndex '" + strconv.Itoa(int(instanceIndex)) + "' does not exist.")
}

func (r *StatefulSetWrappers) GetByName(statefulSetName string) (StatefulSetWrapper, error) {

	for _, statefulSet := range r.statefulSetsSortedByInstanceIndex {
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

/*
func (r *StatefulSetWrappers) GetByArrayIndex(arrayIndex int32) (StatefulSetWrapper, error) {

	arrayLength := len(r.statefulSets)
	if arrayIndex < 0 || int(arrayIndex) >= arrayLength {
		return StatefulSetWrapper{}, errors.New("Given StatefulSet's arrayIndex '" + strconv.Itoa(int(arrayIndex)) + "' does not exist.")
	}

	return r.statefulSets[arrayIndex], nil
}
*/
