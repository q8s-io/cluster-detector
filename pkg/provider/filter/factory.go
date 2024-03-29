package filter

import (
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/provider/filter/kafka"
)

type Factory struct {
}

func NewFilterFactory() *Factory {
	return &Factory{}
}

func (_ Factory) KafkaFilter(sourceType string, runtime *config.Runtime) {
	kafka.Filter(sourceType, runtime)
}
