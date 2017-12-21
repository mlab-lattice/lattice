package util

import (
	"k8s.io/apimachinery/pkg/api/errors"
)

func IdempotentSeed(seedFunc func() (interface{}, error)) (interface{}, error) {
	result, err := seedFunc()
	if err != nil && !errors.IsAlreadyExists(err) {
		return nil, err
	}

	return result, nil
}
