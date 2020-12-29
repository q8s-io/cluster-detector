package filter

import (
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/provider/filter/kafka"
)

type FilterFactory struct {

}


func NewFilterFactory()*FilterFactory{
	return &FilterFactory{}
}

func (_ FilterFactory)KafkaFilter(source interface{},runtime *config.Runtime){
	kafka.Filter(source,runtime)
}