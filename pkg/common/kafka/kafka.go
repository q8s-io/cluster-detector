package kafka

import (
	"encoding/json"
	"fmt"
	"github.com/q8s-io/cluster-detector/configs"
	"time"

	kafka "github.com/Shopify/sarama"
	"k8s.io/klog"
)

const (
	brokerClientID         = "kafka-sink"
	brokerDialTimeout      = 10 * time.Second
	brokerDialRetryLimit   = 1
	brokerDialRetryWait    = 0
	brokerLeaderRetryLimit = 1
	brokerLeaderRetryWait  = 0
	metricsTopic           = "heapster-metrics"
	eventsTopic            = "heapster-events"
	nodesTopic             = "heapster-nodes"
	podsTopic              = "heapster-pods"
	deletesTopic		   = "heapster-deletes"
)

const (
	TimeSeriesTopic = "timeseriestopic"
	EventsTopic     = "eventstopic"
	NodesTopic      = "nodestopic"
	PodsTopic       = "podstopic"
	DeleteTopic     = "deletetopic"

)

type KafkaClient interface {
	Name() string
	Stop()
	ProduceKafkaMessage(msgData interface{}) error
}

type kafkaSink struct {
	producer  kafka.SyncProducer
	dataTopic string
}

func (sink *kafkaSink) ProduceKafkaMessage(msgData interface{}) error {
	start := time.Now()
	msgJson, err := json.Marshal(msgData)
	if err != nil {
		return fmt.Errorf("failed to transform the items to json : %s", err)
	}

	_, _, err = sink.producer.SendMessage(&kafka.ProducerMessage{
		Topic: sink.dataTopic,
		Key:   nil,
		Value: kafka.ByteEncoder(msgJson),
	})
	if err != nil {
		return fmt.Errorf("failed to produce message to %s: %s", sink.dataTopic, err)
	}
	end := time.Now()
	klog.V(4).Infof("Exported %d data to kafka in %s", len(msgJson), end.Sub(start))
	return nil
}

func (sink *kafkaSink) Name() string {
	return "Apache Kafka Sink"
}

func (sink *kafkaSink) Stop() {
	sink.producer.Close()
}

func getTopic(topic string, topicType string) (string, error) {
	if topic != "" {
		return topic, nil
	}
	switch topicType {
	case TimeSeriesTopic:
		topic = metricsTopic
	case EventsTopic:
		topic = eventsTopic
	case NodesTopic:
		topic = nodesTopic
	case PodsTopic:
		topic = podsTopic
	case DeleteTopic:
		topic = deletesTopic
	default:
		return "", fmt.Errorf("Topic type %s is illegal.", topicType)
	}
	return topic, nil
}

func NewKafkaClient(kafkaConfig *configs.Kafka, topicType string) (KafkaClient, error) {
	topic, err := getTopic(kafkaConfig.Topic, topicType)
	if err != nil {
		return nil, err
	}

	//compression, err := getCompression(opts)
	compression := kafka.CompressionNone

	kafkaBrokers := kafkaConfig.Brokers
	if len(kafkaBrokers) < 1 {
		return nil, fmt.Errorf("There is no broker assigned for connecting kafka")
	}
	klog.V(2).Infof("initializing kafka sink with brokers - %v", kafkaBrokers)

	kafka.Logger = GologAdapterLogger{}

	//structure the config of broker
	config := kafka.NewConfig()
	config.ClientID = brokerClientID
	config.Net.DialTimeout = brokerDialTimeout
	config.Metadata.Retry.Max = brokerDialRetryLimit
	config.Metadata.Retry.Backoff = brokerDialRetryWait
	config.Producer.Retry.Max = brokerLeaderRetryLimit
	config.Producer.Retry.Backoff = brokerLeaderRetryWait
	config.Producer.Compression = compression
	config.Producer.Partitioner = kafka.NewRoundRobinPartitioner
	config.Producer.RequiredAcks = kafka.WaitForLocal
	config.Producer.Return.Errors = true
	config.Producer.Return.Successes = true

	//config.Net.TLS.Config, config.Net.TLS.Enable, err = getTlsConfiguration(opts)
	config.Net.TLS.Config, config.Net.TLS.Enable = nil, false

	//config.Net.SASL.User, config.Net.SASL.Password, config.Net.SASL.Enable, err = getSASLConfiguration(opts)
	config.Net.SASL.User, config.Net.SASL.Password, config.Net.SASL.Enable = "", "", false

	// set up producer of kafka server.
	klog.V(3).Infof("attempting to setup kafka sink")
	sinkProducer, err := kafka.NewSyncProducer(kafkaBrokers, config)
	if err != nil {
		return nil, fmt.Errorf("Failed to setup Producer: - %v", err)
	}

	klog.V(3).Infof("kafka sink setup successfully")
	return &kafkaSink{
		producer:  sinkProducer,
		dataTopic: topic,
	}, nil
}
