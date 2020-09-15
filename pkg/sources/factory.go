package sources

import (
	"fmt"
	"github.com/q8s-io/cluster-detector/configs"
	"k8s.io/klog"
	"net/url"

	"github.com/q8s-io/cluster-detector/pkg/core"
	kube "github.com/q8s-io/cluster-detector/pkg/sources/kubernetes"
)

type SourceFactory struct{}

func (_ *SourceFactory) BuildEvents(source configs.Source) ([]core.EventSource, error) {
	var result []core.EventSource
	URL, parseErr := url.ParseRequestURI(source.KubernetesURL)
	if parseErr != nil {
		return nil, fmt.Errorf("Source not recognized: %s\n", parseErr)
	}
	src, err := kube.NewKubernetesSource(URL)
	if err != nil {
		klog.Errorf("Failed to create %s: %v", source, err)
	} else {
		result = append(result, src)
	}
	return result, err
}

func (_ *SourceFactory) BuildNodeInspection(source configs.Source) ([]core.NodeInspectionSource, error) {
	var result []core.NodeInspectionSource
	URL, parseErr := url.ParseRequestURI(source.KubernetesURL)
	if parseErr != nil {
		return nil, fmt.Errorf("Source not recognized: %s\n", parseErr)
	}
	src, err := kube.NewNodeInspectionSource(URL)
	if err != nil {
		klog.Errorf("Failed to create %s: %v", source, err)
	} else {
		result = append(result, src)
	}
	return result, err
}

func (_ *SourceFactory) BuildPodInspection(source configs.Source) ([]core.PodInspectionSource, error) {
	var result []core.PodInspectionSource
	URL, parseErr := url.ParseRequestURI(source.KubernetesURL)
	if parseErr != nil {
		return nil, fmt.Errorf("Source not recognized: %s\n", parseErr)
	}
	src, err := kube.NewPodInspectionSource(URL)
	if err != nil {
		klog.Errorf("Failed to create %s: %v", source, err)
	} else {
		result = append(result, src)
	}
	return result, err
}

func NewSourceFactory() *SourceFactory {
	return &SourceFactory{}
}
