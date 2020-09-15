package configs

var Config Runtime

type Runtime struct {
	Source Source
	//Kafka                Kafka
	EventsConfig         EventsConfig
	NodeInspectionConfig NodeInspectionConfig
	PodInspectionConfig  PodInspectionConfig
}

// kubernetes url
type Source struct {
	KubernetesURL string `toml:"kubernetes_url" json:"kubernetes"`
}

// kafka config
type Kafka struct {
	Enabled bool     `toml:"enabled" json:"enabled"`
	Brokers []string `toml:"brokers" json:"brokers"`
	Topic   string   `toml:"topic" json:"topic"`
}

// WebHook config
type WebHook struct {
	Enabled    bool   `toml:"enabled" json:"enabled"`
	WebHookURL string `toml:"webhook_url" json:"webhook_url"`
}

// Events config
type EventsConfig struct {
	Enabled            bool `toml:"enabled"`
	KafkaEventConfig   KafkaEventConfig
	WebHookEventConfig WebHookEventConfig
}

// NodeInspection config
type NodeInspectionConfig struct {
	Enabled           bool `toml:"enabled"`
	Speed             int  `toml:"speed"`
	KafkaNodeConfig   Kafka
	WebHookNodeConfig WebHook
}

// PodInspection config
type PodInspectionConfig struct {
	Enabled          bool `toml:"enabled"`
	Speed            int  `toml:"speed"`
	TimeoutThreshold int  `toml:"timeout_threshold"`
	KafkaPodConfig   KafkaPodConfig
	WebHookPodConfig WebHookPodConfig
}

type PodInspectionFilter struct {
	Namespaces []string `toml:"namespaces"`
}

type KafkaPodConfig struct {
	Kafka
	PodInspectionFilter
}

type WebHookPodConfig struct {
	WebHook
	PodInspectionFilter
}

// kafkaEventSink filter
type KafkaEventConfig struct {
	Kafka
	Level      string   `toml:"level" json:"level"`
	Namespaces []string `toml:"namespaces" json:"namespaces"`
	Kinds      []string `toml:"kinds" json:"kinds"`
}

// webHookSink filter
type WebHookEventConfig struct {
	WebHook
	Level      string   `toml:"level" json:"level"`
	Namespaces []string `toml:"namespaces" json:"namespaces"`
	Kinds      []string `toml:"kinds" json:"kinds"`
	Reason     []string `toml:"reason" json:"reason"`
}