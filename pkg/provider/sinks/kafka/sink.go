package kafka

import (
	filter "github.com/q8s-io/cluster-detector/pkg/provider/filter/kafka"
)

func KafkaSink(){
	defer close(filter.FilterKafka)
		for {
			select {
			case <-filter.FilterKafka:
			default:
			}
		}
}
