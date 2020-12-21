package cmd

import (
	"time"

	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/log"
	"github.com/q8s-io/cluster-detector/pkg/provider/kube"
	"github.com/q8s-io/cluster-detector/pkg/sinks"
	"github.com/q8s-io/cluster-detector/pkg/sinks/kafka"
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
	//log.Println("--------------Pod Inspections--------------")
	for i := range kafka.KafkaPodInspection {
		log.PrintLog(log.LogMess{
			Namespace: i.Namespace,
			Name:      i.PodName,
			Kind:      "Pod",
			Type:      "Pod Inspections",
			Time:      time.Now(),
			Info:      map[string]interface{}{
				"podStatus":i.Status,
				"podIP":i.PodIP,
				"hostIP":i.HostIP,
				"nodeName":i.NodeName,
			},
		})
	}
	klog.Infof("Starting nodeInspection")
}
