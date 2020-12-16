package cmd

import (
	"fmt"
	"github.com/q8s-io/cluster-detector/pkg/sinks"
	"github.com/q8s-io/cluster-detector/pkg/sinks/kafka"
	"log"

	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/provider/kube"
)

func RunUnusedInspection() {
	argSource := config.Config.Source
	argKafkaSink := &config.Config.DeleteInspectionConfig.KafkaDeleteConfig
	//argWebHookSink := &config.Config.EventsConfig.WebHookEventConfig
	klog.Info(argSource.KubernetesURL, argKafkaSink)

	sourceFactory := kube.NewSourceFactory()
	eventResources, err := sourceFactory.BuildUnusedResource(argSource)
	if err != nil {
		klog.Info("Failed to create sources: %v", err)
	}

	//log.Println(eventResources)
	/*for i := range *eventResources {
		log.Println(i)
	}*/
	sinksFactory := sinks.NewSinkFactory()
	//var sinkList []entity.DeleteSink
	if argKafkaSink.Enabled == true {
		kafkaSink, kafkaErr := sinksFactory.BuildDeleteKafka(argKafkaSink)
		if kafkaErr != nil {
			klog.Fatalf("Failed to create kafkaSink: %v", kafkaErr)
		}
		kafkaSink.ExportDeleteInspection(eventResources)
	}
	fmt.Println("--------------KafKa Deleted Inspections--------------")
	for i := range kafka.KafkaDeleteInspection {
		log.Println(i)
	}
	klog.Info("Starting unused")
	//	sinkList = append(sinkList, kafkaSink)
	//}
	//for _, sink := range sinkList {
	//	klog.Infof("Starting with %s sink", sink.Name())
	//}
	//sinkManagers, smErr := sinkDelete.NewDeleteSinkManager(sinkList, sinkDelete.DefaultSinkExportNodesTimeout, sinkDelete.DefaultSinkStopTimeout)
	//if smErr != nil {
	//	klog.Fatalf("Failed to create sink manager: %v", smErr)
	//}
	//manager, managerErr := delete.NewManager(deleteResource[0], sinkManagers, *argFrequency)
	//if managerErr != nil {
	//	klog.Fatalf("Failed to create main manager: %v", managerErr)
	//}
	//manager.Start()
	klog.Infof("Starting DeleteInspection")
}
