package kube

import (
	"github.com/q8s-io/cluster-detector/pkg/entity"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/provider/kube/event"
)

type SourceFactory struct{}

func NewSourceFactory() *SourceFactory {
	return &SourceFactory{}
}

func (_ *SourceFactory) BuildEvents(source config.Source) *chan *entity.EventInspection {
	return event.NewKubernetesSource()
}

// func (_ *SourceFactory) BuildUnusedResource(source config.Source) (*chan *entity.DeleteInspection, error) {
// 	//	var result []entity.DeleteInspectionSource
// 	requestURI, parseErr := url.ParseRequestURI(source.KubernetesURL)
// 	if parseErr != nil {
// 		return nil, fmt.Errorf("Source not recognized: %s\n", parseErr)
// 	}
// 	src, err := NewDeleteInspectionSource(requestURI)
// 	if err != nil {
// 		klog.Info("Failed to create %s: %v", source, err)
// 	} /* else {
// 		result = append(result, src)
// 	}*/
// 	return src, err
// }
//
// func (_ *SourceFactory) BuildNodeInspection(source config.Source) (*chan *entity.NodeInspection, error) {
// 	// var result []entity.NodeInspectionSource
// 	requestURI, parseErr := url.ParseRequestURI(source.KubernetesURL)
// 	if parseErr != nil {
// 		return nil, fmt.Errorf("Source not recognized: %s\n", parseErr)
// 	}
// 	src, err := node.NewNodeInspectionSource(requestURI)
// 	if err != nil {
// 		klog.Info("Failed to create %s: %v", source, err)
// 	} /*else {
// 		result = append(result, src)
// 	}*/
// 	return src, err
// }
//
// func (_ *SourceFactory) BuildPodInspection(source config.Source) (*chan *entity.PodInspection, error) {
// 	// var result []entity.PodInspectionSource
// 	requestURI, parseErr := url.ParseRequestURI(source.KubernetesURL)
// 	if parseErr != nil {
// 		return nil, fmt.Errorf("Source not recognized: %s\n", parseErr)
// 	}
// 	src, err := pod.NewPodInspectionSource(requestURI)
// 	if err != nil {
// 		klog.Info("Failed to create %s: %v", source, err)
// 	} /*else {
// 		result = append(result, src)
// 	}*/
// 	return src, err
// }
