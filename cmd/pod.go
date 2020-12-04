package cmd

import (
	"github.com/q8s-io/cluster-detector/configs"
	podcore "github.com/q8s-io/cluster-detector/pkg/core"
	m_pod "github.com/q8s-io/cluster-detector/pkg/manager/pod"
	"github.com/q8s-io/cluster-detector/pkg/sinks"
	sm_pod "github.com/q8s-io/cluster-detector/pkg/sinks/manager/pod"
	"github.com/q8s-io/cluster-detector/pkg/sources"
	"k8s.io/klog"
)

func RunPodInspection() {
	argSource := configs.Config.Source
	argKafkaSink := &configs.Config.PodInspectionConfig.KafkaPodConfig
	//argWebHookSink := &configs.Config.PodInspectionConfig.WebHookPodConfig

	// sources
	if argSource.KubernetesURL == "" {
		klog.Fatal("Wrong sources specified")
	}
	sourceFactory := sources.NewSourceFactory()
	podResources, buildErr := sourceFactory.BuildPodInspection(argSource)
	if buildErr != nil {
		klog.Fatalf("Failed to create sources: %v", buildErr)
	}
	if len(podResources) != 1 {
		klog.Fatal("Requires exactly 1 source")
	}

	// SinkFactory build sinks
	sinksFactory := sinks.NewSinkFactory()
	var sinkList []podcore.PodSink
	// kafka sink
	if argKafkaSink.Enabled == true {
		kafkaSink, kafkaErr := sinksFactory.BuildPodKafka(argKafkaSink)
		if kafkaErr != nil {
			klog.Fatalf("Failed to create kafkaSink: %v", kafkaErr)
		}
		sinkList = append(sinkList, kafkaSink)
	}

	// WebHook sink
	/*if argWebHookSink.Enabled == true{
		webHookSink, webHookErr := sinksFactory.BuildPodWebHook(argWebHookSink)
		if webHookErr != nil {
			klog.Fatalf("Failed to create kafkaSink: %v", webHookErr)
		}
		sinkList = append(sinkList, webHookSink)
	}*/

	// sink manager
	sinkManagers, smErr := sm_pod.NewPodSinkManager(sinkList, sm_pod.DefaultSinkExportPodsTimeout, sm_pod.DefaultSinkStopTimeout)
	if smErr != nil {
		klog.Fatalf("Failed to create sink manager: %v", smErr)
	}

	// Main Manager
	manager, managerErr := m_pod.NewManager(podResources[0], sinkManagers, *argFrequency)
	if managerErr != nil {
		klog.Fatalf("Failed to create main manager: %v", managerErr)
	}
	manager.Start()
	klog.Infof("Starting nodeInspection")
}
