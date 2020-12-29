package sinks

import (
	/*"github.com/q8s-io/cluster-detector/pkg/entity"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/provider/sinks/kafka/deletekafka"
	"github.com/q8s-io/cluster-detector/pkg/provider/sinks/kafka/eventkafka"
	"github.com/q8s-io/cluster-detector/pkg/provider/sinks/kafka/podkafka"*/
)

type SinkFactory struct {
}

/*func (_ *SinkFactory) BuildEventKafka(kafkaEventConfig *config.KafkaEventConfig) (entity.EventSink, error) {
	return eventkafka.NewEventKafkaSink(kafkaEventConfig)
}*/

/*func (_ *SinkFactory) BuildNodeKafka(kafkaConfig *config.Kafka) (entity.NodeSink, error) {
	//return nodekafka.NewNodeKafkaSink(kafkaConfig)
}*/

/*func (_ *SinkFactory) BuildPodKafka(kafkaPodConfig *config.KafkaPodConfig) (entity.PodSink, error) {
	return podkafka.NewPodKafkaSink(kafkaPodConfig)
}*/

/*func (_ *SinkFactory) BuildEventWebHook(webHookEventConfig *configs.WebHookEventConfig) (core.EventSink, error) {
	return webhook.NewEventWebHookSink(webHookEventConfig)
}*/

/*func (_ *SinkFactory) BuildNodeWebHook(webHookConfig *configs.WebHook) (core.NodeSink, error) {
	return webhook.NewNodeWebHookSink(webHookConfig)
}*/

/*func (_ *SinkFactory) BuildPodWebHook(webHookConfig *configs.WebHookPodConfig) (core.PodSink, error) {
	return webhook.NewPodWebHookSink(webHookConfig)
}*/

/*func (_ *SinkFactory) BuildDeleteKafka(kafkaDeleteConfig *config.KafkaDeleteConfig) (entity.DeleteSink, error) {
	return deletekafka.NewDeleteKafkaSink(kafkaDeleteConfig)
}*/

func NewSinkFactory() *SinkFactory {
	return &SinkFactory{}
}
