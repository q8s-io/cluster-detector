package cmd

import (
	"log"

	"github.com/q8s-io/cluster-detector/pkg/sinks"
	"github.com/q8s-io/cluster-detector/pkg/sinks/kafka"

	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/provider/kube"
)

func RunUnusedInspection() {
	argSource := config.Config.Source
	argKafkaSink := &config.Config.DeleteInspectionConfig.KafkaDeleteConfig
	klog.Info(argSource.KubernetesURL, argKafkaSink)
	sourceFactory := kube.NewSourceFactory()
	eventResources, err := sourceFactory.BuildUnusedResource(argSource)
	if err != nil {
		klog.Info("Failed to create sources: %v", err)
	}
	sinksFactory := sinks.NewSinkFactory()
	if argKafkaSink.Enabled == true {
		kafkaSink, kafkaErr := sinksFactory.BuildDeleteKafka(argKafkaSink)
		if kafkaErr != nil {
			klog.Fatalf("Failed to create kafkaSink: %v", kafkaErr)
		}
		kafkaSink.ExportDeleteInspection(eventResources)
	}
	log.Println("--------------KafKa Deleted Inspections--------------")
	for i := range kafka.KafkaDeleteInspection {
		log.Println(i)
	}
	klog.Info("Starting unused")
	klog.Infof("Starting DeleteInspection")
}
