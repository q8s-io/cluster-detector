package cmd

import (
	"log"

	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/provider/kube"
)

func RunEventsWatch() {
	argSource := config.Config.Source
	argKafkaSink := &config.Config.EventsConfig.KafkaEventConfig
	argWebHookSink := &config.Config.EventsConfig.WebHookEventConfig
	klog.Info(argSource.KubernetesURL, argKafkaSink, argWebHookSink)

	sourceFactory := kube.NewSourceFactory()
	eventResources, err := sourceFactory.BuildEvents(argSource)
	if err != nil {
		klog.Info("Failed to create sources: %v", err)
	}

	for i := range *eventResources {
		log.Println(i)
	}
	// sink factory
	//sinksFactory := sinks.NewSinkFactory()
	//var sinkList []entity.EventSink

	// kafka sink
	//if argKafkaSink.Enabled == true {
	//	kafkaSink, kafkaErr := sinksFactory.BuildEventKafka(argKafkaSink)
	//	if kafkaErr != nil {
	//		klog.Infof("Failed to create kafkaSink: %v", kafkaErr)
	//	}
	//	sinkList = append(sinkList, kafkaSink)
	//}

	// WebHook sink
	//if argWebHookSink.Enabled == true {
	//	webHookSink, webHookErr := sinksFactory.BuildEventWebHook(argWebHookSink)
	//	if webHookErr != nil {
	//		klog.Fatalf("Failed to create webHookSink: %v", webHookErr)
	//	}
	//	sinkList = append(sinkList, webHookSink)
	//}

	//if len(sinkList) == 0 {
	//	klog.Info("No available sink to use")
	//}
	//for _, sink := range sinkList {
	//	klog.Infof("Starting with %s sink", sink.Name())
	//}

	// SinkManager put events to sinkList
	//sinkManager, err := sinkEvent.NewEventSinkManager(sinkList, sinkEvent.DefaultSinkExportEventsTimeout, sinkEvent.DefaultSinkStopTimeout)
	//if err != nil {
	//	klog.Infof("Failed to create sink manager: %v", err)
	//}

	// main manager
	//manager, err := event.NewManager(eventResources, sinkManager, *argFrequency)
	//if err != nil {
	//	klog.Infof("Failed to create main manager: %v", err)
	//}

	//manager.Start()
	klog.Info("Starting eventer")
}
