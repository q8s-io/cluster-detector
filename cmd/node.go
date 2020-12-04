package cmd

import (
	"github.com/q8s-io/cluster-detector/configs"
	nodecore "github.com/q8s-io/cluster-detector/pkg/core"
	m_node "github.com/q8s-io/cluster-detector/pkg/manager/node"
	"github.com/q8s-io/cluster-detector/pkg/sinks"
	sm_node "github.com/q8s-io/cluster-detector/pkg/sinks/manager/node"
	"github.com/q8s-io/cluster-detector/pkg/sources"
	"k8s.io/klog"
)

func RunNodeInspection() {
	argSource := configs.Config.Source
	argKafkaSink := &configs.Config.NodeInspectionConfig.KafkaNodeConfig
	//argWebHookSink := &configs.Config.NodeInspectionConfig.WebHookNodeConfig

	// sources
	if argSource.KubernetesURL == "" {
		klog.Fatal("Wrong sources specified")
	}
	sourceFactory := sources.NewSourceFactory()
	nodeResources, buildErr := sourceFactory.BuildNodeInspection(argSource)
	if buildErr != nil {
		klog.Fatalf("Failed to create sources: %v", buildErr)
	}
	if len(nodeResources) != 1 {
		klog.Fatal("Requires exactly 1 source")
	}

	// SinkFactory build sinks
	sinksFactory := sinks.NewSinkFactory()
	var sinkList []nodecore.NodeSink
	// kafka sink
	if argKafkaSink.Enabled == true {
		kafkaSink, kafkaErr := sinksFactory.BuildNodeKafka(argKafkaSink)
		if kafkaErr != nil {
			klog.Fatalf("Failed to create kafkaSink: %v", kafkaErr)
		}
		sinkList = append(sinkList, kafkaSink)
	}

	// WebHook sink
	/*if argWebHookSink.Enabled == true {
		webHookSink, webHookErr := sinksFactory.BuildNodeWebHook(argWebHookSink)
		if webHookErr != nil {
			klog.Fatalf("Failed to create webHookSink: %v", webHookErr)
		}
		sinkList = append(sinkList, webHookSink)
	}*/

	if len(sinkList) == 0 {
		klog.Fatal("No available sink to use")
	}
	for _, sink := range sinkList {
		klog.Infof("Starting with %s sink", sink.Name())
	}

	// SinkManager put events to sinkList
	sinkManagers, smErr := sm_node.NewNodeSinkManager(sinkList, sm_node.DefaultSinkExportNodesTimeout, sm_node.DefaultSinkStopTimeout)
	if smErr != nil {
		klog.Fatalf("Failed to create sink manager: %v", smErr)
	}

	// Main Manager
	manager, managerErr := m_node.NewManager(nodeResources[0], sinkManagers, *argFrequency)
	if managerErr != nil {
		klog.Fatalf("Failed to create main manager: %v", managerErr)
	}
	manager.Start()
	klog.Infof("Starting nodeInspection")
}
