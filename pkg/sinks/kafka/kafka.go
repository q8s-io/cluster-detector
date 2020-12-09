package kafka

import (
	"fmt"
	"sync"

	"k8s.io/api/core/v1"
	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/entity"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/kafka"
	"github.com/q8s-io/cluster-detector/pkg/sinks/tool"
)

const (
	EVENT_KAFKA_SINK     = "EventKafkaSink"
	NODE_KAFKA_SINK      = "NodeKafkaSink"
	POD_KAFKA_SINK       = "PodKafkaSink"
	WARNING          int = 2
	NORMAL           int = 1
)

type kafkaSink struct {
	kafka.KafkaClient
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
func (sink *eventKafkaSink) skipEvent(event *v1.Event) bool {
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
		fmt.Println("----sink-kind and this kind-----------")
		fmt.Printf("%v\n", sink.kinds)
		fmt.Println(event.InvolvedObject.Kind)
		fmt.Println("------------------------")
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

func (sink *eventKafkaSink) ExportEvents(eventBatch *entity.EventBatch) {
	sink.kafkaSink.Lock()
	defer sink.kafkaSink.Unlock()

	/*for _, event := range eventBatch.Events {
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
	}*/
	err := sink.kafkaSink.KafkaClient.ProduceKafkaMessage(*eventBatch)
	//fmt.Printf("%+v\n",eventBatch.Events)
	if err != nil {
		klog.Errorf("Failed to produce NodeInspection message: %s", err)
	}
}

func NewEventKafkaSink(kafkaEventConfig *config.KafkaEventConfig) (entity.EventSink, error) {
	// get kafka Client
	client, err := kafka.NewKafkaClient(&kafkaEventConfig.Kafka, kafka.EventsTopic)
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

func (sink *nodeKafkaSink) ExportNodeInspection(nodeBatch *entity.NodeInspectionBatch) {
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

func NewNodeKafkaSink(kafkaConfig *config.Kafka) (entity.NodeSink, error) {
	// get kafka Client
	client, err := kafka.NewKafkaClient(kafkaConfig, kafka.NodesTopic)
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
	namespaces entity.Namespaces
	kafkaSink  *kafkaSink
}

func (sink *podKafkaSink) Name() string {
	return POD_KAFKA_SINK
}

func (sink *podKafkaSink) Stop() {
	return
}

func (sink *podKafkaSink) ExportPodInspection(podBatch *entity.PodInspectionBatch) {
	sink.kafkaSink.Lock()
	defer sink.kafkaSink.Unlock()
	batch := &entity.PodInspectionBatch{
		TimeStamp:   podBatch.TimeStamp,
		Inspections: make([]*entity.PodInspection, 0),
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

func NewPodKafkaSink(kafkaPodConfig *config.KafkaPodConfig) (entity.PodSink, error) {
	// get kafka Client
	client, err := kafka.NewKafkaClient(&kafkaPodConfig.Kafka, kafka.PodsTopic)
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

type deleteKafkaSink kafkaSink

func NewDeleteKafkaSink(kafkaConfig *config.Kafka) (entity.DeleteSink, error) {
	client, err := kafka.NewKafkaClient(kafkaConfig, kafka.DeleteTopic)
	if err != nil {
		return nil, err
	}
	// init kafkaSink
	kafkaSink := &deleteKafkaSink{
		KafkaClient: client,
	}
	return kafkaSink, nil
}

func (sink *deleteKafkaSink) ExportDeleteInspection(deleteBatch *entity.DeleteInspectionBatch) {

	sink.Lock()
	defer sink.Unlock()
	//klog.V(0).Info(nodeBatch)
	err := sink.KafkaClient.ProduceKafkaMessage(deleteBatch)
	if err != nil {
		klog.Errorf("Failed to produce NodeInspection message: %s", err)
	}
}
