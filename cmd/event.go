package cmd

import (
	"github.com/q8s-io/cluster-detector/configs"
	event_core "github.com/q8s-io/cluster-detector/pkg/core"
	m_event "github.com/q8s-io/cluster-detector/pkg/manager/event"
	"github.com/q8s-io/cluster-detector/pkg/sinks"
	sm_event "github.com/q8s-io/cluster-detector/pkg/sinks/manager/event"
	"github.com/q8s-io/cluster-detector/pkg/sources"
	"k8s.io/klog"
)

func RunEventsWatch() {

	argSource := configs.Config.Source
	argKafkaSink := &configs.Config.EventsConfig.KafkaEventConfig
	//argWebHookSink := &configs.Config.EventsConfig.WebHookEventConfig

	// sources
	if argSource.KubernetesURL == "" {
		klog.Fatal("Wrong sources specified")
	}
	sourceFactory := sources.NewSourceFactory()
	eventResources, err := sourceFactory.BuildEvents(argSource)
	if err != nil {
		klog.Fatalf("Failed to create sources: %v", err)
	}
	if len(eventResources) != 1 {
		klog.Fatal("Requires exactly 1 source")
	}

	// sink factory
	sinksFactory := sinks.NewSinkFactory()
	var sinkList []event_core.EventSink
	// kafka sink
	if argKafkaSink.Enabled == true {
		kafkaSink, kafkaErr := sinksFactory.BuildEventKafka(argKafkaSink)
		if kafkaErr != nil {
			klog.Fatalf("Failed to create kafkaSink: %v", kafkaErr)
		}
		sinkList = append(sinkList, kafkaSink)
	}

	// WebHook sink
	/*if argWebHookSink.Enabled == true {
		webHookSink, webHookErr := sinksFactory.BuildEventWebHook(argWebHookSink)
		if webHookErr != nil {
			klog.Fatalf("Failed to create webHookSink: %v", webHookErr)
		}
		sinkList = append(sinkList, webHookSink)
	}*/

	if len(sinkList) == 0 {
		klog.Fatal("No available sink to use")
	}
	for _, sink := range sinkList {
		klog.Infof("Starting with %s sink", sink.Name())
	}

	// SinkManager put events to sinkList
	sinkManager, err := sm_event.NewEventSinkManager(sinkList, sm_event.DefaultSinkExportEventsTimeout, sm_event.DefaultSinkStopTimeout)
	if err != nil {
		klog.Fatalf("Failed to create sink manager: %v", err)
	}

	// main manager
	manager, err := m_event.NewManager(eventResources[0], sinkManager, *argFrequency)
	if err != nil {
		klog.Fatalf("Failed to create main manager: %v", err)
	}

	manager.Start()
	klog.Infof("Starting eventer")
}



