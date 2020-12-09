package node

import (
	"time"

	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/entity"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/manager"
)

type realManager struct {
	source    entity.NodeInspectionSource
	sink      entity.NodeSink
	frequency time.Duration
	stopChan  chan struct{}
}

func NewManager(source entity.NodeInspectionSource, sink entity.NodeSink, frequency time.Duration) (manager.Manager, error) {
	manager := realManager{
		source:    source,
		sink:      sink,
		frequency: frequency,
		stopChan:  make(chan struct{}),
	}

	return &manager, nil
}

func (rm *realManager) Name() string {
	return "NodeInspection-MainManager"
}

func (rm *realManager) Start() {
	go rm.Housekeep()
}

func (rm *realManager) Stop() {
	rm.stopChan <- struct{}{}
}

func (rm *realManager) Housekeep() {
	times := config.Config.NodeInspectionConfig.Speed
	if times == 0 {
		klog.Fatal("Node inspection speed is zero")
		return
	}
	updatepodProbeTimer := time.NewTicker(time.Second * time.Duration(times))
	for {
		rm.housekeep()
		<-updatepodProbeTimer.C
	}
}

func (rm *realManager) housekeep() {
	// No parallelism. Assumes that the events are pushed to Heapster. Add parallelism
	// when this stops to be true.
	ips := rm.source.GetNewNodeInspection()
	klog.V(0).Infof("%s: \t Exporting %d inspection", rm.Name(), len(ips.Inspections))
	rm.sink.ExportNodeInspection(ips)
}
