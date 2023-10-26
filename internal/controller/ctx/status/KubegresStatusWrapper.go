package status

import (
	"context"
	v1 "reactive-tech.io/kubegres/api/v1"
	"reactive-tech.io/kubegres/internal/controller/ctx/log"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KubegresStatusWrapper struct {
	Kubegres             *v1.Kubegres
	Ctx                  context.Context
	Log                  log.LogWrapper
	Client               client.Client
	statusFieldsToUpdate map[string]interface{}
}

func (r *KubegresStatusWrapper) GetLastCreatedInstanceIndex() int32 {
	return r.Kubegres.Status.LastCreatedInstanceIndex
}

func (r *KubegresStatusWrapper) SetLastCreatedInstanceIndex(value int32) {
	r.addStatusFieldToUpdate("LastCreatedInstanceIndex", value)
	r.Kubegres.Status.LastCreatedInstanceIndex = value
}

func (r *KubegresStatusWrapper) GetBlockingOperation() v1.KubegresBlockingOperation {
	return r.Kubegres.Status.BlockingOperation
}

func (r *KubegresStatusWrapper) SetBlockingOperation(value v1.KubegresBlockingOperation) {
	r.addStatusFieldToUpdate("BlockingOperation", value)
	r.Kubegres.Status.BlockingOperation = value
}

func (r *KubegresStatusWrapper) GetEnforcedReplicas() int32 {
	return r.Kubegres.Status.EnforcedReplicas
}

func (r *KubegresStatusWrapper) SetEnforcedReplicas(value int32) {
	r.addStatusFieldToUpdate("EnforcedReplicas", value)
	r.Kubegres.Status.EnforcedReplicas = value
}

func (r *KubegresStatusWrapper) GetPreviousBlockingOperation() v1.KubegresBlockingOperation {
	return r.Kubegres.Status.PreviousBlockingOperation
}

func (r *KubegresStatusWrapper) SetPreviousBlockingOperation(value v1.KubegresBlockingOperation) {
	r.addStatusFieldToUpdate("PreviousBlockingOperation", value)
	r.Kubegres.Status.PreviousBlockingOperation = value
}

func (r *KubegresStatusWrapper) UpdateStatusIfChanged() error {
	if r.statusFieldsToUpdate == nil {
		return nil
	}

	for statusFieldName, statusFieldValue := range r.statusFieldsToUpdate {
		r.Log.Info("Updating Kubegres' status: ",
			"Field", statusFieldName, "New value", statusFieldValue)

	}

	err := r.Client.Status().Update(r.Ctx, r.Kubegres)

	if err != nil {
		r.Log.Error(err, "Failed to update Kubegres status")

	} else {
		r.Log.Info("Kubegres status updated.")
	}

	return err
}

func (r *KubegresStatusWrapper) addStatusFieldToUpdate(statusFieldName string, newValue interface{}) {

	if r.statusFieldsToUpdate == nil {
		r.statusFieldsToUpdate = make(map[string]interface{})
	}

	r.statusFieldsToUpdate[statusFieldName] = newValue
}
