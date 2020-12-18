package cmd

import (
	"log"

	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/provider/kube"
	"github.com/q8s-io/cluster-detector/pkg/sinks"
	"github.com/q8s-io/cluster-detector/pkg/sinks/kafka"
)

func RunEventsWatch() {
	argSource := config.Config.Source
	argKafkaSink := &config.Config.EventsConfig.KafkaEventConfig
	klog.Info(argSource.KubernetesURL, argKafkaSink)
	sourceFactory := kube.NewSourceFactory()
	eventResources, err := sourceFactory.BuildEvents(argSource)
	if err != nil {
		klog.Info("Failed to create sources: %v", err)
	}
	sinksFactory := sinks.NewSinkFactory()
	if argKafkaSink.Enabled == true {

		kafkaSink, kafkaErr := sinksFactory.BuildEventKafka(argKafkaSink)
		if kafkaErr != nil {
			klog.Infof("Failed to create kafkaSink: %v", kafkaErr)
		}
		kafkaSink.ExportEvents(eventResources)
	}
	log.Println("--------------KafKa event--------------")
	for i := range kafka.KafkaEventInspection {
		log.Println(i)
	}
	klog.Info("Starting eventer")
}
