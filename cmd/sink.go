package cmd

import (
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/provider/sinks"
)

func Sink() {
	arg := &config.Config
	if arg.DeleteInspectionConfig.KafkaDeleteConfig.Enabled == true ||
		arg.PodInspectionConfig.KafkaPodConfig.Enabled == true ||
		arg.NodeInspectionConfig.KafkaNodeConfig.Enabled == true ||
		arg.EventsConfig.KafkaEventConfig.Enabled == true {
		sink := sinks.NewSinkFactory()
		sink.BuildKafka()
	}
}
