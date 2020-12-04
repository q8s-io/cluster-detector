package cmd

import (
	"github.com/q8s-io/cluster-detector/configs"
	"github.com/q8s-io/cluster-detector/pkg/core"
	"github.com/q8s-io/cluster-detector/pkg/sinks"
	sm_delete "github.com/q8s-io/cluster-detector/pkg/sinks/manager/delete"
	"github.com/q8s-io/cluster-detector/pkg/sources"
	m_delete "github.com/q8s-io/cluster-detector/pkg/manager/delete"
	"k8s.io/klog"
)

func RunDeleteInspection(){
	argSource := configs.Config.Source
	argKafkaSink := &configs.Config.DeleteInspectionConfig.KafkaDeleteConfig
	if argSource.KubernetesURL == "" {
		klog.Fatal("Wrong sources specified")
	}
	sourceFactory := sources.NewSourceFactory()
	deleteResource,err:=sourceFactory.BuildDeleteSource(argSource)
	if err != nil {
		klog.Fatalf("Failed to create sources: %v", err)
	}
	if len(deleteResource) != 1 {
		klog.Fatal("Requires exactly 1 source")
	}
	sinksFactory := sinks.NewSinkFactory()
	var sinkList []core.DeleteSink
	if argKafkaSink.Enabled == true {
		kafkaSink, kafkaErr := sinksFactory.BuildDeleteKafka(argKafkaSink)
		if kafkaErr != nil {
			klog.Fatalf("Failed to create kafkaSink: %v", kafkaErr)
		}
		sinkList = append(sinkList, kafkaSink)
	}
	for _, sink := range sinkList {
		klog.Infof("Starting with %s sink", sink.Name())
	}
	sinkManagers, smErr := sm_delete.NewDeleteSinkManager(sinkList,sm_delete.DefaultSinkExportNodesTimeout,sm_delete.DefaultSinkStopTimeout)
	if smErr != nil {
		klog.Fatalf("Failed to create sink manager: %v", smErr)
	}
	manager, managerErr :=m_delete.NewManager(deleteResource[0], sinkManagers, *argFrequency)
	if managerErr != nil {
		klog.Fatalf("Failed to create main manager: %v", managerErr)
	}
	manager.Start()
	klog.Infof("Starting DeleteInspection")
}