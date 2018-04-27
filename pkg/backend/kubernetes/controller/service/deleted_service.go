package service

import (
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func (c *Controller) syncDeletedService(service *latticev1.Service) error {
	return nil
}
