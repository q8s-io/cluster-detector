package controller

import (
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/provider/filter"
	"github.com/q8s-io/cluster-detector/pkg/provider/kube"
)

func RunKafka() {
	//cfgSource := config.Config.Source
	kafkaEventCfg := &config.Config.EventsConfig.KafkaEventConfig
	kafkaPodCfg:=&config.Config.PodInspectionConfig
	kafkaNodeCfg:=&config.Config.NodeInspectionConfig
	kafkaDeleteCfg:=&config.Config.DeleteInspectionConfig
	if kafkaEventCfg.Enabled == true {
		sourceFactory := kube.NewSourceFactory()
		eventResources := sourceFactory.BuildEvents()
		eventFilter := filter.NewFilterFactory()
		eventFilter.KafkaFilter(eventResources, &config.Config)
	}
	if kafkaPodCfg.Enabled==true{
		sourceFactory := kube.NewSourceFactory()
		eventResources := sourceFactory.BuildPods()
		eventFilter := filter.NewFilterFactory()
		eventFilter.KafkaFilter(eventResources, &config.Config)
	}
	if kafkaNodeCfg.Enabled==true{
		sourceFactory := kube.NewSourceFactory()
		eventResources := sourceFactory.BuildNodes()
		eventFilter := filter.NewFilterFactory()
		eventFilter.KafkaFilter(eventResources, &config.Config)
	}
	if kafkaDeleteCfg.Enabled==true{
		sourceFactory := kube.NewSourceFactory()
		eventResources := sourceFactory.BuildDeletes()
		eventFilter := filter.NewFilterFactory()
		eventFilter.KafkaFilter(eventResources, &config.Config)
	}
}
