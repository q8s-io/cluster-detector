package event

import (
	"time"

	"github.com/q8s-io/cluster-detector/pkg/entity"
	"github.com/q8s-io/cluster-detector/pkg/manager"
)

type realManager struct {
	source    entity.EventSource
	sink      entity.EventSink
	frequency time.Duration
	stopChan  chan struct{}
}

func NewManager(source entity.EventSource, sink entity.EventSink, frequency time.Duration) (manager.Manager, error) {
	manager := realManager{
		source:    source,
		sink:      sink,
		frequency: frequency,
		stopChan:  make(chan struct{}),
	}

	return &manager, nil
}

func (rm *realManager) Name() string {
	return "Event-MainManager"
}

func (rm *realManager) Start() {
	go rm.Housekeep()
}

func (rm *realManager) Stop() {
	rm.stopChan <- struct{}{}
}

func (rm *realManager) Housekeep() {
	for {
		// Try to invoke housekeep at fixed time.
		now := time.Now()
		start := now.Truncate(rm.frequency)
		end := start.Add(rm.frequency)
		timeToNextSync := end.Sub(now)

		select {
		case <-time.After(timeToNextSync):
			rm.housekeep()
		case <-rm.stopChan:
			rm.sink.Stop()
			return
		}
	}
}

func (rm *realManager) housekeep() {
	// No parallelism.
	// Assumes that the events are pushed to Heapster.
	// Add parallelism when this stops to be true.
	//events := rm.source.GetNewEvents()
	//if len(events.Events) > 0 {
	//	klog.V(0).Infof("%s: \t Exporting %d events", rm.Name(), len(events.Events))
	//}
	//rm.sink.ExportEvents(events)
}
