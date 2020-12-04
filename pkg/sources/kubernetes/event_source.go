// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kubernetes

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/q8s-io/cluster-detector/pkg/common/kubernetes"
	"github.com/q8s-io/cluster-detector/pkg/core"
	kubeapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubewatch "k8s.io/apimachinery/pkg/watch"
	rbac_v1beta1 "k8s.io/api/rbac/v1beta1"
	ext_v1beta1 "k8s.io/api/extensions/v1beta1"
	//kubev1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog"
	"net/url"
	"time"
	k8s "k8s.io/client-go/kubernetes"
)

const (
	// Number of object pointers. Big enough so it won't be hit anytime soon with reasonable GetNewEvents frequency.
	LocalEventsBufferSize = 100000
	ServiceEvent="Service"
	NameSpaceEvent="NameSpace"
	ServiceAccountEvent="ServiceAccount"
	PersistentVolumeEvent="PersistentVolume"
	SecretEvent="Secret"
	ConfigMapEvent="ConfigMap"
	ClusterRoleEvent="ClusterRole"
	IngressEvent="Ingress"
	DefaultEvent="DefaultEvent"
	DELETE="delete"
	ADD="add"
	UPDATE="update"
)

var (
	// Last time of event since unix epoch in seconds
	lastEventTimestamp = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "eventer",
			Subsystem: "scraper",
			Name:      "last_time_seconds",
			Help:      "Last time of event since unix epoch in seconds.",
		})
	totalEventsNum = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "eventer",
			Subsystem: "scraper",
			Name:      "events_total_number",
			Help:      "The total number of events.",
		})
	scrapEventsDuration = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: "eventer",
			Subsystem: "scraper",
			Name:      "duration_milliseconds",
			Help:      "Time spent scraping events in milliseconds.",
		})
)

func init() {
	prometheus.MustRegister(lastEventTimestamp)
	prometheus.MustRegister(totalEventsNum)
	prometheus.MustRegister(scrapEventsDuration)
}

type LocalEventBuffer struct {
	events  interface{}
	eventType string
	time time.Time
	eventSource string
}

// Implements core.EventSource interface.
type KubernetesEventSource struct {
	// Large local buffer, periodically read.
	localEventsBuffer chan *LocalEventBuffer
	stopChannel chan struct{}
	eventClient k8s.Interface
}

type Result struct {
	Timestamp time.Time
	Events []interface{}
}

func (this *KubernetesEventSource) GetNewEvents() *core.EventBatch {
	startTime := time.Now()
	defer func() {
		lastEventTimestamp.Set(float64(time.Now().Unix()))
		scrapEventsDuration.Observe(float64(time.Since(startTime)) / float64(time.Millisecond))
	}()
	result := &core.EventBatch{
		Timestamp: time.Now(),
		Events:    []interface{}{},
	}
	// Get all data from the buffer.
EventLoop:
	for {
		select {
		case event := <-this.localEventsBuffer:
			fmt.Println("-------------获取事件-----------")
			fmt.Printf("%+v\n",*event)
			fmt.Println("----------------------")
			result.Events = append(result.Events, *event)
		default:
			break EventLoop
		}
	}

	totalEventsNum.Add(float64(len(result.Events)))

	return result
}

func (this *KubernetesEventSource)eventWatch(){
	for {
		events, err := this.eventClient.CoreV1().Events(kubeapi.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to load events: %v", err)
			time.Sleep(time.Second)
			continue
		}

		// Do not write old events.

		resourceVersion := events.ResourceVersion

		watcher, err := this.eventClient.CoreV1().Events(kubeapi.NamespaceAll).Watch(
			metav1.ListOptions{
				Watch:           true,
				ResourceVersion: resourceVersion})
		if err != nil {
			klog.Errorf("Failed to start watch for new events: %v", err)
			time.Sleep(time.Second)
			continue
		}

		watchChannel := watcher.ResultChan()
		// Inner loop, for update processing.
	inner_loop:
		for {
			select {
			case watchUpdate, ok := <-watchChannel:
				if !ok {
					klog.Errorf("Event watch channel closed")
					break inner_loop
				}

				if watchUpdate.Type == kubewatch.Error {
					if status, ok := watchUpdate.Object.(*metav1.Status); ok {
						klog.Errorf("Error during watch: %#v", status)
						break inner_loop
					}
					klog.Errorf("Received unexpected error: %#v", watchUpdate.Object)
					break inner_loop
				}

				if event, ok := watchUpdate.Object.(*kubeapi.Event); ok {
					//fmt.Println(event.InvolvedObject.Kind," ",event.InvolvedObject.Name," ",event.InvolvedObject.Namespace," ",event.Name," ",event.Source," ",event.Namespace," ",event.Reason," ",event.Kind)
					switch watchUpdate.Type {
					case kubewatch.Added:
						defaultEvents:=&LocalEventBuffer{
						events: *event,
						eventType: ADD,
						time: time.Now(),
						eventSource: DefaultEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
					case kubewatch.Modified:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: DELETE,
							time: time.Now(),
							eventSource: DefaultEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
					case kubewatch.Deleted:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: DELETE,
							time: time.Now(),
							eventSource: DefaultEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
						// Deleted events are silently ignored.
					default:
						klog.Warningf("Unknown watchUpdate.Type: %#v", watchUpdate.Type)
					}
				} else {
					klog.Errorf("Wrong object received: %v", watchUpdate)
				}

			case <-this.stopChannel:
				watcher.Stop()
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}

func (this *KubernetesEventSource)namespaceWatch(){
	for {
		events, err := this.eventClient.CoreV1().Namespaces().List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to load NameSpace events: %v", err)
			time.Sleep(time.Second)
			continue
		}

		// Do not write old events.

		resourceVersion := events.ResourceVersion

		watcher, err := this.eventClient.CoreV1().Namespaces().Watch(
			metav1.ListOptions{
				Watch:           true,
				ResourceVersion: resourceVersion})
		if err != nil {
			klog.Errorf("Failed to start watch for new NameSpace events: %v", err)
			time.Sleep(time.Second)
			continue
		}

		watchChannel := watcher.ResultChan()
		// Inner loop, for update processing.
	inner_loop:
		for {
			select {
			case watchUpdate, ok := <-watchChannel:
				if !ok {
					klog.Errorf("Event watch channel closed")
					break inner_loop
				}

				if watchUpdate.Type == kubewatch.Error {
					if status, ok := watchUpdate.Object.(*metav1.Status); ok {
						klog.Errorf("Error during watch: %#v", status)
						break inner_loop
					}
					klog.Errorf("Received unexpected error: %#v", watchUpdate.Object)
					break inner_loop
				}

				if event, ok := watchUpdate.Object.(*kubeapi.Namespace); ok {
					//fmt.Println(event.InvolvedObject.Kind," ",event.InvolvedObject.Name," ",event.InvolvedObject.Namespace," ",event.Name," ",event.Source," ",event.Namespace," ",event.Reason," ",event.Kind)
					switch watchUpdate.Type {
					case kubewatch.Added:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: ADD,
							time: time.Now(),
							eventSource: NameSpaceEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
					case kubewatch.Modified:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: DELETE,
							time: time.Now(),
							eventSource: NameSpaceEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
					case kubewatch.Deleted:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: DELETE,
							time: time.Now(),
							eventSource: NameSpaceEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
						// Deleted events are silently ignored.
					default:
						klog.Warningf("Unknown watchUpdate.Type: %#v", watchUpdate.Type)
					}
				} else {
					klog.Errorf("Wrong object received: %v", watchUpdate)
				}

			case <-this.stopChannel:
				watcher.Stop()
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}

func (this *KubernetesEventSource)comfigMapWatch(){
	for {
		events, err := this.eventClient.CoreV1().ConfigMaps(kubeapi.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to load ComfigMap events: %v", err)
			time.Sleep(time.Second)
			continue
		}

		// Do not write old events.

		resourceVersion := events.ResourceVersion

		watcher, err := this.eventClient.CoreV1().ConfigMaps(kubeapi.NamespaceAll).Watch(
			metav1.ListOptions{
				Watch:           true,
				ResourceVersion: resourceVersion})
		if err != nil {
			klog.Errorf("Failed to start watch for new ConfigMap events: %v", err)
			time.Sleep(time.Second)
			continue
		}

		watchChannel := watcher.ResultChan()
		// Inner loop, for update processing.
	inner_loop:
		for {
			select {
			case watchUpdate, ok := <-watchChannel:
				if !ok {
					klog.Errorf("Event watch channel closed")
					break inner_loop
				}

				if watchUpdate.Type == kubewatch.Error {
					if status, ok := watchUpdate.Object.(*metav1.Status); ok {
						klog.Errorf("Error during watch: %#v", status)
						break inner_loop
					}
					klog.Errorf("Received unexpected error: %#v", watchUpdate.Object)
					break inner_loop
				}

				if event, ok := watchUpdate.Object.(*kubeapi.ConfigMap); ok {
					//fmt.Println(event.InvolvedObject.Kind," ",event.InvolvedObject.Name," ",event.InvolvedObject.Namespace," ",event.Name," ",event.Source," ",event.Namespace," ",event.Reason," ",event.Kind)
					switch watchUpdate.Type {
					case kubewatch.Added:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: ADD,
							time: time.Now(),
							eventSource: ConfigMapEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
					case kubewatch.Modified:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: DELETE,
							time: time.Now(),
							eventSource: ConfigMapEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
					case kubewatch.Deleted:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: DELETE,
							time: time.Now(),
							eventSource: ConfigMapEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
						// Deleted events are silently ignored.
					default:
						klog.Warningf("Unknown watchUpdate.Type: %#v", watchUpdate.Type)
					}
				} else {
					klog.Errorf("Wrong object received: %v", watchUpdate)
				}

			case <-this.stopChannel:
				watcher.Stop()
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}

func (this *KubernetesEventSource)secretEventWatch(){
	for {
		events, err := this.eventClient.CoreV1().Secrets(kubeapi.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to load Secret events: %v", err)
			time.Sleep(time.Second)
			continue
		}

		// Do not write old events.

		resourceVersion := events.ResourceVersion

		watcher, err := this.eventClient.CoreV1().Secrets(kubeapi.NamespaceAll).Watch(
			metav1.ListOptions{
				Watch:           true,
				ResourceVersion: resourceVersion})
		if err != nil {
			klog.Errorf("Failed to start watch for new Secret events: %v", err)
			time.Sleep(time.Second)
			continue
		}

		watchChannel := watcher.ResultChan()
		// Inner loop, for update processing.
	inner_loop:
		for {
			select {
			case watchUpdate, ok := <-watchChannel:
				if !ok {
					klog.Errorf("Event watch channel closed")
					break inner_loop
				}

				if watchUpdate.Type == kubewatch.Error {
					if status, ok := watchUpdate.Object.(*metav1.Status); ok {
						klog.Errorf("Error during watch: %#v", status)
						break inner_loop
					}
					klog.Errorf("Received unexpected error: %#v", watchUpdate.Object)
					break inner_loop
				}

				if event, ok := watchUpdate.Object.(*kubeapi.Secret); ok {
					//fmt.Println(event.InvolvedObject.Kind," ",event.InvolvedObject.Name," ",event.InvolvedObject.Namespace," ",event.Name," ",event.Source," ",event.Namespace," ",event.Reason," ",event.Kind)
					switch watchUpdate.Type {
					case kubewatch.Added:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: ADD,
							time: time.Now(),
							eventSource: SecretEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
					case kubewatch.Modified:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: DELETE,
							time: time.Now(),
							eventSource: SecretEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
					case kubewatch.Deleted:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: DELETE,
							time: time.Now(),
							eventSource: SecretEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
						// Deleted events are silently ignored.
					default:
						klog.Warningf("Unknown watchUpdate.Type: %#v", watchUpdate.Type)
					}
				} else {
					klog.Errorf("Wrong object received: %v", watchUpdate)
				}

			case <-this.stopChannel:
				watcher.Stop()
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}

func (this *KubernetesEventSource)serviceEventWatch(){
	for {
		events, err := this.eventClient.CoreV1().Services(kubeapi.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to load Service events: %v", err)
			time.Sleep(time.Second)
			continue
		}

		// Do not write old events.

		resourceVersion := events.ResourceVersion

		watcher, err := this.eventClient.CoreV1().Services(kubeapi.NamespaceAll).Watch(
			metav1.ListOptions{
				Watch:           true,
				ResourceVersion: resourceVersion})
		if err != nil {
			klog.Errorf("Failed to start watch for new Service events: %v", err)
			time.Sleep(time.Second)
			continue
		}

		watchChannel := watcher.ResultChan()
		// Inner loop, for update processing.
	inner_loop:
		for {
			select {
			case watchUpdate, ok := <-watchChannel:
				if !ok {
					klog.Errorf("Event watch channel closed")
					break inner_loop
				}

				if watchUpdate.Type == kubewatch.Error {
					if status, ok := watchUpdate.Object.(*metav1.Status); ok {
						klog.Errorf("Error during watch: %#v", status)
						break inner_loop
					}
					klog.Errorf("Received unexpected error: %#v", watchUpdate.Object)
					break inner_loop
				}

				if event, ok := watchUpdate.Object.(*kubeapi.Service); ok {
					//fmt.Println(event.InvolvedObject.Kind," ",event.InvolvedObject.Name," ",event.InvolvedObject.Namespace," ",event.Name," ",event.Source," ",event.Namespace," ",event.Reason," ",event.Kind)
					switch watchUpdate.Type {
					case kubewatch.Added:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: ADD,
							time: time.Now(),
							eventSource: ServiceEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
					case kubewatch.Modified:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: DELETE,
							time: time.Now(),
							eventSource: ServiceEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
					case kubewatch.Deleted:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: DELETE,
							time: time.Now(),
							eventSource: ServiceEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
						// Deleted events are silently ignored.
					default:
						klog.Warningf("Unknown watchUpdate.Type: %#v", watchUpdate.Type)
					}
				} else {
					klog.Errorf("Wrong object received: %v", watchUpdate)
				}

			case <-this.stopChannel:
				watcher.Stop()
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}


func (this *KubernetesEventSource)serviceAccountEvent(){
	for {
		events, err := this.eventClient.CoreV1().ServiceAccounts(kubeapi.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to load ServiceAccount events: %v", err)
			time.Sleep(time.Second)
			continue
		}

		// Do not write old events.

		resourceVersion := events.ResourceVersion

		watcher, err := this.eventClient.CoreV1().ServiceAccounts(kubeapi.NamespaceAll).Watch(
			metav1.ListOptions{
				Watch:           true,
				ResourceVersion: resourceVersion})
		if err != nil {
			klog.Errorf("Failed to start watch for new ServiceAccount events: %v", err)
			time.Sleep(time.Second)
			continue
		}

		watchChannel := watcher.ResultChan()
		// Inner loop, for update processing.
	inner_loop:
		for {
			select {
			case watchUpdate, ok := <-watchChannel:
				if !ok {
					klog.Errorf("Event watch channel closed")
					break inner_loop
				}

				if watchUpdate.Type == kubewatch.Error {
					if status, ok := watchUpdate.Object.(*metav1.Status); ok {
						klog.Errorf("Error during watch: %#v", status)
						break inner_loop
					}
					klog.Errorf("Received unexpected error: %#v", watchUpdate.Object)
					break inner_loop
				}

				if event, ok := watchUpdate.Object.(*kubeapi.ServiceAccount); ok {
					//fmt.Println(event.InvolvedObject.Kind," ",event.InvolvedObject.Name," ",event.InvolvedObject.Namespace," ",event.Name," ",event.Source," ",event.Namespace," ",event.Reason," ",event.Kind)
					switch watchUpdate.Type {
					case kubewatch.Added:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: ADD,
							time: time.Now(),
							eventSource: ServiceAccountEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
					case kubewatch.Modified:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: DELETE,
							time: time.Now(),
							eventSource: ServiceAccountEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
					case kubewatch.Deleted:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: DELETE,
							time: time.Now(),
							eventSource: ServiceAccountEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
						// Deleted events are silently ignored.
					default:
						klog.Warningf("Unknown watchUpdate.Type: %#v", watchUpdate.Type)
					}
				} else {
					klog.Errorf("Wrong object received: %v", watchUpdate)
				}

			case <-this.stopChannel:
				watcher.Stop()
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}

func (this *KubernetesEventSource)persistentVolumeEventWatch(){
	for {
		events, err := this.eventClient.CoreV1().PersistentVolumes().List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to load PersistentVolume events: %v", err)
			time.Sleep(time.Second)
			continue
		}

		// Do not write old events.

		resourceVersion := events.ResourceVersion

		watcher, err := this.eventClient.CoreV1().PersistentVolumes().Watch(
			metav1.ListOptions{
				Watch:           true,
				ResourceVersion: resourceVersion})
		if err != nil {
			klog.Errorf("Failed to start watch for new PersistentVolume events: %v", err)
			time.Sleep(time.Second)
			continue
		}

		watchChannel := watcher.ResultChan()
		// Inner loop, for update processing.
	inner_loop:
		for {
			select {
			case watchUpdate, ok := <-watchChannel:
				if !ok {
					klog.Errorf("Event watch channel closed")
					break inner_loop
				}

				if watchUpdate.Type == kubewatch.Error {
					if status, ok := watchUpdate.Object.(*metav1.Status); ok {
						klog.Errorf("Error during watch: %#v", status)
						break inner_loop
					}
					klog.Errorf("Received unexpected error: %#v", watchUpdate.Object)
					break inner_loop
				}

				if event, ok := watchUpdate.Object.(*kubeapi.PersistentVolume); ok {
					//fmt.Println(event.InvolvedObject.Kind," ",event.InvolvedObject.Name," ",event.InvolvedObject.Namespace," ",event.Name," ",event.Source," ",event.Namespace," ",event.Reason," ",event.Kind)
					switch watchUpdate.Type {
					case kubewatch.Added:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: ADD,
							time: time.Now(),
							eventSource: PersistentVolumeEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
					case kubewatch.Modified:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: DELETE,
							time: time.Now(),
							eventSource: PersistentVolumeEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
					case kubewatch.Deleted:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: DELETE,
							time: time.Now(),
							eventSource: PersistentVolumeEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
						// Deleted events are silently ignored.
					default:
						klog.Warningf("Unknown watchUpdate.Type: %#v", watchUpdate.Type)
					}
				} else {
					klog.Errorf("Wrong object received: %v", watchUpdate)
				}

			case <-this.stopChannel:
				watcher.Stop()
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}

func (this *KubernetesEventSource)clusterRoleEventWatch(){
	for {
		events, err := this.eventClient.RbacV1beta1().ClusterRoles().List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to load ClusterRole events: %v", err)
			time.Sleep(time.Second)
			continue
		}

		// Do not write old events.

		resourceVersion := events.ResourceVersion

		watcher, err := this.eventClient.RbacV1beta1().ClusterRoles().Watch(
			metav1.ListOptions{
				Watch:           true,
				ResourceVersion: resourceVersion})
		if err != nil {
			klog.Errorf("Failed to start watch for new ClusterRole events: %v", err)
			time.Sleep(time.Second)
			continue
		}

		watchChannel := watcher.ResultChan()
		// Inner loop, for update processing.
	inner_loop:
		for {
			select {
			case watchUpdate, ok := <-watchChannel:
				if !ok {
					klog.Errorf("Event watch channel closed")
					break inner_loop
				}

				if watchUpdate.Type == kubewatch.Error {
					if status, ok := watchUpdate.Object.(*metav1.Status); ok {
						klog.Errorf("Error during watch: %#v", status)
						break inner_loop
					}
					klog.Errorf("Received unexpected error: %#v", watchUpdate.Object)
					break inner_loop
				}

				if event, ok := watchUpdate.Object.(*rbac_v1beta1.ClusterRole); ok {
					//fmt.Println(event.InvolvedObject.Kind," ",event.InvolvedObject.Name," ",event.InvolvedObject.Namespace," ",event.Name," ",event.Source," ",event.Namespace," ",event.Reason," ",event.Kind)
					switch watchUpdate.Type {
					case kubewatch.Added:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: ADD,
							time: time.Now(),
							eventSource: ClusterRoleEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
					case kubewatch.Modified:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: DELETE,
							time: time.Now(),
							eventSource: ClusterRoleEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
					case kubewatch.Deleted:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: DELETE,
							time: time.Now(),
							eventSource: ClusterRoleEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
						// Deleted events are silently ignored.
					default:
						klog.Warningf("Unknown watchUpdate.Type: %#v", watchUpdate.Type)
					}
				} else {
					klog.Errorf("Wrong object received: %v", watchUpdate)
				}

			case <-this.stopChannel:
				watcher.Stop()
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}

func (this *KubernetesEventSource)ingressEventWatch(){
	for {
		events, err := this.eventClient.ExtensionsV1beta1().Ingresses(kubeapi.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to load Ingress events: %v", err)
			time.Sleep(time.Second)
			continue
		}

		// Do not write old events.

		resourceVersion := events.ResourceVersion

		watcher, err := this.eventClient.ExtensionsV1beta1().Ingresses(kubeapi.NamespaceAll).Watch(
			metav1.ListOptions{
				Watch:           true,
				ResourceVersion: resourceVersion})
		if err != nil {
			klog.Errorf("Failed to start watch for new Ingress events: %v", err)
			time.Sleep(time.Second)
			continue
		}

		watchChannel := watcher.ResultChan()
		// Inner loop, for update processing.
	inner_loop:
		for {
			select {
			case watchUpdate, ok := <-watchChannel:
				if !ok {
					klog.Errorf("Event watch channel closed")
					break inner_loop
				}

				if watchUpdate.Type == kubewatch.Error {
					if status, ok := watchUpdate.Object.(*metav1.Status); ok {
						klog.Errorf("Error during watch: %#v", status)
						break inner_loop
					}
					klog.Errorf("Received unexpected error: %#v", watchUpdate.Object)
					break inner_loop
				}

				if event, ok := watchUpdate.Object.(*ext_v1beta1.Ingress); ok {
					//fmt.Println(event.InvolvedObject.Kind," ",event.InvolvedObject.Name," ",event.InvolvedObject.Namespace," ",event.Name," ",event.Source," ",event.Namespace," ",event.Reason," ",event.Kind)
					switch watchUpdate.Type {
					case kubewatch.Added:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: ADD,
							time: time.Now(),
							eventSource: IngressEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
					case kubewatch.Modified:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: DELETE,
							time: time.Now(),
							eventSource: IngressEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
					case kubewatch.Deleted:
						defaultEvents:=&LocalEventBuffer{
							events: *event,
							eventType: DELETE,
							time: time.Now(),
							eventSource: IngressEvent,
						}
						select {
						case this.localEventsBuffer <- defaultEvents:
							// Ok, buffer not full.
						default:
							// Buffer full, need to drop the event.
							klog.Errorf("Event buffer full, dropping event")
						}
						// Deleted events are silently ignored.
					default:
						klog.Warningf("Unknown watchUpdate.Type: %#v", watchUpdate.Type)
					}
				} else {
					klog.Errorf("Wrong object received: %v", watchUpdate)
				}

			case <-this.stopChannel:
				watcher.Stop()
				klog.Infof("Event watching stopped")
				return
			}
		}
	}
}

func (this *KubernetesEventSource) watch() {
	// Outer loop, for reconnections
		go this.eventWatch()
		go this.clusterRoleEventWatch()
		go this.secretEventWatch()
		go this.serviceAccountEvent()
		go this.serviceEventWatch()
		go this.ingressEventWatch()
		go this.namespaceWatch()
		go this.comfigMapWatch()
		go this.persistentVolumeEventWatch()

	for{
	}
}

var Events = make(chan *kubeapi.Event,LocalEventsBufferSize*10)

func NewKubernetesSource(uri *url.URL) (*KubernetesEventSource, error) {
	kubeClient, err := kubernetes.GetKubernetesClient(uri)
	if err != nil {
		klog.Errorf("Failed to create kubernetes client,because of %v", err)
		return nil, err
	}
	//eventClient := kubeClient.CoreV1().Events(kubeapi.NamespaceAll)
	//EventsBuffer := make(chan interface{},LocalEventsBufferSize*10)

	result := KubernetesEventSource{
		localEventsBuffer: make(chan *LocalEventBuffer,LocalEventsBufferSize),
		stopChannel:       make(chan struct{}),
		eventClient:       kubeClient,
	}
//	result.Start(kubeapi.NamespaceAll)
	go result.watch()
	fmt.Println("------begin to watch ------------")
	fmt.Println("----------------------------------------------")
	return &result, nil
}

