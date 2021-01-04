package collect

import (
	"github.com/q8s-io/cluster-detector/pkg/entity"
	"github.com/q8s-io/cluster-detector/pkg/provider/collect/event"
	"github.com/q8s-io/cluster-detector/pkg/provider/collect/node"
	"github.com/q8s-io/cluster-detector/pkg/provider/collect/pod"
	"github.com/q8s-io/cluster-detector/pkg/provider/collect/release"
)

type SourceFactory struct{}

func NewSourceFactory() *SourceFactory {
	return &SourceFactory{}
}

func (_ *SourceFactory) BuildEvents() *chan *entity.EventInspection {
	return event.NewKubernetesSource()
}

func (_ *SourceFactory) BuildPods() *chan *entity.PodInspection {
	return pod.NewKubernetesSource()
}

func (_ *SourceFactory) BuildNodes() *chan *entity.NodeInspection {
	return node.NewKubernetesSource()
}

func (_ *SourceFactory) BuildDeletes() *chan *entity.DeleteInspection {
	return release.NewKubernetesSource()
}
