package cmd

import (
	"fmt"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/provider/kube"
	"github.com/q8s-io/cluster-detector/pkg/sinks"
	"k8s.io/klog"
	"log"
)

func RunPodInspection() {
	argSource := config.Config.Source
	argKafkaSink := &config.Config.PodInspectionConfig.KafkaPodConfig
	if argSource.KubernetesURL == "" {
		klog.Fatal("Wrong sources specified")
	}
	sourceFactory := kube.NewSourceFactory()
	podResources, buildErr := sourceFactory.BuildPodInspection(argSource)
	if buildErr != nil {
		klog.Fatalf("Failed to create sources: %v", buildErr)
	}
	sinksFactory := sinks.NewSinkFactory()
	if argKafkaSink.Enabled == true {
		kafkaSink, kafkaErr := sinksFactory.BuildPodKafka(argKafkaSink)
		if kafkaErr != nil {
			klog.Fatalf("Failed to create kafkaSink: %v", kafkaErr)
		}
		kafkaSink.ExportPodInspection(podResources)
	}
	fmt.Println("--------------Pod Inspections--------------")
	for i := range *podResources {
		log.Println(i)
	}
	klog.Infof("Starting nodeInspection")
}
