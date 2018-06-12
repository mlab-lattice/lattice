package service_node

import (
	// "encoding/json"
	"reflect"
	"sync"

	"github.com/golang/glog"

	envoycore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoycache "github.com/envoyproxy/go-control-plane/pkg/cache"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	xdsapi "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2"
)

// XXX: rename to ServiceNode

type ServiceNode struct {
	Id                 string
	latticeServiceName string

	EnvoyNode *envoycore.Node

	lock    sync.Mutex
	deleted bool

	clusters  []envoycache.Resource
	endpoints []envoycache.Resource
	routes    []envoycache.Resource
	listeners []envoycache.Resource
}

func NewServiceNode(id string, envoyNode *envoycore.Node) *ServiceNode {
	return &ServiceNode{
		Id:        id,
		EnvoyNode: envoyNode,
	}
}

func (s *ServiceNode) Domain() string {
	return s.EnvoyNode.GetId()
}

func (s *ServiceNode) Path() (tree.NodePath, error) {
	tnPath, err := tree.NodePathFromDomain(s.EnvoyNode.GetId())
	if err != nil {
		return "", err
	}
	return tnPath, nil
}

func (s *ServiceNode) ServiceCluster() string {
	return s.EnvoyNode.GetCluster()
}

func (s *ServiceNode) GetLatticeServiceName() string {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.latticeServiceName
}

func (s *ServiceNode) SetLatticeServiceName(latticeServiceName string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.latticeServiceName = latticeServiceName
}

func (s *ServiceNode) Update(backend xdsapi.Backend) error {
	glog.Info("ServiceNode.Update called")
	// disallow concurrent updates to service state
	s.lock.Lock()
	defer s.lock.Unlock()

	systemServices, err := backend.SystemServices(s.ServiceCluster())
	if err != nil {
		return err
	}

	clusters, err := s.getClusters(systemServices)
	if err != nil {
		return err
	}
	endpoints, err := s.getEndpoints(clusters, systemServices)
	if err != nil {
		return err
	}
	listeners, err := s.getListeners(systemServices)
	if err != nil {
		return err
	}
	routes, err := s.getRoutes(systemServices)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(clusters, s.clusters) ||
		!reflect.DeepEqual(endpoints, s.endpoints) ||
		!reflect.DeepEqual(listeners, s.listeners) ||
		!reflect.DeepEqual(routes, s.routes) {
		s.clusters = clusters
		s.endpoints = endpoints
		s.listeners = listeners
		s.routes = routes
		if !s.deleted {
			glog.V(4).Info("ServiceNode.Update updating XDS cache")
			err := backend.SetXDSCacheSnapshot(s.Id, s.endpoints, s.clusters, s.routes, s.listeners)
			if err != nil {
				return err
			}
		} else {
			// XXX: call clear on cache here?
			glog.Warning("ServiceNode.Update called on deleted node")
		}
	}

	return nil
}

func (s *ServiceNode) Cleanup(backend xdsapi.Backend) error {
	glog.V(4).Info("ServiceNode.Cleanup called")

	s.lock.Lock()
	defer s.lock.Unlock()

	backend.ClearXDSCacheSnapshot(s.Id)
	s.deleted = true

	return nil
}
