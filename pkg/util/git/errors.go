package git

import (
	"fmt"
)

// CloneError is returned when there is an error cloning a repository.
type CloneError struct {
	err error
}

func (e *CloneError) Error() string {
	return fmt.Sprintf("error cloning repository: %v", e.err)
}

func newCloneError(err error) error {
	if err == nil {
		return nil
	}

	return &CloneError{
		err: err,
	}
}

// FetchError is returned when there is an error fetching.
type FetchError struct {
	err error
}

func (e *FetchError) Error() string {
	return fmt.Sprintf("error fetching: %v", e.err)
}

func newFetchError(err error) error {
	if err == nil {
		return nil
	}

	return &FetchError{
		err: err,
	}
}
