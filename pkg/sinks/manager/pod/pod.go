package pod

import (
	"sync"
	"time"

	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/entity"
)

const (
	DefaultSinkExportPodsTimeout = 20 * time.Second
	DefaultSinkStopTimeout       = 60 * time.Second
	SINK_MANAGER_NAME            = "PodInspection-Manager"
)

type sinkHolder struct {
	sink            entity.PodSink
	podBatchChannel chan *entity.PodInspectionBatch
	stopChannel     chan bool
}

// Sink Manager - a special sink that distributes data to other sinks. It pushes data
// only to these sinks that completed their previous exports. Data that could not be
// pushed in the defined time is dropped and not retried.
type sinkManager struct {
	sinkHolders        []sinkHolder
	exportNodesTimeout time.Duration
	// Should be larger than exportEventsTimeout, although it is not a hard requirement.
	stopTimeout time.Duration
}


// Guarantees that the export will complete in exportEventsTimeout.
func (this *sinkManager) ExportPodInspection(data *entity.PodInspectionBatch) {
	var wg sync.WaitGroup
	for _, sh := range this.sinkHolders {
		wg.Add(1)
		go func(sh sinkHolder, wg *sync.WaitGroup) {
			defer wg.Done()
			klog.V(2).Infof("%s: Pushing nodes to: %s", SINK_MANAGER_NAME, sh.sink.Name())
			select {
			case sh.podBatchChannel <- data:
				klog.V(2).Infof("%s: Data nodes completed: %s", SINK_MANAGER_NAME, sh.sink.Name())
				// everything ok
			case <-time.After(this.exportNodesTimeout):
				klog.Warningf("%s: Failed to nodes data to sink: %s", SINK_MANAGER_NAME, sh.sink.Name())
			}
		}(sh, &wg)
	}
	// Wait for all pushes to complete or timeout.
	wg.Wait()
}

func (this *sinkManager) Name() string {
	return SINK_MANAGER_NAME
}

func (this *sinkManager) Stop() {
	for _, sh := range this.sinkHolders {
		klog.V(2).Infof("Running stop for: %s", sh.sink.Name())

		go func(sh sinkHolder) {
			select {
			case sh.stopChannel <- true:
				// everything ok
				klog.V(2).Infof("Stop sent to sink: %s", sh.sink.Name())
			case <-time.After(this.stopTimeout):
				klog.Warningf("Failed to stop sink: %s", sh.sink.Name())
			}
			return
		}(sh)
	}
}

/*func export(s entity.PodSink, data *entity.PodInspectionBatch) {
	s.ExportPodInspection(data)
}*/
