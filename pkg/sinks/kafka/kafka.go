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

package kafka

import (
	"github.com/q8s-io/cluster-detector/configs"
	kafka_common "github.com/q8s-io/cluster-detector/pkg/common/kafka"
	"github.com/q8s-io/cluster-detector/pkg/core"
	"github.com/q8s-io/cluster-detector/pkg/sinks/tool"
	kube_api "k8s.io/api/core/v1"
	"k8s.io/klog"
	"sync"
)

const (
	EVENT_KAFKA_SINK     = "EventKafkaSink"
	NODE_KAFKA_SINK      = "NodeKafkaSink"
	POD_KAFKA_SINK       = "PodKafkaSink"
	WARNING          int = 2
	NORMAL           int = 1
)

type kafkaSink struct {
	kafka_common.KafkaClient
	sync.RWMutex
}

type eventKafkaSink struct {
	level      int
	namespaces []string
	kinds      []string
	kafkaSink  *kafkaSink
}

func (sink *eventKafkaSink) Name() string {
	return EVENT_KAFKA_SINK
}

func (sink *eventKafkaSink) Stop() {
	return
}

// filter event by given conditions
func (sink *eventKafkaSink) skipEvent(event *kube_api.Event) bool {
	// filter level
	// If warning level is given, don’t care about normal level.
	// However, giving normal level also accept warning level.
	score := tool.GetLevel(event.Type)
	if score < sink.level {
		return true
	}

	// filter namespaces
	if sink.namespaces != nil {
		skip := true
		for _, namespace := range sink.namespaces {
			if namespace == event.Namespace {
				skip = false
			}
		}
		if skip {
			return true
		}
	}

	// filter kinds
	if sink.kinds != nil {
		skip := true
		for _, kind := range sink.kinds {
			if kind == event.InvolvedObject.Kind {
				skip = false
				break
			}
		}
		if skip {
			return true
		}
	}
	return false
}

func (sink *eventKafkaSink) ExportEvents(eventBatch *core.EventBatch) {
	sink.kafkaSink.Lock()
	defer sink.kafkaSink.Unlock()

	for _, event := range eventBatch.Events {
		if !sink.skipEvent(event) {
			sinkPoint := &core.EventSinkPoint{}
			sinkPoint, err := sinkPoint.EventToPoint(event)
			if err != nil {
				klog.Warningf("Failed to convert event to point: %v", err)
			}

			err = sink.kafkaSink.KafkaClient.ProduceKafkaMessage(*sinkPoint)
			if err != nil {
				klog.Errorf("Failed to produce event message: %s", err)
			}
		}
	}
}

func NewEventKafkaSink(kafkaEventConfig *configs.KafkaEventConfig) (core.EventSink, error) {
	// get kafka Client
	client, err := kafka_common.NewKafkaClient(&kafkaEventConfig.Kafka, kafka_common.EventsTopic)
	if err != nil {
		return nil, err
	}

	// init kafkaSink
	sink := &kafkaSink{
		KafkaClient: client,
		RWMutex:     sync.RWMutex{},
	}
	eventKafkaSink := &eventKafkaSink{
		level:     WARNING,
		kafkaSink: sink,
	}
	eventKafkaSink.level = tool.GetLevel(kafkaEventConfig.Level)
	eventKafkaSink.namespaces = kafkaEventConfig.Namespaces
	eventKafkaSink.kinds = kafkaEventConfig.Kinds
	return eventKafkaSink, nil
}

// nodeInspection sink
type nodeKafkaSink kafkaSink

func (sink *nodeKafkaSink) Name() string {
	return NODE_KAFKA_SINK
}

func (sink *nodeKafkaSink) ExportNodeInspection(nodeBatch *core.NodeInspectionBatch) {
	sink.Lock()
	defer sink.Unlock()
	//klog.V(0).Info(nodeBatch)
	err := sink.KafkaClient.ProduceKafkaMessage(nodeBatch)
	if err != nil {
		klog.Errorf("Failed to produce NodeInspection message: %s", err)
	}
}

func (sink *nodeKafkaSink) Stop() {
	return
}

func NewNodeKafkaSink(kafkaConfig *configs.Kafka) (core.NodeSink, error) {
	// get kafka Client
	client, err := kafka_common.NewKafkaClient(kafkaConfig, kafka_common.NodesTopic)
	if err != nil {
		return nil, err
	}
	// init kafkaSink
	kafkaSink := &nodeKafkaSink{
		KafkaClient: client,
	}
	return kafkaSink, nil
}

type podKafkaSink struct {
	namespaces core.Namespaces
	kafkaSink  *kafkaSink
}

func (sink *podKafkaSink) Name() string {
	return POD_KAFKA_SINK
}

func (sink *podKafkaSink) Stop() {
	return
}

func (sink *podKafkaSink) ExportPodInspection(podBatch *core.PodInspectionBatch) {
	sink.kafkaSink.Lock()
	defer sink.kafkaSink.Unlock()
	batch := &core.PodInspectionBatch{
		TimeStamp:   podBatch.TimeStamp,
		Inspections: make([]*core.PodInspection, 0),
	}
	for _, inspection := range podBatch.Inspections {
		skip := sink.namespaces.SkipPod(inspection)
		if skip {
			continue
		}
		batch.Inspections = append(batch.Inspections, inspection)
	}
	err := sink.kafkaSink.KafkaClient.ProduceKafkaMessage(batch)
	if err != nil {
		klog.Errorf("Failed to produce podInspection message: %s", err)
	}
}

func NewPodKafkaSink(kafkaPodConfig *configs.KafkaPodConfig) (core.PodSink, error) {
	// get kafka Client
	client, err := kafka_common.NewKafkaClient(&kafkaPodConfig.Kafka, kafka_common.PodsTopic)
	if err != nil {
		return nil, err
	}

	// init kafkaSink
	sink := &kafkaSink{
		KafkaClient: client,
		RWMutex:     sync.RWMutex{},
	}
	podKafkaSink := &podKafkaSink{
		kafkaSink:  sink,
		namespaces: kafkaPodConfig.Namespaces,
	}
	return podKafkaSink, nil
}
