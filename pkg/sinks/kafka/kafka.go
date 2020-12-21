package kafka

import (
	"strings"
	"sync"

	// "k8s.io/api/core/v1"
	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/entity"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/kafka"
)

const (
	LocalBufferSize       = 100000
	EVENT_KAFKA_SINK      = "EventKafkaSink"
	NODE_KAFKA_SINK       = "NodeKafkaSink"
	POD_KAFKA_SINK        = "PodKafkaSink"
	DELETE_KAFKA_SINK     = "DeleteKafkaSink"
	WARNING           int = 2
	NORMAL            int = 1
)

var (
	KafkaEventInspection  chan *entity.EventInspection
	KafkaPodInspection    chan *entity.PodInspection
	KafkaDeleteInspection chan *entity.DeleteInspection
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
func (sink *eventKafkaSink) skipEvent(buffer *entity.EventInspection) bool {
	if sink.namespaces != nil && buffer.EventNamespace != "" {
		skip := true
		for _, namespace := range sink.namespaces {
			if namespace == buffer.EventNamespace {
				skip = false
			}
		}
		if skip {
			return true
		}
	}
	if sink.kinds != nil {
		skip := true
		for _, kind := range sink.kinds {
			if strings.ToLower(kind) == strings.ToLower(buffer.EventKind) {
				skip = false
			}
		}
		if skip {
			return true
		}
	}
	return false
}

func (sink *eventKafkaSink) ExportEvents(eventList *chan *entity.EventInspection) {
	sink.kafkaSink.Lock()
	defer sink.kafkaSink.Unlock()
	KafkaEventInspection = make(chan *entity.EventInspection, LocalBufferSize)
	go func() {
		for event := range *eventList {
			if !sink.skipEvent(event) {
				KafkaEventInspection <- event
			}
		}
	}()
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
	//eventKafkaSink.level = tool.GetLevel(kafkaEventConfig.Level)
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

func (sink *podKafkaSink) ExportPodInspection(podBatch *chan *entity.PodInspection) {
	sink.kafkaSink.Lock()
	defer sink.kafkaSink.Unlock()
	KafkaPodInspection = make(chan *entity.PodInspection, LocalBufferSize)
	go func() {
		for inspection := range *podBatch {
			skip := sink.namespaces.SkipPod(inspection)
			//fmt.Println(skip,"  ",inspection.Namespace)
			if !skip {
				KafkaPodInspection <- inspection
			}
		}
	}()
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

type deleteKafkaSink struct {
	level      int
	namespaces []string
	kinds      []string
	kafkaSink  *kafkaSink
}

func (sink *deleteKafkaSink) Name() string {
	return DELETE_KAFKA_SINK
}

func (sink *deleteKafkaSink) Stop() {
	return
}

func (sink *deleteKafkaSink) ExportDeleteInspection(deleteList *chan *entity.DeleteInspection) {
	sink.kafkaSink.Lock()
	defer sink.kafkaSink.Unlock()
	KafkaDeleteInspection = make(chan *entity.DeleteInspection, LocalBufferSize)
	go func() {
		for event := range *deleteList {
			if !sink.skipSource(event) {
				//	fmt.Println(event.NameSpace, "/", event.Kind,"/",event.Name)
				KafkaDeleteInspection <- event
			}
		}
	}()
}

func (sink *deleteKafkaSink) skipSource(buffer *entity.DeleteInspection) bool {
	//fmt.Println(buffer)
	if sink.namespaces != nil && buffer.NameSpace != "" {
		skip := true
		for _, namespace := range sink.namespaces {
			if namespace == buffer.NameSpace {
				skip = false
			}
		}
		if skip {
			return true
		}
	}
	if sink.kinds != nil {
		skip := true
		for _, kind := range sink.kinds {
			if strings.ToLower(kind) == strings.ToLower(buffer.Kind) {
				skip = false
			}
		}
		if skip {
			return true
		}
	}
	return false
}

func NewDeleteKafkaSink(kafkaConfig *config.KafkaDeleteConfig) (entity.DeleteSink, error) {
	client, err := kafka.NewKafkaClient(&kafkaConfig.Kafka, kafka.EventsTopic)
	if err != nil {
		return nil, err
	}

	// init kafkaSink
	sink := &kafkaSink{
		KafkaClient: client,
		RWMutex:     sync.RWMutex{},
	}
	deleteKafkaSink := &deleteKafkaSink{
		level:     WARNING,
		kafkaSink: sink,
	}
	//eventKafkaSink.level = tool.GetLevel(kafkaEventConfig.Level)
	deleteKafkaSink.namespaces = kafkaConfig.Namespaces
	deleteKafkaSink.kinds = kafkaConfig.Kinds
	//fmt.Println(deleteKafkaSink.namespaces)
	//fmt.Println(deleteKafkaSink.kinds)
	return deleteKafkaSink, nil
}

/*func (sink *deleteKafkaSink) ExportDeleteInspection(deleteBatch *entity.DeleteInspectionBatch) {

	sink.Lock()
	defer sink.Unlock()
	//klog.V(0).Info(nodeBatch)
	err := sink.KafkaClient.ProduceKafkaMessage(deleteBatch)
	if err != nil {
		klog.Errorf("Failed to produce NodeInspection message: %s", err)
	}
}*/
