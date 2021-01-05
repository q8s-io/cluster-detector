package collect

import (
	"github.com/q8s-io/cluster-detector/pkg/provider/collect/event"
	"github.com/q8s-io/cluster-detector/pkg/provider/collect/node"
	"github.com/q8s-io/cluster-detector/pkg/provider/collect/pod"
	delete2 "github.com/q8s-io/cluster-detector/pkg/provider/collect/release/delete"
)

type SourceFactory struct{}

func NewSourceFactory() *SourceFactory {
	return &SourceFactory{}
}

func (_ *SourceFactory) InitEventChan() {
	event.NewKubernetesSource()
}

func (_ *SourceFactory) InitPodChan() {
	pod.NewKubernetesSource()
}

func (_ *SourceFactory) InitNodeChan() {
	node.NewKubernetesSource()
}

func (_ *SourceFactory) InitDeleteChan() {
	delete2.NewKubernetesSource()
}
