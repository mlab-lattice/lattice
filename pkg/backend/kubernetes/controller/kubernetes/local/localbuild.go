package local

import (

	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/informers/externalversions/lattice/v1"
	latticelisters "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"

	"github.com/golang/glog"
	"k8s.io/client-go/tools/cache"
	"k8s.io/apimachinery/pkg/util/runtime"
)

// Directory name probably needs to change

type Controller struct {
	//Contains the controller specific for updating DNS, Watches Address changes.

	latticeClient latticeclientset.Interface

	addressLister latticelisters.SystemLister
	//What is the purpose of this guy
	addressListerSynced cache.InformerSynced
}

func NewController(
	latticeClient latticeclientset.Interface,
	addressInformer latticeinformers.SystemInformer,
) *Controller {
	// Make a controller
	sbc := &Controller{
		latticeClient: latticeClient,
	}

	//Add event handlers.
	addressInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: sbc.addAddress,
		UpdateFunc: sbc.updateAddress,
	})
	sbc.addressLister = addressInformer.Lister()
	sbc.addressListerSynced = addressInformer.Informer().HasSynced

	return sbc
}

func (sbc *Controller) Run(workers int, stopCh <-chan struct{}) {

	defer runtime.HandleCrash()

	glog.Infof("Starting sys")

	<-stopCh
}

func (sbc *Controller) addAddress(obj interface{}) {
	// New address resource has arrived
	glog.V(1).Infof("MyController just got an add")
}

func (sbc *Controller) updateAddress(old, cur interface{}) {
	// Address object has been modified
	glog.V(1).Infof("MyController just got an update")
}

// Naive solution for now - on any update, completely rewrite the resolv.config we are using, and send SIGHUP
func (sbc *Controller) rewriteDNS() {
	// List from the informer given in the controller.
}