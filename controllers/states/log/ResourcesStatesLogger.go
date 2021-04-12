package log

import (
	"reactive-tech.io/kubegres/controllers/ctx"
	"reactive-tech.io/kubegres/controllers/states"
	"reactive-tech.io/kubegres/controllers/states/statefulset"
)

type ResourcesStatesLogger struct {
	kubegresContext ctx.KubegresContext
	resourcesStates states.ResourcesStates
}

func CreateResourcesStatesLogger(kubegresContext ctx.KubegresContext, resourcesStates states.ResourcesStates) ResourcesStatesLogger {
	return ResourcesStatesLogger{kubegresContext: kubegresContext, resourcesStates: resourcesStates}
}

func (r *ResourcesStatesLogger) Log() {
	r.logDbStorageClassStates()
	r.logConfigStates()
	r.logStatefulSetsStates()
	r.logServicesStates()
	r.logBackUpStates()
}

func (r *ResourcesStatesLogger) logDbStorageClassStates() {
	r.kubegresContext.Log.Info("Database StorageClass states.",
		"IsDeployed", r.resourcesStates.DbStorageClass.IsDeployed,
		"name", r.resourcesStates.DbStorageClass.StorageClassName)
}

func (r *ResourcesStatesLogger) logConfigStates() {
	r.kubegresContext.Log.Info("Base Config states",
		"IsDeployed", r.resourcesStates.Config.IsBaseConfigDeployed,
		"name", r.resourcesStates.Config.BaseConfigName)

	if r.resourcesStates.Config.BaseConfigName != r.resourcesStates.Config.CustomConfigName {
		r.kubegresContext.Log.Info("Custom Config states",
			"IsDeployed", r.resourcesStates.Config.IsCustomConfigDeployed,
			"name", r.resourcesStates.Config.CustomConfigName)
	}
}

func (r *ResourcesStatesLogger) logStatefulSetsStates() {
	statefulSets := r.resourcesStates.StatefulSets
	r.kubegresContext.Log.Info("All StatefulSets deployment states: ",
		"Spec expected to deploy", statefulSets.SpecNbreToDeploy,
		"Nbre Deployed", statefulSets.NbreDeployed)

	r.logStatefulSetWrapper("Primary states", statefulSets.Primary)

	for _, replicaStatefulSetWrapper := range statefulSets.Replicas.All.GetAllSortedByInstanceIndex() {
		r.logStatefulSetWrapper("Replica states", replicaStatefulSetWrapper)
	}
}

func (r *ResourcesStatesLogger) logStatefulSetWrapper(logLabel string, statefulSetWrapper statefulset.StatefulSetWrapper) {

	if !statefulSetWrapper.IsDeployed {
		r.kubegresContext.Log.Info(logLabel+": ",
			"IsDeployed", statefulSetWrapper.IsDeployed)

	} else {
		r.kubegresContext.Log.Info(logLabel+": ",
			"IsDeployed", statefulSetWrapper.IsDeployed,
			"Name", statefulSetWrapper.StatefulSet.Name,
			"IsReady", statefulSetWrapper.IsReady,
			"Pod Name", statefulSetWrapper.Pod.Pod.Name,
			"Pod IsDeployed", statefulSetWrapper.Pod.IsDeployed,
			"Pod IsReady", statefulSetWrapper.Pod.IsReady,
			"Pod IsStuck", statefulSetWrapper.Pod.IsStuck)
	}
}

func (r *ResourcesStatesLogger) logServicesStates() {
	r.logServiceWrapper("Primary Service states", r.resourcesStates.Services.Primary)
	r.logServiceWrapper("Replica Service states", r.resourcesStates.Services.Replica)
}

func (r *ResourcesStatesLogger) logServiceWrapper(logLabel string, serviceWrapper states.ServiceWrapper) {

	if !serviceWrapper.IsDeployed {
		r.kubegresContext.Log.Info(logLabel+": ",
			"IsDeployed", serviceWrapper.IsDeployed)

	} else {
		r.kubegresContext.Log.Info(logLabel+": ",
			"IsDeployed", serviceWrapper.IsDeployed,
			"name", serviceWrapper.Name)
	}
}

func (r *ResourcesStatesLogger) logBackUpStates() {
	r.kubegresContext.Log.Info("BackUp states.",
		"IsCronJobDeployed", r.resourcesStates.BackUp.IsCronJobDeployed,
		"IsPvcDeployed", r.resourcesStates.BackUp.IsPvcDeployed,
		"CronJobLastScheduleTime", r.resourcesStates.BackUp.CronJobLastScheduleTime)
}
