package kube

import (
	"net/url"

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
	LocalEventsBufferSize = 100000
	NameSpaceEvent        = "NameSpace"
	ConfigMapEvent        = "ConfigMap"
	SecretEvent           = "Secret"
	ServiceEvent          = "Service"
	IngressEvent          = "Ingress"
	PersistentVolumeEvent = "PersistentVolume"
	ServiceAccountEvent   = "ServiceAccount"
	ClusterRoleEvent      = "ClusterRole"
)

type EventClient struct {
	kubeClient  k8s.Interface
	stopChannel chan struct{}
}

type LocalEventBuffer struct {
	eventKind         string
	eventNamespace    string
	eventResourceName string
	eventType         string
	eventInfo         interface{}
}

var EventList chan *LocalEventBuffer

func NewKubernetesSource(uri *url.URL) (*chan *LocalEventBuffer, error) {
	EventList = make(chan *LocalEventBuffer, LocalEventsBufferSize)

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
					newEvent := &LocalEventBuffer{
						eventKind:         event.InvolvedObject.Kind,
						eventNamespace:    event.InvolvedObject.Namespace,
						eventResourceName: event.InvolvedObject.Name,
						eventType:         string(watchUpdate.Type),
						eventInfo:         *event,
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
					newEvent := &LocalEventBuffer{
						eventKind:         NameSpaceEvent,
						eventNamespace:    event.Namespace,
						eventResourceName: event.Name,
						eventType:         string(watchUpdate.Type),
						eventInfo:         *event,
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
					newEvent := &LocalEventBuffer{
						eventKind:         ConfigMapEvent,
						eventNamespace:    event.Namespace,
						eventResourceName: event.Name,
						eventType:         string(watchUpdate.Type),
						eventInfo:         *event,
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
					newEvent := &LocalEventBuffer{
						eventKind:         SecretEvent,
						eventNamespace:    event.Namespace,
						eventResourceName: event.Name,
						eventType:         string(watchUpdate.Type),
						eventInfo:         *event,
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
					newEvent := &LocalEventBuffer{
						eventKind:         ServiceEvent,
						eventNamespace:    event.Namespace,
						eventResourceName: event.Name,
						eventType:         string(watchUpdate.Type),
						eventInfo:         *event,
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
					newEvent := &LocalEventBuffer{
						eventKind:         IngressEvent,
						eventNamespace:    event.Namespace,
						eventResourceName: event.Name,
						eventType:         string(watchUpdate.Type),
						eventInfo:         *event,
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
					newEvent := &LocalEventBuffer{
						eventKind:         PersistentVolumeEvent,
						eventNamespace:    event.Namespace,
						eventResourceName: event.Name,
						eventType:         string(watchUpdate.Type),
						eventInfo:         *event,
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
					newEvent := &LocalEventBuffer{
						eventKind:         ServiceAccountEvent,
						eventNamespace:    event.Namespace,
						eventResourceName: event.Name,
						eventType:         string(watchUpdate.Type),
						eventInfo:         *event,
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
					newEvent := &LocalEventBuffer{
						eventKind:         ClusterRoleEvent,
						eventNamespace:    event.Namespace,
						eventResourceName: event.Name,
						eventType:         string(watchUpdate.Type),
						eventInfo:         *event,
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
