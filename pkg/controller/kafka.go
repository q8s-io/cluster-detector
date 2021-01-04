package controller

import (
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/provider/collect"
	"github.com/q8s-io/cluster-detector/pkg/provider/filter"
)

func RunKafka() {
	// cfgSource := config.Config.Source
	cfg := &config.Config
	sourceFactory := collect.NewSourceFactory()
	eventFilter := filter.NewFilterFactory()

	if cfg.EventsConfig.KafkaEventConfig.Enabled == true {
		eventResources := sourceFactory.BuildEvents()
		eventFilter.KafkaFilter(eventResources, &config.Config)
	}
	if cfg.PodInspectionConfig.Enabled == true {
		eventResources := sourceFactory.BuildPods()
		eventFilter.KafkaFilter(eventResources, &config.Config)
	}
	if cfg.NodeInspectionConfig.Enabled == true {
		eventResources := sourceFactory.BuildNodes()
		eventFilter.KafkaFilter(eventResources, &config.Config)
	}
	if cfg.DeleteInspectionConfig.Enabled == true {
		eventResources := sourceFactory.BuildDeletes()
		eventFilter.KafkaFilter(eventResources, &config.Config)
	}
}
