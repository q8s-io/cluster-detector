package kafka

import (
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/provider/collect"
	"github.com/q8s-io/cluster-detector/pkg/provider/filter"
)

// start
func RunKafka() {
	// cfgSource := config.Config.Source
	cfg := &config.Config
	sourceFactory := collect.NewSourceFactory()
	filterFactory := filter.NewFilterFactory()

	if cfg.EventsConfig.KafkaEventConfig.Enabled == true {
		eventResources := sourceFactory.BuildEvents()
		filterFactory.KafkaFilter(eventResources, &config.Config)
	}
	if cfg.PodInspectionConfig.Enabled == true {
		podResources := sourceFactory.BuildPods()
		filterFactory.KafkaFilter(podResources, &config.Config)
	}
	if cfg.NodeInspectionConfig.Enabled == true {
		nodeResources := sourceFactory.BuildNodes()
		filterFactory.KafkaFilter(nodeResources, &config.Config)
	}
	if cfg.DeleteInspectionConfig.Enabled == true {
		deleteResources := sourceFactory.BuildDeletes()
		filterFactory.KafkaFilter(deleteResources, &config.Config)
	}
}
