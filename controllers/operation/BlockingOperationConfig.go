package operation

import v1 "reactive-tech.io/kubegres/api/v1"

type IsOperationCompleted func(operation v1.KubegresBlockingOperation) bool

type BlockingOperationConfig struct {
	OperationId       string
	StepId            string
	TimeOutInSeconds  int64
	CompletionChecker IsOperationCompleted

	// This flag is set to false by default, meaning once an operation is completed (operation = operationId + stepId),
	// it will be automatically removed as active operation and added as previous active operation.
	//
	// If this flag is set to false and the operation times-out, it is not removed and remain as an active operation
	// until it is manually removed.
	//
	// If this flag is set to true, after the completion of a stepId, we will add a new stepId and keep the same operationId.
	// The new stepId will be set to the value of the constant "BlockingOperation.transition_step_id".
	// This logic allows other types of operations (e.g. FailOver) to not start until an active operation (e.g. Spec update)
	// is either terminated manually or is waiting for the next step to start.
	AfterCompletionMoveToTransitionStep bool
}

const (
	TransitionOperationStepId = "Transition step: waiting either for the next step to start or for the operation to be removed ..."

	OperationIdBaseConfigCountSpecEnforcement = "Base config count spec enforcement"
	OperationStepIdBaseConfigDeploying        = "Base config is deploying"

	OperationIdPrimaryDbCountSpecEnforcement         = "Primary DB count spec enforcement"
	OperationStepIdPrimaryDbDeploying                = "Primary DB is deploying"
	OperationStepIdPrimaryDbWaitingBeforeFailingOver = "Waiting few seconds before failing over by promoting a Replica DB as a Primary DB"
	OperationStepIdPrimaryDbFailingOver              = "Failing over by promoting a Replica DB as a Primary DB"

	OperationIdReplicaDbCountSpecEnforcement = "Replica DB count spec enforcement"
	OperationStepIdReplicaDbDeploying        = "Replica DB is deploying"
	OperationStepIdReplicaDbUndeploying      = "Replica DB is undeploying"

	OperationIdStatefulSetSpecEnforcing         = "Enforcing StatefulSet's Spec"
	OperationStepIdStatefulSetSpecUpdating      = "StatefulSet's spec is updating"
	OperationStepIdStatefulSetPodSpecUpdating   = "StatefulSet Pod's spec is updating"
	OperationStepIdStatefulSetWaitingOnStuckPod = "Attempting to fix a stuck Pod by recreating it"
)
