package ads

import (
	"encoding/json"
	"reflect"
	"sync"

	"github.com/golang/glog"

	envoycore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoycache "github.com/envoyproxy/go-control-plane/pkg/cache"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	xdsapi "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh/envoy/xdsapi/v2"
)

// XXX: rename to ServiceNode

type Service struct {
	Id string

	EnvoyNode *envoycore.Node

	lock sync.Mutex

	clusters  []envoycache.Resource
	endpoints []envoycache.Resource
	routes    []envoycache.Resource
	listeners []envoycache.Resource
}

func NewService(id string, envoyNode *envoycore.Node) *Service {
	return &Service{
		Id:        id,
		EnvoyNode: envoyNode,
	}
}

func (s *Service) Path() (tree.NodePath, error) {
	tnPath, err := tree.NodePathFromDomain(s.EnvoyNode.GetId())
	if err != nil {
		return "", err
	}
	return tnPath, nil
}

func (s *Service) Namespace() string {
	return s.EnvoyNode.GetCluster()
}

func (s *Service) Update(backend xdsapi.Backend) error {
	glog.Info("Service.update called")
	// disallow concurrent updates to service state
	s.lock.Lock()
	defer s.lock.Unlock()

	systemServices, err := backend.SystemServices(s.Namespace())
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
		clustersJson, _ := json.MarshalIndent(s.clusters, "", "  ")
		endpointsJson, _ := json.MarshalIndent(s.endpoints, "", "  ")
		listenersJson, _ := json.MarshalIndent(s.listeners, "", "  ")
		routesJson, _ := json.MarshalIndent(s.routes, "", "  ")
		glog.Infof("Setting new snapshot for %v\nclusters\n%v\nendpoints\n%v\nlisteners\n%v\nroutes\n%v",
			s.Id, string(clustersJson), string(endpointsJson), string(listenersJson), string(routesJson))
		err := backend.SetXDSCacheSnapshot(s.Id, s.endpoints, s.clusters, s.routes, s.listeners)
		if err != nil {
			return err
		}
	}

	return nil
}
