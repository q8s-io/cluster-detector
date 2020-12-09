package pod

import (
	"time"

	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/entity"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/manager"
)

type realManager struct {
	source    entity.PodInspectionSource
	sink      entity.PodSink
	frequency time.Duration
	stopChan  chan struct{}
}

func NewManager(source entity.PodInspectionSource, sink entity.PodSink, frequency time.Duration) (manager.Manager, error) {
	manager := realManager{
		source:    source,
		sink:      sink,
		frequency: frequency,
		stopChan:  make(chan struct{}),
	}

	return &manager, nil
}

func (rm *realManager) Name() string {
	return "PodInspection-MainManager"
}

func (rm *realManager) Start() {
	go rm.Housekeep()
}

func (rm *realManager) Stop() {
	rm.stopChan <- struct{}{}
}

func (rm *realManager) Housekeep() {
	times := config.Config.PodInspectionConfig.Speed
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
	ips := rm.source.GetNewPodInspection()
	klog.V(0).Infof("%s: \t Exporting %d inspection", rm.Name(), len(ips.Inspections))
	rm.sink.ExportPodInspection(ips)
}
