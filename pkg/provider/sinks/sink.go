package sinks

import (
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
)

func Sink() {
	cfg := &config.Config
	if cfg.DeleteInspectionConfig.KafkaDeleteConfig.Enabled == true ||
		cfg.PodInspectionConfig.KafkaPodConfig.Enabled == true ||
		cfg.NodeInspectionConfig.KafkaNodeConfig.Enabled == true ||
		cfg.EventsConfig.KafkaEventConfig.Enabled == true {
		sink := NewSinkFactory()
		sink.BuildKafka()
	}
}
