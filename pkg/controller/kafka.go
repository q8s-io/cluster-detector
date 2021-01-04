package controller

import (
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/provider/filter"
	"github.com/q8s-io/cluster-detector/pkg/provider/kube"
)

func RunKafka() {
	cfgSource := config.Config.Source
	kafkaCfg := &config.Config.EventsConfig.KafkaEventConfig

	if kafkaCfg.Enabled == true {
		sourceFactory := kube.NewSourceFactory()
		eventResources := sourceFactory.BuildEvents(cfgSource)
		eventFilter := filter.NewFilterFactory()
		eventFilter.KafkaFilter(eventResources, &config.Config)
	}
}
