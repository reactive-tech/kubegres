package statefulset_spec

type StatefulSetSpecDifference struct {
	SpecName string
	Current  string
	Expected string
}

func (r *StatefulSetSpecDifference) IsThereDifference() bool {
	return r.SpecName != ""
}

type StatefulSetSpecDifferences struct {
	Differences []StatefulSetSpecDifference
}

func (r *StatefulSetSpecDifferences) IsThereDifference() bool {
	return len(r.Differences) > 0
}

func (r *StatefulSetSpecDifferences) GetSpecDifferencesAsString() string {
	if !r.IsThereDifference() {
		return ""
	}

	specDifferencesStr := ""
	separator := ""
	for _, specDiff := range r.Differences {
		specDifferencesStr += separator + specDiff.SpecName + ": " + specDiff.Expected
		separator = ", "
	}

	return specDifferencesStr
}
