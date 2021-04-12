package operation

const (
	errorTypeThereIsAlreadyAnActiveOperation  = "There is already an active operation which is running. We cannot have more than 1 active operation running."
	errorTypeOperationIdHasNoAssociatedConfig = "The given operationId has not an associated config. Please associate it by calling the method BlockingOperation.AddConfig()."
)

type BlockingOperationError struct {
	ThereIsAlreadyAnActiveOperation  bool
	OperationIdHasNoAssociatedConfig bool

	errorType string
}

func CreateBlockingOperationError(errorType, operationId string) *BlockingOperationError {
	return &BlockingOperationError{
		errorType:                        "OperationId: '" + operationId + "' - " + errorType,
		ThereIsAlreadyAnActiveOperation:  errorType == errorTypeThereIsAlreadyAnActiveOperation,
		OperationIdHasNoAssociatedConfig: errorType == errorTypeOperationIdHasNoAssociatedConfig,
	}
}

func (r *BlockingOperationError) Error() string {
	return "Cannot active a blocking operation. Reason: " + r.errorType
}
