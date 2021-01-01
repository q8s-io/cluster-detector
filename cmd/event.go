package cmd

import (
	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/provider/filter"
	"github.com/q8s-io/cluster-detector/pkg/provider/kube"
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
	if argKafkaSink.Enabled == true {
		eventFilter := filter.NewFilterFactory()
		eventFilter.KafkaFilter(eventResources, &config.Config)
	}
}
