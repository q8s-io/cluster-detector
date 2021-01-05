package initChan

import (
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/provider/collect"
	"github.com/q8s-io/cluster-detector/pkg/provider/filter"
	"github.com/q8s-io/cluster-detector/pkg/provider/filter/kafka"
)

// start
func RunKafka() {
	// cfgSource := config.Config.Source
	cfg := &config.Config
	sourceFactory := collect.NewSourceFactory()
	filterFactory := filter.NewFilterFactory()

	if cfg.EventsConfig.KafkaEventConfig.Enabled == true {
		sourceFactory.InitEventChan()
		filterFactory.KafkaFilter(kafka.EventType, &config.Config)
	}
	if cfg.PodInspectionConfig.Enabled == true {
		sourceFactory.InitPodChan()
		filterFactory.KafkaFilter(kafka.PodType, &config.Config)
	}
	if cfg.NodeInspectionConfig.Enabled == true {
		sourceFactory.InitNodeChan()
		filterFactory.KafkaFilter(kafka.NodeType, &config.Config)
	}
	if cfg.DeleteInspectionConfig.Enabled == true {
		sourceFactory.InitDeleteChan()
		filterFactory.KafkaFilter(kafka.DeleteType, &config.Config)
	}
}
