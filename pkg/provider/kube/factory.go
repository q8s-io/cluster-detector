package kube

import (
	"fmt"
	"net/url"

	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/entity"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
)

type SourceFactory struct{}

func NewSourceFactory() *SourceFactory {
	return &SourceFactory{}
}

func (_ *SourceFactory) BuildEvents(source config.Source) (*chan *LocalEventBuffer, error) {
	requestURI, parseErr := url.ParseRequestURI(source.KubernetesURL)
	if parseErr != nil {
		return nil, fmt.Errorf("Source not recognized: %s\n", parseErr)
	}
	src, err := NewKubernetesSource(requestURI)
	if err != nil {
		klog.Info("Failed to create %s: %v", source, err)
	}
	return src, err
}

func (_ *SourceFactory) BuildUnusedResource(source config.Source) ([]entity.DeleteInspectionSource, error) {
	var result []entity.DeleteInspectionSource
	requestURI, parseErr := url.ParseRequestURI(source.KubernetesURL)
	if parseErr != nil {
		return nil, fmt.Errorf("Source not recognized: %s\n", parseErr)
	}
	src, err := NewDeleteInspectionSource(requestURI)
	if err != nil {
		klog.Info("Failed to create %s: %v", source, err)
	} else {
		result = append(result, src)
	}
	return result, err
}

func (_ *SourceFactory) BuildNodeInspection(source config.Source) ([]entity.NodeInspectionSource, error) {
	var result []entity.NodeInspectionSource
	requestURI, parseErr := url.ParseRequestURI(source.KubernetesURL)
	if parseErr != nil {
		return nil, fmt.Errorf("Source not recognized: %s\n", parseErr)
	}
	src, err := NewNodeInspectionSource(requestURI)
	if err != nil {
		klog.Info("Failed to create %s: %v", source, err)
	} else {
		result = append(result, src)
	}
	return result, err
}

func (_ *SourceFactory) BuildPodInspection(source config.Source) ([]entity.PodInspectionSource, error) {
	var result []entity.PodInspectionSource
	requestURI, parseErr := url.ParseRequestURI(source.KubernetesURL)
	if parseErr != nil {
		return nil, fmt.Errorf("Source not recognized: %s\n", parseErr)
	}
	src, err := NewPodInspectionSource(requestURI)
	if err != nil {
		klog.Info("Failed to create %s: %v", source, err)
	} else {
		result = append(result, src)
	}
	return result, err
}
