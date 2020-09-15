// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sinks

import (
	"github.com/q8s-io/cluster-detector/configs"
	"github.com/q8s-io/cluster-detector/pkg/core"
	"github.com/q8s-io/cluster-detector/pkg/sinks/kafka"
	"github.com/q8s-io/cluster-detector/pkg/sinks/webhook"
)

type SinkFactory struct {
}

func (_ *SinkFactory) BuildEventKafka(kafkaEventConfig *configs.KafkaEventConfig) (core.EventSink, error) {
	return kafka.NewEventKafkaSink(kafkaEventConfig)
}

func (_ *SinkFactory) BuildNodeKafka(kafkaConfig *configs.Kafka) (core.NodeSink, error) {
	return kafka.NewNodeKafkaSink(kafkaConfig)
}

func (_ *SinkFactory) BuildPodKafka(kafkaPodConfig *configs.KafkaPodConfig) (core.PodSink, error) {
	return kafka.NewPodKafkaSink(kafkaPodConfig)
}

func (_ *SinkFactory) BuildEventWebHook(webHookEventConfig *configs.WebHookEventConfig) (core.EventSink, error) {
	return webhook.NewEventWebHookSink(webHookEventConfig)
}

func (_ *SinkFactory) BuildNodeWebHook(webHookConfig *configs.WebHook) (core.NodeSink, error) {
	return webhook.NewNodeWebHookSink(webHookConfig)
}

func (_ *SinkFactory) BuildPodWebHook(webHookConfig *configs.WebHookPodConfig) (core.PodSink, error) {
	return webhook.NewPodWebHookSink(webHookConfig)
}

func NewSinkFactory() *SinkFactory {
	return &SinkFactory{}
}
