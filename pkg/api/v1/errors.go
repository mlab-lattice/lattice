package v1

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type ErrorCode string

const (
	ErrorCodeUnknown              ErrorCode = "UNKNOWN"
	ErrorCodeInvalidSystemOptions ErrorCode = "INVALID_SYSTEM_OPTIONS"
	ErrorCodeSystemAlreadyExists  ErrorCode = "SYSTEM_ALREADY_EXISTS"
	ErrorCodeInvalidSystemID      ErrorCode = "INVALID_SYSTEM_ID"
	ErrorCodeInvalidSystemVersion ErrorCode = "INVALID_SYSTEM_VERSION"
	ErrorCodeInvalidBuildID       ErrorCode = "INVALID_BUILD_ID"
	ErrorCodeInvalidDeployID      ErrorCode = "INVALID_DEPLOY_ID"
	ErrorCodeInvalidTeardownID    ErrorCode = "INVALID_TEARDOWN_ID"
	ErrorCodeInvalidServicePath   ErrorCode = "INVALID_SERVICE_PATH"
	ErrorCodeInvalidSidecar       ErrorCode = "INVALID_SIDECAR"
	ErrorCodeInvalidSystemSecret  ErrorCode = "INVALID_SYSTEM_SECRET"
	ErrorCodeConflict             ErrorCode = "CONFLICT"
)

type Error interface {
	error
	Code() ErrorCode
}

func NewUnknownError() *UnknownError {
	return &UnknownError{}
}

type UnknownError struct{}

func (e *UnknownError) Error() string {
	return fmt.Sprintf("unknown error")
}

func (e *UnknownError) Code() ErrorCode {
	return ErrorCodeUnknown
}

type InvalidSystemOptionsError struct {
	Reason string `json:"reason"`
}

func (e *InvalidSystemOptionsError) Error() string {
	return fmt.Sprintf("invalid system: %v", e.Reason)
}

func (e *InvalidSystemOptionsError) Code() ErrorCode {
	return ErrorCodeInvalidSystemOptions
}

func NewSystemAlreadyExistsError(id SystemID) *SystemAlreadyExistsError {
	return &SystemAlreadyExistsError{
		ID: id,
	}
}

type SystemAlreadyExistsError struct {
	ID SystemID `json:"id"`
}

func (e *SystemAlreadyExistsError) Error() string {
	return fmt.Sprintf("system %v already exists", e.ID)
}

func (e *SystemAlreadyExistsError) Code() ErrorCode {
	return ErrorCodeSystemAlreadyExists
}

func NewInvalidSystemIDError(id SystemID) *InvalidSystemIDError {
	return &InvalidSystemIDError{
		ID: id,
	}
}

type InvalidSystemIDError struct {
	ID SystemID `json:"id"`
}

func (e *InvalidSystemIDError) Error() string {
	return fmt.Sprintf("invalid system %v", e.ID)
}

func (e *InvalidSystemIDError) Code() ErrorCode {
	return ErrorCodeInvalidSystemID
}

func NewSystemNotCreatedError(id SystemID, state SystemState) *SystemNotCreatedError {
	return &SystemNotCreatedError{
		ID:    id,
		State: state,
	}
}

type SystemNotCreatedError struct {
	ID    SystemID    `json:"id"`
	State SystemState `json:"state"`
}

func (e *SystemNotCreatedError) Error() string {
	return fmt.Sprintf("system %v is in state %v", e.ID, e.State)
}

func (e *SystemNotCreatedError) Code() ErrorCode {
	return ErrorCodeInvalidSystemID
}

type InvalidSystemVersionError struct {
	Version string `json:"version"`
}

func (e *InvalidSystemVersionError) Code() ErrorCode {
	return ErrorCodeInvalidSystemVersion
}

func (e *InvalidSystemVersionError) Error() string {
	return fmt.Sprintf("invalid system version %v", e.Version)
}

func NewInvalidBuildIDError(id BuildID) *InvalidBuildIDError {
	return &InvalidBuildIDError{
		ID: id,
	}
}

type InvalidBuildIDError struct {
	ID BuildID `json:"id"`
}

func (e *InvalidBuildIDError) Error() string {
	return fmt.Sprintf("invalid build %v", e.ID)
}

func (e *InvalidBuildIDError) Code() ErrorCode {
	return ErrorCodeInvalidBuildID
}

func NewInvalidDeployIDError(id DeployID) *InvalidDeployIDError {
	return &InvalidDeployIDError{
		ID: id,
	}
}

type InvalidDeployIDError struct {
	ID DeployID `json:"id"`
}

func (e *InvalidDeployIDError) Error() string {
	return fmt.Sprintf("invalid rollout %v", e.ID)
}

func (e *InvalidDeployIDError) Code() ErrorCode {
	return ErrorCodeInvalidDeployID
}

func NewInvalidTeardownIDError(id TeardownID) *InvalidTeardownIDError {
	return &InvalidTeardownIDError{
		ID: id,
	}
}

type InvalidTeardownIDError struct {
	ID TeardownID `json:"id"`
}

func (e *InvalidTeardownIDError) Error() string {
	return fmt.Sprintf("invalid teardown %v", e.ID)
}

func (e *InvalidTeardownIDError) Code() ErrorCode {
	return ErrorCodeInvalidTeardownID
}

func NewInvalidServicePathError(path tree.NodePath) *InvalidServicePathError {
	return &InvalidServicePathError{
		Path: path,
	}
}

type InvalidServicePathError struct {
	Path tree.NodePath `json:"path"`
}

func (e *InvalidServicePathError) Error() string {
	return fmt.Sprintf("invalid service %v", e.Path)
}

func (e *InvalidServicePathError) Code() ErrorCode {
	return ErrorCodeInvalidServicePath
}

func NewInvalidSidecarError(sidecar string) *InvalidSidecarError {
	return &InvalidSidecarError{
		Sidecar: sidecar,
	}
}

type InvalidSidecarError struct {
	Sidecar string `json:"sidecar"`
}

func (e *InvalidSidecarError) Error() string {
	return fmt.Sprintf("invalid component %v", e.Sidecar)
}

func (e *InvalidSidecarError) Code() ErrorCode {
	return ErrorCodeInvalidSidecar
}

func NewInvalidSystemSecretError(path tree.NodePath, name string) *InvalidSystemSecretError {
	return &InvalidSystemSecretError{
		Path: path,
		Name: name,
	}
}

type InvalidSystemSecretError struct {
	Path tree.NodePath `json:"path"`
	Name string        `json:"name"`
}

func (e *InvalidSystemSecretError) Error() string {
	return fmt.Sprintf("invalid secret %v:%v", e.Path, e.Name)
}

func (e *InvalidSystemSecretError) Code() ErrorCode {
	return ErrorCodeInvalidSystemSecret
}

func NewConflictError(reason string) *ConflictError {
	return &ConflictError{
		Reason: reason,
	}
}

type ConflictError struct {
	Reason string `json:"reason"`
}

func (e *ConflictError) Error() string {
	msg := "conflict"
	if e.Reason != "" {
		msg += fmt.Sprintf(": %v", e.Reason)
	}
	return msg
}

func (e *ConflictError) Code() ErrorCode {
	return ErrorCodeConflict
}
