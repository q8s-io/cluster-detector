package event

import (
	"net/url"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	// "fmt"
	"github.com/q8s-io/cluster-detector/pkg/entity"

	corev1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	rbac "k8s.io/api/rbac/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/infrastructure/kubernetes"
)

const (
	// Number of object pointers.
	// Big enough so it won't be hit anytime soon with reasonable GetNewEvents frequency.
	LocalEventsBufferSize      = 100000
	DELETE                     = "Deleted"
	JobEvent                   = "Job"
	PodEvent                   = "Pod"
	DeploymentEvent            = "Deployment"
	ReplicationControllerEvent = "ReplicationController"
	NodeEvent                  = "Node"
	ReplicaSetEvent            = "ReplicaSet"
	DaemonSetEvent             = "DaemonSet"
	StatefulSetEvent           = "StatefulSet"
	NameSpaceEvent             = "NameSpace"
	ConfigMapEvent             = "ConfigMap"
	SecretEvent                = "Secret"
	ServiceEvent               = "Service"
	IngressEvent               = "Ingress"
	PersistentVolumeEvent      = "PersistentVolume"
	ServiceAccountEvent        = "ServiceAccount"
	ClusterRoleEvent           = "ClusterRole"
)

type EventClient struct {
	kubeClient  k8s.Interface
	stopChannel chan struct{}
}

var EventList chan *entity.EventInspection

func NewKubernetesSource(uri *url.URL) (*chan *entity.EventInspection, error) {
	EventList = make(chan *entity.EventInspection, LocalEventsBufferSize)

	kubeClient, err := kubernetes.GetKubernetesClient(uri)
	if err != nil {
		klog.Infof("Failed to create kubernetes client,because of %v", err)
		return nil, err
	}
	eventClient := EventClient{
		kubeClient:  kubeClient,
		stopChannel: make(chan struct{}),
	}
	go eventClient.watch()
	return &EventList, nil
}

func (harvester *EventClient) watch() {
	// Outer loop, for reconnections
	go harvester.normalWatch()
	go harvester.deleteWatch()
	go harvester.namespaceWatch()
	go harvester.comfigMapWatch()
	go harvester.secretEventWatch()
	go harvester.serviceEventWatch()
	go harvester.ingressEventWatch()
	go harvester.persistentVolumeEventWatch()
	go harvester.serviceAccountEvent()
	go harvester.clusterRoleEventWatch()
	select {}
}

func (harvester *EventClient) normalWatch() {
	for {
		// Do not write old events.
		events, err := harvester.kubeClient.CoreV1().Events(corev1.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			klog.Infof("Failed to load events: %v", err)
			continue
		}
		resourceVersion := events.ResourceVersion
		watcher, err := harvester.kubeClient.CoreV1().Events(corev1.NamespaceAll).Watch(
			metav1.ListOptions{
				Watch:           true,
				ResourceVersion: resourceVersion})
		if err != nil {
			klog.Infof("Failed to start watch for new events: %v", err)
			continue
		}
		watchChannel := watcher.ResultChan()
		// Inner loop, for update processing.
	innerLoop:
		for {
			select {
			case watchUpdate, ok := <-watchChannel:
				if !ok {
					klog.Info("Event watch channel closed")
					break innerLoop
				}
				if watchUpdate.Type == watch.Error {
					if status, ok := watchUpdate.Object.(*metav1.Status); ok {
						klog.Infof("Error during watch: %#v", status)
						break innerLoop
					}
					klog.Infof("Received unexpected error: %#v", watchUpdate.Object)
					break innerLoop
				}
				if event, ok := watchUpdate.Object.(*corev1.Event); ok {
					newEvent := &entity.EventInspection{
						EventKind:         event.InvolvedObject.Kind,
						EventNamespace:    event.InvolvedObject.Namespace,
						EventResourceName: event.InvolvedObject.Name,
						EventType:         string(watchUpdate.Type),
						EventInfo:         *event,
						EventTime:         time.Now(),
						//	EventUID: string(event.UID),
					}
					select {
					case EventList <- newEvent:
						// Ok, buffer not full.
					default:
						// Buffer full, need to drop the event.
						klog.Info("Event buffer full, dropping event")
					}
				} else {
					klog.Infof("Wrong object received: %v", watchUpdate)
				}
			case <-harvester.stopChannel:
				watcher.Stop()
				klog.Info("Event watching stopped")
				return
			}
		}
	}
}

func (harvester *EventClient) deleteWatch() {
	//var deleteinformer []cache.SharedIndexInformer
	stop := make(chan struct{})
	defer close(stop)
	factory := informers.NewSharedInformerFactory(harvester.kubeClient, time.Second*30)
	podInformer(factory)
	jobInformer(factory)
	replicaSetInformer(factory)
	replicationControllerInformer(factory)
	daemonSetInformer(factory)
	deploymentInformer(factory)
	nodeInformer(factory)
	statefulSetInformer(factory)
	go factory.Start(stop)
	<-stop
}

//TODO Pod
func podInformer(factory informers.SharedInformerFactory){
	podInformer := factory.Core().V1().Pods().Informer()
	registDeleteHandler(podInformer, PodEvent)
}

//TODO Job
func jobInformer(factory informers.SharedInformerFactory){
	jobInformer := factory.Batch().V1().Jobs().Informer()
	registDeleteHandler(jobInformer, JobEvent)
}

//TODO ReplicaSet
func replicaSetInformer(factory informers.SharedInformerFactory){
	rsInformer := factory.Apps().V1().ReplicaSets().Informer()
	registDeleteHandler(rsInformer, ReplicaSetEvent)
}

//TODO ReplicationController
func replicationControllerInformer(factory informers.SharedInformerFactory){
	rcInformer := factory.Core().V1().ReplicationControllers().Informer()
	registDeleteHandler(rcInformer, ReplicationControllerEvent)
}

//TODO DaemonSet
func daemonSetInformer(factory informers.SharedInformerFactory){
	dsInformer := factory.Apps().V1().DaemonSets().Informer()
	registDeleteHandler(dsInformer, DaemonSetEvent)
}

//TODO Deployment
func deploymentInformer(factory informers.SharedInformerFactory){
	dpInformer := factory.Apps().V1().Deployments().Informer()
	registDeleteHandler(dpInformer, DeploymentEvent)
}

//TODO Node
func nodeInformer(factory informers.SharedInformerFactory){
	nodeInformer := factory.Core().V1().Nodes().Informer()
	registDeleteHandler(nodeInformer, NodeEvent)
}

//TODO StatefulSet
func statefulSetInformer(factory informers.SharedInformerFactory){
	sfInformer := factory.Apps().V1().StatefulSets().Informer()
	registDeleteHandler(sfInformer, StatefulSetEvent)
}

func registDeleteHandler(informer cache.SharedIndexInformer, resourceType string) {
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				klog.Infof("can't get delete resource namespace and name\n")
			}
			nameSpace, name, err := cache.SplitMetaNamespaceKey(key)
			if err != nil {
				klog.Infof("can't split delete resource namespace and name\n")
			}

			inspection := &entity.EventInspection{
				EventKind:         resourceType,
				EventNamespace:    nameSpace,
				EventResourceName: name,
				EventType:         DELETE,
				EventTime:         time.Now(),
				EventInfo:         obj,
			}
			select {
			case EventList <- inspection:
				//OK not full
			default:
				klog.Info("Event buffer full, dropping event")
			}
		},
	})
}

func (harvester *EventClient) namespaceWatch() {
	for {
		// Do not write old events.
		events, err := harvester.kubeClient.CoreV1().Namespaces().List(metav1.ListOptions{})
		if err != nil {
			klog.Infof("Failed to load NameSpace events: %v", err)
			continue
		}
		resourceVersion := events.ResourceVersion
		watcher, err := harvester.kubeClient.CoreV1().Namespaces().Watch(
			metav1.ListOptions{
				Watch:           true,
				ResourceVersion: resourceVersion})
		if err != nil {
			klog.Infof("Failed to start watch for new NameSpace events: %v", err)
			continue
		}
		watchChannel := watcher.ResultChan()
		// Inner loop, for update processing.
	innerLoop:
		for {
			select {
			case watchUpdate, ok := <-watchChannel:
				if !ok {
					klog.Infof("Event watch channel closed")
					break innerLoop
				}
				if watchUpdate.Type == watch.Error {
					if status, ok := watchUpdate.Object.(*metav1.Status); ok {
						klog.Infof("Error during watch: %#v", status)
						break innerLoop
					}
					klog.Infof("Received unexpected error: %#v", watchUpdate.Object)
					break innerLoop
				}
				if event, ok := watchUpdate.Object.(*corev1.Namespace); ok {
					newEvent := &entity.EventInspection{
						EventKind:         NameSpaceEvent,
						EventNamespace:    event.Namespace,
						EventResourceName: event.Name,
						EventType:         string(watchUpdate.Type),
						EventInfo:         *event,
						EventTime:         time.Now(),
						//	EventUID: string(event.UID),
					}
					select {
					case EventList <- newEvent:
						// Ok, buffer not full.
					default:
						// Buffer full, need to drop the event.
						klog.Info("Event buffer full, dropping event")
					}
				} else {
					klog.Infof("Wrong object received: %v", watchUpdate)
				}
			case <-harvester.stopChannel:
				watcher.Stop()
				klog.Info("Event watching stopped")
				return
			}
		}
	}
}

func (harvester *EventClient) comfigMapWatch() {
	for {
		// Do not write old events.
		events, err := harvester.kubeClient.CoreV1().ConfigMaps(corev1.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			klog.Infof("Failed to load ComfigMap events: %v", err)
			continue
		}
		resourceVersion := events.ResourceVersion
		watcher, err := harvester.kubeClient.CoreV1().ConfigMaps(corev1.NamespaceAll).Watch(
			metav1.ListOptions{
				Watch:           true,
				ResourceVersion: resourceVersion})
		if err != nil {
			klog.Infof("Failed to start watch for new ConfigMap events: %v", err)
			continue
		}
		watchChannel := watcher.ResultChan()
		// Inner loop, for update processing.
	innerLoop:
		for {
			select {
			case watchUpdate, ok := <-watchChannel:
				if !ok {
					klog.Infof("Event watch channel closed")
					break innerLoop
				}
				if watchUpdate.Type == watch.Error {
					if status, ok := watchUpdate.Object.(*metav1.Status); ok {
						klog.Infof("Error during watch: %#v", status)
						break innerLoop
					}
					klog.Infof("Received unexpected error: %#v", watchUpdate.Object)
					break innerLoop
				}
				if event, ok := watchUpdate.Object.(*corev1.ConfigMap); ok {
					newEvent := &entity.EventInspection{
						EventKind:         ConfigMapEvent,
						EventNamespace:    event.Namespace,
						EventResourceName: event.Name,
						EventType:         string(watchUpdate.Type),
						EventInfo:         *event,
						EventTime:         time.Now(),
						//		EventUID: string(event.UID),
					}
					select {
					case EventList <- newEvent:
						// Ok, buffer not full.
					default:
						// Buffer full, need to drop the event.
						klog.Infof("Event buffer full, dropping event")
					}
				} else {
					klog.Infof("Wrong object received: %v", watchUpdate)
				}
			case <-harvester.stopChannel:
				watcher.Stop()
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}

func (harvester *EventClient) secretEventWatch() {
	for {
		// Do not write old events.
		events, err := harvester.kubeClient.CoreV1().Secrets(corev1.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to load Secret events: %v", err)
			continue
		}
		resourceVersion := events.ResourceVersion
		watcher, err := harvester.kubeClient.CoreV1().Secrets(corev1.NamespaceAll).Watch(
			metav1.ListOptions{
				Watch:           true,
				ResourceVersion: resourceVersion})
		if err != nil {
			klog.Errorf("Failed to start watch for new Secret events: %v", err)
			continue
		}
		watchChannel := watcher.ResultChan()
		// Inner loop, for update processing.
	innerLoop:
		for {
			select {
			case watchUpdate, ok := <-watchChannel:
				if !ok {
					klog.Errorf("Event watch channel closed")
					break innerLoop
				}
				if watchUpdate.Type == watch.Error {
					if status, ok := watchUpdate.Object.(*metav1.Status); ok {
						klog.Errorf("Error during watch: %#v", status)
						break innerLoop
					}
					klog.Errorf("Received unexpected error: %#v", watchUpdate.Object)
					break innerLoop
				}
				if event, ok := watchUpdate.Object.(*corev1.Secret); ok {
					newEvent := &entity.EventInspection{
						EventKind:         SecretEvent,
						EventNamespace:    event.Namespace,
						EventResourceName: event.Name,
						EventType:         string(watchUpdate.Type),
						EventInfo:         *event,
						EventTime:         time.Now(),
						//		EventUID: string(event.UID),
					}
					select {
					case EventList <- newEvent:
						// Ok, buffer not full.
					default:
						// Buffer full, need to drop the event.
						klog.Errorf("Event buffer full, dropping event")
					}
				} else {
					klog.Errorf("Wrong object received: %v", watchUpdate)
				}
			case <-harvester.stopChannel:
				watcher.Stop()
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}

func (harvester *EventClient) serviceEventWatch() {
	for {
		// Do not write old events.
		events, err := harvester.kubeClient.CoreV1().Services(corev1.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to load Service events: %v", err)
			continue
		}
		resourceVersion := events.ResourceVersion
		watcher, err := harvester.kubeClient.CoreV1().Services(corev1.NamespaceAll).Watch(
			metav1.ListOptions{
				Watch:           true,
				ResourceVersion: resourceVersion})
		if err != nil {
			klog.Errorf("Failed to start watch for new Service events: %v", err)
			continue
		}
		watchChannel := watcher.ResultChan()
		// Inner loop, for update processing.
	innerLoop:
		for {
			select {
			case watchUpdate, ok := <-watchChannel:
				if !ok {
					klog.Errorf("Event watch channel closed")
					break innerLoop
				}
				if watchUpdate.Type == watch.Error {
					if status, ok := watchUpdate.Object.(*metav1.Status); ok {
						klog.Errorf("Error during watch: %#v", status)
						break innerLoop
					}
					klog.Errorf("Received unexpected error: %#v", watchUpdate.Object)
					break innerLoop
				}
				if event, ok := watchUpdate.Object.(*corev1.Service); ok {
					newEvent := &entity.EventInspection{
						EventKind:         ServiceEvent,
						EventNamespace:    event.Namespace,
						EventResourceName: event.Name,
						EventType:         string(watchUpdate.Type),
						EventInfo:         *event,
						EventTime:         time.Now(),
						//			EventUID: string(event.UID),
					}
					select {
					case EventList <- newEvent:
						// Ok, buffer not full.
					default:
						// Buffer full, need to drop the event.
						klog.Errorf("Event buffer full, dropping event")
					}
				} else {
					klog.Errorf("Wrong object received: %v", watchUpdate)
				}
			case <-harvester.stopChannel:
				watcher.Stop()
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}

func (harvester *EventClient) ingressEventWatch() {
	for {
		// Do not write old events.
		events, err := harvester.kubeClient.ExtensionsV1beta1().Ingresses(corev1.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to load Ingress events: %v", err)
			continue
		}
		resourceVersion := events.ResourceVersion
		watcher, err := harvester.kubeClient.ExtensionsV1beta1().Ingresses(corev1.NamespaceAll).Watch(
			metav1.ListOptions{
				Watch:           true,
				ResourceVersion: resourceVersion})
		if err != nil {
			klog.Errorf("Failed to start watch for new Ingress events: %v", err)
			continue
		}
		watchChannel := watcher.ResultChan()
		// Inner loop, for update processing.
	innerLoop:
		for {
			select {
			case watchUpdate, ok := <-watchChannel:
				if !ok {
					klog.Errorf("Event watch channel closed")
					break innerLoop
				}
				if watchUpdate.Type == watch.Error {
					if status, ok := watchUpdate.Object.(*metav1.Status); ok {
						klog.Errorf("Error during watch: %#v", status)
						break innerLoop
					}
					klog.Errorf("Received unexpected error: %#v", watchUpdate.Object)
					break innerLoop
				}
				if event, ok := watchUpdate.Object.(*extensions.Ingress); ok {
					newEvent := &entity.EventInspection{
						EventKind:         IngressEvent,
						EventNamespace:    event.Namespace,
						EventResourceName: event.Name,
						EventType:         string(watchUpdate.Type),
						EventInfo:         *event,
						EventTime:         time.Now(),
						//			EventUID: string(event.UID),
					}
					select {
					case EventList <- newEvent:
						// Ok, buffer not full.
					default:
						// Buffer full, need to drop the event.
						klog.Errorf("Event buffer full, dropping event")
					}
				} else {
					klog.Errorf("Wrong object received: %v", watchUpdate)
				}
			case <-harvester.stopChannel:
				watcher.Stop()
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}

func (harvester *EventClient) persistentVolumeEventWatch() {
	for {
		// Do not write old events.
		events, err := harvester.kubeClient.CoreV1().PersistentVolumes().List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to load PersistentVolume events: %v", err)
			continue
		}
		resourceVersion := events.ResourceVersion
		watcher, err := harvester.kubeClient.CoreV1().PersistentVolumes().Watch(
			metav1.ListOptions{
				Watch:           true,
				ResourceVersion: resourceVersion})
		if err != nil {
			klog.Errorf("Failed to start watch for new PersistentVolume events: %v", err)
			continue
		}
		watchChannel := watcher.ResultChan()
		// Inner loop, for update processing.
	innerLoop:
		for {
			select {
			case watchUpdate, ok := <-watchChannel:
				if !ok {
					klog.Errorf("Event watch channel closed")
					break innerLoop
				}
				if watchUpdate.Type == watch.Error {
					if status, ok := watchUpdate.Object.(*metav1.Status); ok {
						klog.Errorf("Error during watch: %#v", status)
						break innerLoop
					}
					klog.Errorf("Received unexpected error: %#v", watchUpdate.Object)
					break innerLoop
				}
				if event, ok := watchUpdate.Object.(*corev1.PersistentVolume); ok {
					newEvent := &entity.EventInspection{
						EventKind:         PersistentVolumeEvent,
						EventNamespace:    event.Namespace,
						EventResourceName: event.Name,
						EventType:         string(watchUpdate.Type),
						EventInfo:         *event,
						EventTime:         time.Now(),
						//			EventUID: string(event.UID),
					}
					select {
					case EventList <- newEvent:
						// Ok, buffer not full.
					default:
						// Buffer full, need to drop the event.
						klog.Errorf("Event buffer full, dropping event")
					}
				} else {
					klog.Errorf("Wrong object received: %v", watchUpdate)
				}
			case <-harvester.stopChannel:
				watcher.Stop()
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}

func (harvester *EventClient) serviceAccountEvent() {
	for {
		// Do not write old events.
		events, err := harvester.kubeClient.CoreV1().ServiceAccounts(corev1.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to load ServiceAccount events: %v", err)
			continue
		}
		resourceVersion := events.ResourceVersion
		watcher, err := harvester.kubeClient.CoreV1().ServiceAccounts(corev1.NamespaceAll).Watch(
			metav1.ListOptions{
				Watch:           true,
				ResourceVersion: resourceVersion})
		if err != nil {
			klog.Errorf("Failed to start watch for new ServiceAccount events: %v", err)
			continue
		}
		watchChannel := watcher.ResultChan()
		// Inner loop, for update processing.
	innerLoop:
		for {
			select {
			case watchUpdate, ok := <-watchChannel:
				if !ok {
					klog.Errorf("Event watch channel closed")
					break innerLoop
				}
				if watchUpdate.Type == watch.Error {
					if status, ok := watchUpdate.Object.(*metav1.Status); ok {
						klog.Errorf("Error during watch: %#v", status)
						break innerLoop
					}
					klog.Errorf("Received unexpected error: %#v", watchUpdate.Object)
					break innerLoop
				}
				if event, ok := watchUpdate.Object.(*corev1.ServiceAccount); ok {
					newEvent := &entity.EventInspection{
						EventKind:         ServiceAccountEvent,
						EventNamespace:    event.Namespace,
						EventResourceName: event.Name,
						EventType:         string(watchUpdate.Type),
						EventInfo:         *event,
						EventTime:         time.Now(),
						//			EventUID: string(event.UID),
					}
					select {
					case EventList <- newEvent:
						// Ok, buffer not full.
					default:
						// Buffer full, need to drop the event.
						klog.Errorf("Event buffer full, dropping event")
					}
				} else {
					klog.Errorf("Wrong object received: %v", watchUpdate)
				}
			case <-harvester.stopChannel:
				watcher.Stop()
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}

func (harvester *EventClient) clusterRoleEventWatch() {
	for {
		// Do not write old events.
		events, err := harvester.kubeClient.RbacV1beta1().ClusterRoles().List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to load ClusterRole events: %v", err)
			continue
		}
		resourceVersion := events.ResourceVersion
		watcher, err := harvester.kubeClient.RbacV1beta1().ClusterRoles().Watch(
			metav1.ListOptions{
				Watch:           true,
				ResourceVersion: resourceVersion})
		if err != nil {
			klog.Errorf("Failed to start watch for new ClusterRole events: %v", err)
			continue
		}
		watchChannel := watcher.ResultChan()
		// Inner loop, for update processing.
	innerLoop:
		for {
			select {
			case watchUpdate, ok := <-watchChannel:
				if !ok {
					klog.Errorf("Event watch channel closed")
					break innerLoop
				}
				if watchUpdate.Type == watch.Error {
					if status, ok := watchUpdate.Object.(*metav1.Status); ok {
						klog.Errorf("Error during watch: %#v", status)
						break innerLoop
					}
					klog.Errorf("Received unexpected error: %#v", watchUpdate.Object)
					break innerLoop
				}
				if event, ok := watchUpdate.Object.(*rbac.ClusterRole); ok {
					newEvent := &entity.EventInspection{
						EventKind:         ClusterRoleEvent,
						EventNamespace:    event.Namespace,
						EventResourceName: event.Name,
						EventType:         string(watchUpdate.Type),
						EventInfo:         *event,
						EventTime:         time.Now(),
						//		EventUID: string(event.UID),
					}
					select {
					case EventList <- newEvent:
						// Ok, buffer not full.
					default:
						// Buffer full, need to drop the event.
						klog.Errorf("Event buffer full, dropping event")
					}
				} else {
					klog.Errorf("Wrong object received: %v", watchUpdate)
				}
			case <-harvester.stopChannel:
				watcher.Stop()
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}
