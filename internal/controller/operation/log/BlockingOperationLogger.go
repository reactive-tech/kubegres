package log

import (
	v1 "reactive-tech.io/kubegres/api/v1"
	"reactive-tech.io/kubegres/internal/controller/ctx"
	"reactive-tech.io/kubegres/internal/controller/operation"
)

type BlockingOperationLogger struct {
	kubegresContext   ctx.KubegresContext
	blockingOperation *operation.BlockingOperation
}

func CreateBlockingOperationLogger(kubegresContext ctx.KubegresContext, blockingOperation *operation.BlockingOperation) BlockingOperationLogger {
	return BlockingOperationLogger{kubegresContext: kubegresContext, blockingOperation: blockingOperation}
}

func (r *BlockingOperationLogger) Log() {
	r.logActiveOperation()
	r.logPreviouslyActiveOperation()
}

func (r *BlockingOperationLogger) logActiveOperation() {

	activeOperation := r.blockingOperation.GetActiveOperation()
	activeOperationId := activeOperation.OperationId
	nbreSecondsLeftBeforeTimeOut := r.blockingOperation.GetNbreSecondsLeftBeforeTimeOut()

	keyAndValuesToLog := r.logOperation(activeOperation)
	keyAndValuesToLog = append(keyAndValuesToLog, "NbreSecondsLeftBeforeTimeOut", nbreSecondsLeftBeforeTimeOut)

	if activeOperationId == "" {
		r.kubegresContext.Log.Info("Active Blocking-Operation: None")
	} else {
		r.kubegresContext.Log.Info("Active Blocking-Operation ", keyAndValuesToLog...)
	}
}

func (r *BlockingOperationLogger) logPreviouslyActiveOperation() {

	previousActiveOperation := r.blockingOperation.GetPreviouslyActiveOperation()
	operationId := previousActiveOperation.OperationId

	keyAndValuesToLog := r.logOperation(previousActiveOperation)

	if operationId == "" {
		r.kubegresContext.Log.Info("Previous Blocking-Operation: None")
	} else {
		r.kubegresContext.Log.Info("Previous Blocking-Operation ", keyAndValuesToLog...)
	}
}

func (r *BlockingOperationLogger) logOperation(operation v1.KubegresBlockingOperation) []interface{} {

	operationId := operation.OperationId
	stepId := operation.StepId
	hasTimedOut := operation.HasTimedOut
	statefulSetInstanceIndex := operation.StatefulSetOperation.InstanceIndex
	statefulSetSpecDifferences := operation.StatefulSetSpecUpdateOperation.SpecDifferences

	var keysAndValues []interface{}
	keysAndValues = append(keysAndValues, "OperationId", operationId)
	keysAndValues = append(keysAndValues, "StepId", stepId)
	keysAndValues = append(keysAndValues, "HasTimedOut", hasTimedOut)

	if statefulSetInstanceIndex != 0 {
		keysAndValues = append(keysAndValues, "StatefulSetInstanceIndex", statefulSetInstanceIndex)
	}

	if statefulSetSpecDifferences != "" {
		keysAndValues = append(keysAndValues, "StatefulSetSpecDifferences", statefulSetSpecDifferences)
	}

	return keysAndValues
}
