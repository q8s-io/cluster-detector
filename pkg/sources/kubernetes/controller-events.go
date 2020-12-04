package kubernetes
/*
Copyright 2016 Skippbox, Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*package kubernetes

import (
	"fmt"
	"github.com/bitnami-labs/kubewatch/pkg/utils"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"

	api_v1 "k8s.io/api/core/v1"
	ext_v1beta1 "k8s.io/api/extensions/v1beta1"
	rbac_v1beta1 "k8s.io/api/rbac/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"time"
)

const maxRetries = 5

var serverStartTime time.Time

const(
	ServiceEvent="service"
	NameSpaceEvent="nameSpace"
	ServiceAccountEvent="serviceAccount"
	PersistentVolumeEvent="pv"
	SecretEvent="secret"
	ConfigMapEvent="cm"
	ClusterRoleEvent="cr"
	IngressEvent="ingress"
	DefaultEvent="event"
	DELETE="delete"
	UPDATE="update"
	ADD="add"
)

// Event indicate the informerEvent
type Event struct {
	eventType    string
	events       string
	time time.Time
	resource interface{}
}

// Controller object
type Controller struct {
	logger       *logrus.Entry
	clientset    kubernetes.Interface
	queue        workqueue.RateLimitingInterface
	informer     cache.SharedIndexInformer
}

// Start prepares watchers and run their controllers, then waits for process termination signals
func (this *KubernetesEventSource)Start(nameSpace string) {
	var kubeClient kubernetes.Interface

	if _, err := rest.InClusterConfig(); err != nil {
		kubeClient = utils.GetClientOutOfCluster()
	} else {
		kubeClient = utils.GetClient()
	}

	// Adding Default Critical Alerts
	// For Capturing Critical Event NodeNotReady in Nodes
	/*nodeNotReadyInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				options.FieldSelector = "involvedObject.kind=Node,type=Normal,reason=NodeNotReady"
				return kubeClient.CoreV1().Events(nameSpace).List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				options.FieldSelector = "involvedObject.kind=Node,type=Normal,reason=NodeNotReady"
				return kubeClient.CoreV1().Events(nameSpace).Watch(options)
			},
		},
		&api_v1.Event{},
		0, //Skip resync
		cache.Indexers{},
	)

	nodeNotReadyController := newResourceController(kubeClient, nodeNotReadyInformer, "NodeNotReady")
	stopNodeNotReadyCh := make(chan struct{})
	defer close(stopNodeNotReadyCh)

	go nodeNotReadyController.Run(stopNodeNotReadyCh)

	// For Capturing Critical Event NodeReady in Nodes
	nodeReadyInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				options.FieldSelector = "involvedObject.kind=Node,type=Normal,reason=NodeReady"
				return kubeClient.CoreV1().Events(nameSpace).List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				options.FieldSelector = "involvedObject.kind=Node,type=Normal,reason=NodeReady"
				return kubeClient.CoreV1().Events(nameSpace).Watch(options)
			},
		},
		&api_v1.Event{},
		0, //Skip resync
		cache.Indexers{},
	)

	nodeReadyController := newResourceController(kubeClient,  nodeReadyInformer, "NodeReady")
	stopNodeReadyCh := make(chan struct{})
	defer close(stopNodeReadyCh)

	go nodeReadyController.Run(stopNodeReadyCh)

	// For Capturing Critical Event NodeRebooted in Nodes
	nodeRebootedInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				options.FieldSelector = "involvedObject.kind=Node,type=Warning,reason=Rebooted"
				return kubeClient.CoreV1().Events(nameSpace).List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				options.FieldSelector = "involvedObject.kind=Node,type=Warning,reason=Rebooted"
				return kubeClient.CoreV1().Events(nameSpace).Watch(options)
			},
		},
		&api_v1.Event{},
		0, //Skip resync
		cache.Indexers{},
	)

	nodeRebootedController := newResourceController(kubeClient, nodeRebootedInformer, "NodeRebooted")
	stopNodeRebootedCh := make(chan struct{})
	defer close(stopNodeRebootedCh)

	go nodeRebootedController.Run(stopNodeRebootedCh)

	// User Configured Events

		Podinformer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
					return kubeClient.CoreV1().Pods(nameSpace).List(options)
				},
				WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
					return kubeClient.CoreV1().Pods(nameSpace).Watch(options)
				},
			},
			&api_v1.Pod{},
			0, //Skip resync
			cache.Indexers{},
		)

		c1 := newResourceController(kubeClient,  Podinformer, "pod")
		stopCh := make(chan struct{})
		defer close(stopCh)

		go c1.Run(stopCh)

		// For Capturing CrashLoopBackOff Events in pods
		backoffInformer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
					options.FieldSelector = "involvedObject.kind=Pod,type=Warning,reason=BackOff"
					return kubeClient.CoreV1().Events(nameSpace).List(options)
				},
				WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
					options.FieldSelector = "involvedObject.kind=Pod,type=Warning,reason=BackOff"
					return kubeClient.CoreV1().Events(nameSpace).Watch(options)
				},
			},
			&api_v1.Event{},
			0, //Skip resync
			cache.Indexers{},
		)

		backoffcontroller := newResourceController(kubeClient, backoffInformer, "Backoff")
		stopBackoffCh := make(chan struct{})
		defer close(stopBackoffCh)

		go backoffcontroller.Run(stopBackoffCh)




		DSinformer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
					return kubeClient.AppsV1().DaemonSets(nameSpace).List(options)
				},
				WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
					return kubeClient.AppsV1().DaemonSets(nameSpace).Watch(options)
				},
			},
			&apps_v1.DaemonSet{},
			0, //Skip resync
			cache.Indexers{},
		)

		c2 := newResourceController(kubeClient, DSinformer, "daemon set")
		stopCh2 := make(chan struct{})
		defer close(stopCh2)

		go c2.Run(stopCh2)



		RSinformer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
					return kubeClient.AppsV1().ReplicaSets(nameSpace).List(options)
				},
				WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
					return kubeClient.AppsV1().ReplicaSets(nameSpace).Watch(options)
				},
			},
			&apps_v1.ReplicaSet{},
			0, //Skip resync
			cache.Indexers{},
		)

		c3 := newResourceController(kubeClient,  RSinformer, "replica set")
		stopCh3 := make(chan struct{})
		defer close(stopCh3)

		go c3.Run(stopCh3)
*/

		/*Serviceinformer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
					return kubeClient.CoreV1().Services(nameSpace).List(options)
				},
				WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
					return kubeClient.CoreV1().Services(nameSpace).Watch(options)
				},
			},
			&api_v1.Service{},
			0, //Skip resync
			cache.Indexers{},
		)
		this.newResourceController(kubeClient,Serviceinformer,ServiceEvent)
	/*	c4 := newResourceController(kubeClient, Serviceinformer, "service")
		stopCh4 := make(chan struct{})
		defer close(stopCh4)

		go c4.Run(stopCh4)*/



		/*Deployinformer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
					return kubeClient.AppsV1().Deployments(nameSpace).List(options)
				},
				WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
					return kubeClient.AppsV1().Deployments(nameSpace).Watch(options)
				},
			},
			&apps_v1.Deployment{},
			0, //Skip resync
			cache.Indexers{},
		)

		c5 := newResourceController(kubeClient, Deployinformer, "deployment")
		stopCh5 := make(chan struct{})
		defer close(stopCh5)

		go c5.Run(stopCh5)*/



	/*	Namespaceinformer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
					return kubeClient.CoreV1().Namespaces().List(options)
				},
				WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
					return kubeClient.CoreV1().Namespaces().Watch(options)
				},
			},
			&api_v1.Namespace{},
			0, //Skip resync
			cache.Indexers{},
		)
		this.newResourceController(kubeClient,Namespaceinformer,NameSpaceEvent)
	/*	c6 := newResourceController(kubeClient,Namespaceinformer, "namespace")
		stopCh6 := make(chan struct{})
		defer close(stopCh6)

		go c6.Run(stopCh6)
*/


		/*RCinformer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
					return kubeClient.CoreV1().ReplicationControllers(nameSpace).List(options)
				},
				WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
					return kubeClient.CoreV1().ReplicationControllers(nameSpace).Watch(options)
				},
			},
			&api_v1.ReplicationController{},
			0, //Skip resync
			cache.Indexers{},
		)

		c7 := newResourceController(kubeClient, RCinformer, "replication controller")
		stopCh7 := make(chan struct{})
		defer close(stopCh7)

		go c7.Run(stopCh7)
*/

		/*Jobinformer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
					return kubeClient.BatchV1().Jobs(nameSpace).List(options)
				},
				WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
					return kubeClient.BatchV1().Jobs(nameSpace).Watch(options)
				},
			},
			&batch_v1.Job{},
			0, //Skip resync
			cache.Indexers{},
		)

		c8 := newResourceController(kubeClient, Jobinformer, "job")
		stopCh8 := make(chan struct{})
		defer close(stopCh8)

		go c8.Run(stopCh8)
*/


		/*Nodeinformer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
					return kubeClient.CoreV1().Nodes().List(options)
				},
				WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
					return kubeClient.CoreV1().Nodes().Watch(options)
				},
			},
			&api_v1.Node{},
			0, //Skip resync
			cache.Indexers{},
		)

		c9 := newResourceController(kubeClient, Nodeinformer, "node")
		stopCh9 := make(chan struct{})
		defer close(stopCh9)

		go c9.Run(stopCh9)
*/


		/*SAinformer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
					return kubeClient.CoreV1().ServiceAccounts(nameSpace).List(options)
				},
				WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
					return kubeClient.CoreV1().ServiceAccounts(nameSpace).Watch(options)
				},
			},
			&api_v1.ServiceAccount{},
			0, //Skip resync
			cache.Indexers{},
		)
		this.newResourceController(kubeClient,SAinformer,ServiceAccountEvent)
	/*	c10 := newResourceController(kubeClient,  SAinformer, "service account")
		stopCh10 := make(chan struct{})
		defer close(stopCh10)

		go c10.Run(stopCh10)*/



		/*CRinformer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
					return kubeClient.RbacV1beta1().ClusterRoles().List(options)
				},
				WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
					return kubeClient.RbacV1beta1().ClusterRoles().Watch(options)
				},
			},
			&rbac_v1beta1.ClusterRole{},
			0, //Skip resync
			cache.Indexers{},
		)
		this.newResourceController(kubeClient,CRinformer,ClusterRoleEvent)
		/*c11 := newResourceController(kubeClient, CRinformer, "cluster role")
		stopCh11 := make(chan struct{})
		defer close(stopCh11)

		go c11.Run(stopCh11)*/



		/*PVinformer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
					return kubeClient.CoreV1().PersistentVolumes().List(options)
				},
				WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
					return kubeClient.CoreV1().PersistentVolumes().Watch(options)
				},
			},
			&api_v1.PersistentVolume{},
			0, //Skip resync
			cache.Indexers{},
		)
		this.newResourceController(kubeClient,PVinformer,PersistentVolumeEvent)
	/*	c12 := newResourceController(kubeClient,PVinformer, "persistent volume")
		stopCh12 := make(chan struct{})
		defer close(stopCh12)

		go c12.Run(stopCh12)*/



		/*Secretinformer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
					return kubeClient.CoreV1().Secrets(nameSpace).List(options)
				},
				WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
					return kubeClient.CoreV1().Secrets(nameSpace).Watch(options)
				},
			},
			&api_v1.Secret{},
			0, //Skip resync
			cache.Indexers{},
		)
		this.newResourceController(kubeClient,Secretinformer,SecretEvent)
		/*c13 := newResourceController(kubeClient,Secretinformer, "secret")
		stopCh13 := make(chan struct{})
		defer close(stopCh13)

		go c13.Run(stopCh13)*/



	/*	CMinformer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
					return kubeClient.CoreV1().ConfigMaps(nameSpace).List(options)
				},
				WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
					return kubeClient.CoreV1().ConfigMaps(nameSpace).Watch(options)
				},
			},
			&api_v1.ConfigMap{},
			0, //Skip resync
			cache.Indexers{},
		)
		this.newResourceController(kubeClient,CMinformer,ConfigMapEvent)
		/*c14 := newResourceController(kubeClient, CMinformer, "configmap")
		stopCh14 := make(chan struct{})
		defer close(stopCh14)

		go c14.Run(stopCh14)
*/


	/*	Ingressinformer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
					return kubeClient.ExtensionsV1beta1().Ingresses(nameSpace).List(options)
				},
				WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
					return kubeClient.ExtensionsV1beta1().Ingresses(nameSpace).Watch(options)
				},
			},
			&ext_v1beta1.Ingress{},
			0, //Skip resync
			cache.Indexers{},
		)
		this.newResourceController(kubeClient,Ingressinformer,IngressEvent)
		/*c15 := newResourceController(kubeClient,  Ingressinformer, "ingress")
		stopCh15 := make(chan struct{})
		defer close(stopCh15)

		go c15.Run(stopCh15)*/

	/*	Eventinformer := cache.NewSharedIndexInformer(
				&cache.ListWatch{
					ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
						return kubeClient.CoreV1().Events(nameSpace).List(options)
					},
					WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
						return kubeClient.CoreV1().Events(nameSpace).Watch(options)
					},
				},
				&api_v1.Event{},
				0,
				cache.Indexers{},
		)
		this.newResourceController(kubeClient,Eventinformer,DefaultEvent)


	/*	c16 := newResourceController(kubeClient,Eventinformer,"event")
		stopCh16:=make(chan struct{})
		defer close(stopCh16)*/
	//	go c16.Run(stopCh16)
	/*sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	signal.Notify(sigterm, syscall.SIGINT)
	<-sigterm
}

func (this *KubernetesEventSource)newResourceController(client kubernetes.Interface, informer cache.SharedIndexInformer, resourceType string){
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			addEvent:=getResource(resourceType,ADD,obj)
			//logrus.WithField("pkg", "kubewatch-"+resourceType).Infof("Processing add to %v: %s", resourceType, newEvent.key)
			fmt.Printf("%+v\n",addEvent)
		},
		UpdateFunc: func(old, new interface{}) {
			updateEvent:=getResource(resourceType,UPDATE,old)
			fmt.Printf("%+v\n",updateEvent)
		},
		DeleteFunc: func(obj interface{}) {
			deleteEvent:=getResource(resourceType,DELETE,obj)
			fmt.Printf("%+v\n",deleteEvent)
		},
	})
}

func getResource(eventTpye string,event string,obj interface{})(*Event){
	switch eventTpye {
	case NameSpaceEvent:
		nameSpaceEvent:=obj.(*api_v1.Namespace)
		return &Event{
			time: time.Now(),
			events: event,
			eventType: NameSpaceEvent,
			resource: nameSpaceEvent,
		}
	case ConfigMapEvent:
		comfigMapEvent:=obj.(*api_v1.ConfigMap)
		return &Event{
			time: time.Now(),
			events:event,
			eventType: ConfigMapEvent,
			resource: comfigMapEvent,
		}
	case SecretEvent:
		secretEvent:=obj.(*api_v1.Secret)
		return &Event{
			time: time.Now(),
			events: event,
			eventType: SecretEvent,
			resource: secretEvent,
		}
	case IngressEvent:
		ingressEvent:=obj.(*ext_v1beta1.Ingress)
		return &Event{
			time: time.Now(),
			events: event,
			eventType: SecretEvent,
			resource: ingressEvent,
		}
	case ServiceEvent:
		serviceEvent:=obj.(*api_v1.Service)
		return &Event{
			time:time.Now(),
			events: event,
			eventType: ServiceEvent,
			resource: serviceEvent,
		}
	case PersistentVolumeEvent:
		pvEvent:=obj.(*api_v1.PersistentVolume)
		return &Event{
			time:time.Now(),
			events: event,
			eventType: PersistentVolumeEvent,
			resource: pvEvent,
		}
	case ClusterRoleEvent:
		crEvent:=obj.(*rbac_v1beta1.ClusterRole)
		return &Event{
			time: time.Now(),
			events: event,
			eventType: ClusterRoleEvent,
			resource: crEvent,
		}
	case DefaultEvent:
		deafultEvent:=obj.(*api_v1.Event)
		return &Event{
			time: time.Now(),
			events: event,
			eventType: DefaultEvent,
			resource: deafultEvent,
		}
	default:
		logrus.Infof("UNKonwn events %s\n",eventTpye)
		return nil
	}
}

// Run starts the kubewatch controller
/*func (c *Controller) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Info("Starting kubewatch controller")
	serverStartTime = time.Now().Local()

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	c.logger.Info("Kubewatch controller synced and ready")

	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced is required for the cache.Controller interface.
func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}

// LastSyncResourceVersion is required for the cache.Controller interface.
func (c *Controller) LastSyncResourceVersion() string {
	return c.informer.LastSyncResourceVersion()
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
		// continue looping
	}
}

func (c *Controller) processNextItem() bool {
	newEvent, quit := c.queue.Get()

	if quit {
		return false
	}
	defer c.queue.Done(newEvent)
	err := c.processItem(newEvent.(Event))
	if err == nil {
		// No error, reset the ratelimit counters
		c.queue.Forget(newEvent)
	} else if c.queue.NumRequeues(newEvent) < maxRetries {
		c.logger.Errorf("Error processing %s (will retry): %v", newEvent.(Event).key, err)
		c.queue.AddRateLimited(newEvent)
	} else {
		// err != nil and too many retries
		c.logger.Errorf("Error processing %s (giving up): %v", newEvent.(Event).key, err)
		c.queue.Forget(newEvent)
		utilruntime.HandleError(err)
	}

	return true
}

/* TODOs
- Enhance event creation using client-side cacheing machanisms - pending
- Enhance the processItem to classify events - done
- Send alerts correspoding to events - done
*/
/*
func (c *Controller) processItem(newEvent Event) error {
	obj, _, err := c.informer.GetIndexer().GetByKey(newEvent.key)
	if err != nil {
		return fmt.Errorf("Error fetching object with key %s from store: %v", newEvent.key, err)
	}
	// get object's metedata
	objectMeta := utils.GetObjectMetaData(obj)

	// hold status type for default critical alerts
	var status string

	// namespace retrived from event key incase namespace value is empty
	if newEvent.namespace == "" && strings.Contains(newEvent.key, "/") {
		substring := strings.Split(newEvent.key, "/")
		newEvent.namespace = substring[0]
		newEvent.key = substring[1]
	}

	// process events based on its type
	switch newEvent.eventType {
	case "create":
		// compare CreationTimestamp and serverStartTime and alert only on latest events
		// Could be Replaced by using Delta or DeltaFIFO
		if objectMeta.CreationTimestamp.Sub(serverStartTime).Seconds() > 0 {
			switch newEvent.resourceType {
			case "NodeNotReady":
				status = "Danger"
			case "NodeReady":
				status = "Normal"
			case "NodeRebooted":
				status = "Danger"
			case "Backoff":
				status = "Danger"
			default:
				status = "Normal"
			}
			kbEvent := event.Event{
				Name:      objectMeta.Name,
				Namespace: newEvent.namespace,
				Kind:      newEvent.resourceType,
				Status:    status,
				Reason:    "Created",
			}
			fmt.Println("------------create-----------")
			fmt.Printf("%+v\n",kbEvent)
			event:=&api_v1.Event{
				TypeMeta: meta_v1.TypeMeta{
					Kind: kbEvent.Kind,
				},
				ObjectMeta:meta_v1.ObjectMeta{
					Name: kbEvent.Name,
					Namespace: kbEvent.Namespace,
				},
				Reason: kbEvent.Reason,
				Message: kbEvent.Status,
			}
			Events<-event
			return nil
		}
	case "update":
		/* TODOs
		- enahace update event processing in such a way that, it send alerts about what got changed.
		*/
	/*	switch newEvent.resourceType {
		case "Backoff":
			status = "Danger"
		default:
			status = "Warning"
		}
		kbEvent := event.Event{
			Name:      newEvent.key,
			Namespace: newEvent.namespace,
			Kind:      newEvent.resourceType,
			Status:    status,
			Reason:    "Updated",
		}
		fmt.Println("------------update-----------")
		fmt.Printf("%+v\n",kbEvent)
		event:=&api_v1.Event{
			TypeMeta: meta_v1.TypeMeta{
				Kind: kbEvent.Kind,
			},
			ObjectMeta:meta_v1.ObjectMeta{
				Name: kbEvent.Name,
				Namespace: kbEvent.Namespace,
			},
			Reason: kbEvent.Reason,
			Message: kbEvent.Status,
		}
		Events<-event
		return nil
	case "delete":
		kbEvent := event.Event{
			Name:      newEvent.key,
			Namespace: newEvent.namespace,
			Kind:      newEvent.resourceType,
			Status:    "Danger",
			Reason:    "Deleted",
		}
		fmt.Println("------------delete-----------")
		fmt.Printf("%+v\n",kbEvent)
		event:=&api_v1.Event{
			TypeMeta: meta_v1.TypeMeta{
				Kind: kbEvent.Kind,
			},
			ObjectMeta:meta_v1.ObjectMeta{
				Name: kbEvent.Name,
				Namespace: kbEvent.Namespace,
			},
			Reason: kbEvent.Reason,
			Message: kbEvent.Status,
		}
		Events<-event
		return nil
	}
	return nil
}*/