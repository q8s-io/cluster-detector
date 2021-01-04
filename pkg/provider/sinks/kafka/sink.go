package kafka

import (
	filter "github.com/q8s-io/cluster-detector/pkg/provider/filter/kafka"
)

func Sink() {
	defer close(filter.FilterKafka)
	for {
		select {
		case <-filter.FilterKafka:
		default:
		}
	}
}
