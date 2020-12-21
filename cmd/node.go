package cmd

import (
	"time"

	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/log"
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
		//log.Println(i)
		log.PrintLog(log.LogMess{
			Namespace: "",
			Name:      i.Name,
			Kind:      "Node",
			Type:      "Node Inspections",
			Time:      time.Now(),
			Info:      i.Conditions,
		})
	}
	klog.Infof("Starting nodeInspection")
}
