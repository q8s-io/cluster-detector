package provider

import (
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/provider/sinks"
)

func Sink() {
	cfg := &config.Config
	if cfg.DeleteInspectionConfig.KafkaDeleteConfig.Enabled == true ||
		cfg.PodInspectionConfig.KafkaPodConfig.Enabled == true ||
		cfg.NodeInspectionConfig.KafkaNodeConfig.Enabled == true ||
		cfg.EventsConfig.KafkaEventConfig.Enabled == true {
		sink := sinks.NewSinkFactory()
		sink.BuildKafka()
	}
}
