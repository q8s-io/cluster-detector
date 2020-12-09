package delete

import (
	"time"

	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/q8s-io/cluster-detector/pkg/entity"
	"github.com/q8s-io/cluster-detector/pkg/manager"
)

var (
	// Last time of eventer housekeep since unix epoch in seconds
	lastHousekeepTimestamp = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "DeleteInspection",
			Subsystem: "manager",
			Name:      "last_time_seconds",
			Help:      "Last time of eventer housekeep since unix epoch in seconds.",
		})
)

func init() {
	prometheus.MustRegister(lastHousekeepTimestamp)
}

type realManager struct {
	source    entity.DeleteInspectionSource
	sink      entity.DeleteSink
	frequency time.Duration
	stopChan  chan struct{}
}

func NewManager(source entity.DeleteInspectionSource, sink entity.DeleteSink, frequency time.Duration) (manager.Manager, error) {
	manager := realManager{
		source:    source,
		sink:      sink,
		frequency: frequency,
		stopChan:  make(chan struct{}),
	}

	return &manager, nil
}

func (rm *realManager) Name() string {
	return "DeleteInspection-MainManager"
}

func (rm *realManager) Start() {
	go rm.Housekeep()
}

func (rm *realManager) Stop() {
	rm.stopChan <- struct{}{}
}

func (rm *realManager) Housekeep() {
	times := config.Config.DeleteInspectionConfig.Speed
	if times == 0 {
		klog.Fatal("Delete inspection speed is zero")
		return
	}

	updatepodProbeTimer := time.NewTicker(time.Second * time.Duration(times))
	for {
		rm.housekeep()
		<-updatepodProbeTimer.C
	}
}

func (rm *realManager) housekeep() {
	defer func() {
		lastHousekeepTimestamp.Set(float64(time.Now().Unix()))
	}()

	// No parallelism. Assumes that the events are pushed to Heapster. Add parallelism
	// when this stops to be true.
	ips := rm.source.GetNewDeleteInspection()
	klog.V(0).Infof("%s: \t Exporting %d inspection", rm.Name(), len(ips.Inspections))
	rm.sink.ExportDeleteInspection(ips)
}
