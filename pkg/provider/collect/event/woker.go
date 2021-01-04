package event

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	rbac "k8s.io/api/rbac/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/entity"
)

func (c *Client) normalWatch() {
	for {
		// Do not write old events.
		events, err := c.KubeClient.CoreV1().Events(corev1.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			klog.Infof("Failed to load events: %v", err)
			continue
		}
		resourceVersion := events.ResourceVersion
		watcher, err := c.KubeClient.CoreV1().Events(corev1.NamespaceAll).Watch(
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
					case List <- newEvent:
						// Ok, buffer not full.
					default:
						// Buffer full, need to drop the event.
						klog.Info("Event buffer full, dropping event")
					}
				} else {
					klog.Infof("Wrong object received: %v", watchUpdate)
				}
			case <-c.StopChannel:
				watcher.Stop()
				klog.Info("Event watching stopped")
				return
			}
		}
	}
}

func (c *Client) deleteWatch() {
	// var delete informer []cache.SharedIndexInformer
	stop := make(chan struct{})
	defer close(stop)
	factory := informers.NewSharedInformerFactory(c.KubeClient, time.Second*30)
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

func (c *Client) namespaceWatch() {
	for {
		// Do not write old events.
		events, err := c.KubeClient.CoreV1().Namespaces().List(metav1.ListOptions{})
		if err != nil {
			klog.Infof("Failed to load NameSpace events: %v", err)
			continue
		}
		resourceVersion := events.ResourceVersion
		watcher, err := c.KubeClient.CoreV1().Namespaces().Watch(
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
					case List <- newEvent:
						// Ok, buffer not full.
					default:
						// Buffer full, need to drop the event.
						klog.Info("Event buffer full, dropping event")
					}
				} else {
					klog.Infof("Wrong object received: %v", watchUpdate)
				}
			case <-c.StopChannel:
				watcher.Stop()
				klog.Info("Event watching stopped")
				return
			}
		}
	}
}

func (c *Client) configMapWatch() {
	for {
		// Do not write old events.
		events, err := c.KubeClient.CoreV1().ConfigMaps(corev1.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			klog.Infof("Failed to load ComfigMap events: %v", err)
			continue
		}
		resourceVersion := events.ResourceVersion
		watcher, err := c.KubeClient.CoreV1().ConfigMaps(corev1.NamespaceAll).Watch(
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
					case List <- newEvent:
						// Ok, buffer not full.
					default:
						// Buffer full, need to drop the event.
						klog.Infof("Event buffer full, dropping event")
					}
				} else {
					klog.Infof("Wrong object received: %v", watchUpdate)
				}
			case <-c.StopChannel:
				watcher.Stop()
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}

func (c *Client) secretEventWatch() {
	for {
		// Do not write old events.
		events, err := c.KubeClient.CoreV1().Secrets(corev1.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to load Secret events: %v", err)
			continue
		}
		resourceVersion := events.ResourceVersion
		watcher, err := c.KubeClient.CoreV1().Secrets(corev1.NamespaceAll).Watch(
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
					case List <- newEvent:
						// Ok, buffer not full.
					default:
						// Buffer full, need to drop the event.
						klog.Errorf("Event buffer full, dropping event")
					}
				} else {
					klog.Errorf("Wrong object received: %v", watchUpdate)
				}
			case <-c.StopChannel:
				watcher.Stop()
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}

func (c *Client) serviceEventWatch() {
	for {
		// Do not write old events.
		events, err := c.KubeClient.CoreV1().Services(corev1.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to load Service events: %v", err)
			continue
		}
		resourceVersion := events.ResourceVersion
		watcher, err := c.KubeClient.CoreV1().Services(corev1.NamespaceAll).Watch(
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
					case List <- newEvent:
						// Ok, buffer not full.
					default:
						// Buffer full, need to drop the event.
						klog.Errorf("Event buffer full, dropping event")
					}
				} else {
					klog.Errorf("Wrong object received: %v", watchUpdate)
				}
			case <-c.StopChannel:
				watcher.Stop()
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}

func (c *Client) ingressEventWatch() {
	for {
		// Do not write old events.
		events, err := c.KubeClient.ExtensionsV1beta1().Ingresses(corev1.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to load Ingress events: %v", err)
			continue
		}
		resourceVersion := events.ResourceVersion
		watcher, err := c.KubeClient.ExtensionsV1beta1().Ingresses(corev1.NamespaceAll).Watch(
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
					case List <- newEvent:
						// Ok, buffer not full.
					default:
						// Buffer full, need to drop the event.
						klog.Errorf("Event buffer full, dropping event")
					}
				} else {
					klog.Errorf("Wrong object received: %v", watchUpdate)
				}
			case <-c.StopChannel:
				watcher.Stop()
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}

func (c *Client) persistentVolumeEventWatch() {
	for {
		// Do not write old events.
		events, err := c.KubeClient.CoreV1().PersistentVolumes().List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to load PersistentVolume events: %v", err)
			continue
		}
		resourceVersion := events.ResourceVersion
		watcher, err := c.KubeClient.CoreV1().PersistentVolumes().Watch(
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
					case List <- newEvent:
						// Ok, buffer not full.
					default:
						// Buffer full, need to drop the event.
						klog.Errorf("Event buffer full, dropping event")
					}
				} else {
					klog.Errorf("Wrong object received: %v", watchUpdate)
				}
			case <-c.StopChannel:
				watcher.Stop()
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}

func (c *Client) serviceAccountEvent() {
	for {
		// Do not write old events.
		events, err := c.KubeClient.CoreV1().ServiceAccounts(corev1.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to load ServiceAccount events: %v", err)
			continue
		}
		resourceVersion := events.ResourceVersion
		watcher, err := c.KubeClient.CoreV1().ServiceAccounts(corev1.NamespaceAll).Watch(
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
					case List <- newEvent:
						// Ok, buffer not full.
					default:
						// Buffer full, need to drop the event.
						klog.Errorf("Event buffer full, dropping event")
					}
				} else {
					klog.Errorf("Wrong object received: %v", watchUpdate)
				}
			case <-c.StopChannel:
				watcher.Stop()
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}

func (c *Client) clusterRoleEventWatch() {
	for {
		// Do not write old events.
		events, err := c.KubeClient.RbacV1beta1().ClusterRoles().List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to load ClusterRole events: %v", err)
			continue
		}
		resourceVersion := events.ResourceVersion
		watcher, err := c.KubeClient.RbacV1beta1().ClusterRoles().Watch(
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
					case List <- newEvent:
						// Ok, buffer not full.
					default:
						// Buffer full, need to drop the event.
						klog.Errorf("Event buffer full, dropping event")
					}
				} else {
					klog.Errorf("Wrong object received: %v", watchUpdate)
				}
			case <-c.StopChannel:
				watcher.Stop()
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}

// TODO Pod
func podInformer(factory informers.SharedInformerFactory) {
	podInformer := factory.Core().V1().Pods().Informer()
	registerDeleteHandler(podInformer, PodEvent)
}

// TODO Job
func jobInformer(factory informers.SharedInformerFactory) {
	jobInformer := factory.Batch().V1().Jobs().Informer()
	registerDeleteHandler(jobInformer, JobEvent)
}

// TODO ReplicaSet
func replicaSetInformer(factory informers.SharedInformerFactory) {
	rsInformer := factory.Apps().V1().ReplicaSets().Informer()
	registerDeleteHandler(rsInformer, ReplicaSetEvent)
}

// TODO ReplicationController
func replicationControllerInformer(factory informers.SharedInformerFactory) {
	rcInformer := factory.Core().V1().ReplicationControllers().Informer()
	registerDeleteHandler(rcInformer, ReplicationControllerEvent)
}

// TODO DaemonSet
func daemonSetInformer(factory informers.SharedInformerFactory) {
	dsInformer := factory.Apps().V1().DaemonSets().Informer()
	registerDeleteHandler(dsInformer, DaemonSetEvent)
}

// TODO Deployment
func deploymentInformer(factory informers.SharedInformerFactory) {
	dpInformer := factory.Apps().V1().Deployments().Informer()
	registerDeleteHandler(dpInformer, DeploymentEvent)
}

// TODO Node
func nodeInformer(factory informers.SharedInformerFactory) {
	nodeInformer := factory.Core().V1().Nodes().Informer()
	registerDeleteHandler(nodeInformer, NodeEvent)
}

// TODO StatefulSet
func statefulSetInformer(factory informers.SharedInformerFactory) {
	sfInformer := factory.Apps().V1().StatefulSets().Informer()
	registerDeleteHandler(sfInformer, StatefulSetEvent)
}

func registerDeleteHandler(informer cache.SharedIndexInformer, resourceType string) {
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				klog.Infof("can't get delete resource namespace and name\n")
				return
			}
			nameSpace, name, err := cache.SplitMetaNamespaceKey(key)
			if err != nil {
				klog.Infof("can't split delete resource namespace and name\n")
				return
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
			case List <- inspection:
				// OK not full
			default:
				klog.Info("Event buffer full, dropping event")
			}
		},
	})
}
