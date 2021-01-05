package initChan

import (
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/provider/collect"
	"github.com/q8s-io/cluster-detector/pkg/provider/filter"
)

const (
	EventType  = "event"
	PodType    = "pod"
	NodeType   = "node"
	DeleteType = "delete"
)

// start
func RunKafka() {
	// cfgSource := config.Config.Source
	cfg := &config.Config
	sourceFactory := collect.NewSourceFactory()
	filterFactory := filter.NewFilterFactory()

	if cfg.EventsConfig.KafkaEventConfig.Enabled == true {
		sourceFactory.InitEventChan()
		filterFactory.KafkaFilter(EventType, &config.Config)
	}
	if cfg.PodInspectionConfig.Enabled == true {
		sourceFactory.InitPodChan()
		filterFactory.KafkaFilter(PodType, &config.Config)
	}
	if cfg.NodeInspectionConfig.Enabled == true {
		sourceFactory.InitNodeChan()
		filterFactory.KafkaFilter(NodeType, &config.Config)
	}
	if cfg.DeleteInspectionConfig.Enabled == true {
		sourceFactory.InitDeleteChan()
		filterFactory.KafkaFilter(DeleteType, &config.Config)
	}
}
