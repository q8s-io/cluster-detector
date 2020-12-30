package cmd

import (
	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/provider/filter"
	"github.com/q8s-io/cluster-detector/pkg/provider/kube"
)

func RunUnusedInspection() {
	argSource := config.Config.Source
	argKafkaSink := &config.Config.DeleteInspectionConfig.KafkaDeleteConfig
	klog.Info(argSource.KubernetesURL, argKafkaSink)
	sourceFactory := kube.NewSourceFactory()
	deleteResources, err := sourceFactory.BuildUnusedResource(argSource)
	if err != nil {
		klog.Info("Failed to create sources: %v", err)
	}
	if argKafkaSink.Enabled==true{
		deleteFilter:=filter.NewFilterFactory()
		deleteFilter.KafkaFilter(deleteResources,&config.Config)
	}
	//klog.Info("Starting unused")
	klog.Infof("Starting DeleteInspection")
}
