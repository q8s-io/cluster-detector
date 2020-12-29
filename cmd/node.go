package cmd

import (
	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/provider/filter"
	"github.com/q8s-io/cluster-detector/pkg/provider/kube"
)

func RunNodeInspection() {
	argSource := config.Config.Source
	argKafkaSink := &config.Config.NodeInspectionConfig.KafkaNodeConfig
	if argSource.KubernetesURL == "" {
		klog.Fatal("Wrong sources specified")
	}
	sourceFactory := kube.NewSourceFactory()
	nodeResources, buildErr := sourceFactory.BuildNodeInspection(argSource)
	if buildErr != nil {
		klog.Fatalf("Failed to create sources: %v", buildErr)
	}
	if argKafkaSink.Enabled==true{
		nodeFilter:=filter.NewFilterFactory()
		nodeFilter.KafkaFilter(nodeResources,&config.Config)
	}
	klog.Infof("Starting nodeInspection")
}
