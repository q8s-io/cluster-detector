package cmd

import (
	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/provider/filter"
	"github.com/q8s-io/cluster-detector/pkg/provider/kube"
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
	if argKafkaSink.Enabled == true {
		podFilter := filter.NewFilterFactory()
		podFilter.KafkaFilter(podResources, &config.Config)
	}
	klog.Infof("Starting nodeInspection")
}
