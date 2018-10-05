package v1

type ErrorCode string

const (
	ErrorCodeUnknown  ErrorCode = "UNKNOWN"
	ErrorCodeConflict ErrorCode = "CONFLICT"

	ErrorCodeInvalidBuildID ErrorCode = "INVALID_BUILD_ID"

	ErrorCodeInvalidDeployID ErrorCode = "INVALID_DEPLOY_ID"

	ErrorCodeInvalidJobID    ErrorCode = "INVALID_JOB_ID"
	ErrorCodeInvalidJobRunID ErrorCode = "INVALID_JOB_RUN_ID"

	ErrorCodeInvalidSecret ErrorCode = "INVALID_SECRET"

	ErrorCodeInvalidServiceID         ErrorCode = "INVALID_SERVICE_ID"
	ErrorCodeInvalidServiceInstanceID ErrorCode = "INVALID_SERVICE_INSTANCE_ID"

	ErrorCodeSystemAlreadyExists  ErrorCode = "SYSTEM_ALREADY_EXISTS"
	ErrorCodeInvalidSystemID      ErrorCode = "INVALID_SYSTEM_ID"
	ErrorCodeSystemDeleting       ErrorCode = "SYSTEM_DELETING"
	ErrorCodeSystemFailed         ErrorCode = "SYSTEM_FAILED"
	ErrorCodeSystemPending        ErrorCode = "SYSTEM_PENDING"
	ErrorCodeInvalidSystemOptions ErrorCode = "INVALID_SYSTEM_OPTIONS"

	ErrorCodeInvalidTeardownID ErrorCode = "INVALID_TEARDOWN_ID"

	ErrorCodeInvalidComponentType ErrorCode = "INVALID_COMPONENT_TYP"
	ErrorCodeInvalidPath          ErrorCode = "INVALID_PATH"
	ErrorCodeInvalidSidecar       ErrorCode = "INVALID_SIDECAR"
	ErrorCodeInvalidVersion       ErrorCode = "INVALID_VERSION"
	ErrorCodeLogsUnavailable      ErrorCode = "LOGS_UNAVAILABLE"
)

type Error struct {
	Code ErrorCode `json:"code"`
}

func NewError(code ErrorCode) *Error {
	return &Error{code}
}

func (e *Error) Error() string {
	return string(e.Code)
}

func NewUnknownError() *Error {
	return NewError(ErrorCodeUnknown)
}

func NewConflictError() *Error {
	return NewError(ErrorCodeConflict)
}

func NewInvalidBuildIDError() *Error {
	return NewError(ErrorCodeInvalidBuildID)
}

func NewInvalidDeployIDError() *Error {
	return NewError(ErrorCodeInvalidDeployID)
}

func NewInvalidJobIDError() *Error {
	return NewError(ErrorCodeInvalidJobID)
}

func NewInvalidJobRunIDError() *Error {
	return NewError(ErrorCodeInvalidJobRunID)
}

func NewInvalidSecretError() *Error {
	return NewError(ErrorCodeInvalidSecret)
}

func NewInvalidServiceIDError() *Error {
	return NewError(ErrorCodeInvalidServiceID)
}

func NewInvalidServiceInstanceIDError() *Error {
	return NewError(ErrorCodeInvalidServiceInstanceID)
}

func NewSystemAlreadyExistsError() *Error {
	return NewError(ErrorCodeSystemAlreadyExists)
}

func NewSystemDeletingError() *Error {
	return NewError(ErrorCodeSystemDeleting)
}

func NewSystemFailedError() *Error {
	return NewError(ErrorCodeSystemFailed)
}

func NewSystemPendingError() *Error {
	return NewError(ErrorCodeSystemPending)
}

func NewInvalidSystemOptionsError() *Error {
	return NewError(ErrorCodeInvalidSystemOptions)
}

func NewInvalidSystemIDError() *Error {
	return NewError(ErrorCodeInvalidSystemID)
}

func NewInvalidTeardownIDError() *Error {
	return NewError(ErrorCodeInvalidTeardownID)
}

func NewInvalidComponentTypeError() *Error {
	return NewError(ErrorCodeInvalidComponentType)
}

func NewInvalidPathError() *Error {
	return NewError(ErrorCodeInvalidPath)
}

func NewInvalidSidecarError() *Error {
	return NewError(ErrorCodeInvalidSidecar)
}

func NewInvalidVersionError() *Error {
	return NewError(ErrorCodeInvalidVersion)
}

func NewLogsUnavailableError() *Error {
	return NewError(ErrorCodeInvalidComponentType)
}
