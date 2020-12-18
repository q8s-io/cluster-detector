package cmd

import (
	"k8s.io/klog"
	"log"

	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/provider/kube"
)

func RunNodeInspection() {
	argSource := config.Config.Source
	if argSource.KubernetesURL == "" {
		klog.Fatal("Wrong sources specified")
	}
	sourceFactory := kube.NewSourceFactory()
	nodeResources, buildErr := sourceFactory.BuildNodeInspection(argSource)
	if buildErr != nil {
		klog.Fatalf("Failed to create sources: %v", buildErr)
	}
	for i := range *nodeResources {
		log.Println(i)
	}
	klog.Infof("Starting nodeInspection")
}
