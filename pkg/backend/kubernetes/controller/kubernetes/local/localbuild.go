package local

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"

	"github.com/golang/glog"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/apimachinery/pkg/util/runtime"
)

// Directory name probably needs to change

type Controller struct {
	//Contains the controller specific for updating DNS, Watches Address changes.
	syncAddressUpdate	func(bKey string) error
	enqueueAddressUpdate func(sysBuild *crv1.SystemBuild)

	latticeClient latticeclientset.Interface

	addressLister latticelisters.SystemLister
	addressListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewController(
	latticeClient latticeclientset.Interface,
	addressInformer latticeinformers.SystemInformer,
) *Controller {
	// Make a controller
	// TODO :: Rename sbc to addrc or something
	sbc := &Controller{
		latticeClient: latticeClient,
	}

	sbc.syncAddressUpdate = sbc.rewriteDNS
	sbc.enqueueAddressUpdate = sbc.enqueue

	//Add event handlers
	addressInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: sbc.addAddress,
		UpdateFunc: sbc.updateAddress,
		DeleteFunc: sbc.deleteAddress,
	})
	sbc.addressLister = addressInformer.Lister()
	sbc.addressListerSynced = addressInformer.Informer().HasSynced

	return sbc
}

func (sbc *Controller) enqueue(sysb *crv1.SystemBuild) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(sysb)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", sysb, err))
		return
	}

	sbc.queue.Add(key)
}

func (sbc *Controller) Run(workers int, stopCh <-chan struct{}) {

	defer runtime.HandleCrash()

	glog.Infof("Starting sys")

	<-stopCh
}

func (sbc *Controller) addAddress(obj interface{}) {
	// New address resource has arrived
	glog.V(1).Infof("MyController just got an add")
	address := obj.(*crv1.SystemBuild)

	sbc.enqueueAddressUpdate(address)
}

func (sbc *Controller) updateAddress(old, cur interface{}) {
	// Address object has been modified
	glog.V(1).Infof("MyController just got an update")
	address := cur.(*crv1.SystemBuild)

	sbc.enqueueAddressUpdate(address)
}

func (sbc *Controller) deleteAddress(obj interface{}) {
	// Address object has been modified
	glog.V(1).Infof("MyController just got a delete")
	address := obj.(*crv1.SystemBuild)

	sbc.enqueueAddressUpdate(address)
}

// Naive solution for now - on any update, completely rewrite the resolv.config we are using, and send SIGHUP
func (sbc *Controller) rewriteDNS(key string) error {
	// List from the informer given in the controller.
	glog.V(1).Infof("Called rewrite DNS")
	defer func() {
		glog.V(4).Infof("Finished rewrite DNS")
	}()

	// Work with the cache here.
	return nil
}