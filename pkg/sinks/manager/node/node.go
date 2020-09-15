// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package node

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/q8s-io/cluster-detector/pkg/core"
	"k8s.io/klog"
)

const (
	DefaultSinkExportNodesTimeout = 20 * time.Second
	DefaultSinkStopTimeout        = 60 * time.Second
	SINK_MANAGER_NAME             = "NodeInspection-Manager"
)

var (
	// Time spent exporting events to sink in milliseconds.
	exporterDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "nodeInspection",
			Subsystem: "exporter",
			Name:      "duration_milliseconds",
			Help:      "Time spent exporting events to sink in milliseconds.",
		},
		[]string{"exporter"},
	)
)

func init() {
	prometheus.MustRegister(exporterDuration)
}

type sinkHolder struct {
	sink             core.NodeSink
	nodeBatchChannel chan *core.NodeInspectionBatch
	stopChannel      chan bool
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

func NewNodeSinkManager(sinks []core.NodeSink, exportEventsTimeout, stopTimeout time.Duration) (core.NodeSink, error) {
	var sinkHolders []sinkHolder
	for _, sink := range sinks {
		sh := sinkHolder{
			sink:             sink,
			nodeBatchChannel: make(chan *core.NodeInspectionBatch),
			stopChannel:      make(chan bool),
		}
		sinkHolders = append(sinkHolders, sh)
		go func(sh sinkHolder) {
			for {
				select {
				case data := <-sh.nodeBatchChannel:
					export(sh.sink, data)
				case isStop := <-sh.stopChannel:
					klog.V(2).Infof("Stop received: %s", sh.sink.Name())
					if isStop {
						sh.sink.Stop()
						return
					}
				}
			}
		}(sh)
	}
	return &sinkManager{
		sinkHolders:        sinkHolders,
		exportNodesTimeout: exportEventsTimeout,
		stopTimeout:        stopTimeout,
	}, nil
}

// Guarantees that the export will complete in exportEventsTimeout.
func (this *sinkManager) ExportNodeInspection(data *core.NodeInspectionBatch) {
	var wg sync.WaitGroup
	for _, sh := range this.sinkHolders {
		wg.Add(1)
		go func(sh sinkHolder, wg *sync.WaitGroup) {
			defer wg.Done()
			klog.V(2).Infof("%s: Pushing nodes to: %s", SINK_MANAGER_NAME, sh.sink.Name())
			select {
			case sh.nodeBatchChannel <- data:
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

func export(s core.NodeSink, data *core.NodeInspectionBatch) {
	startTime := time.Now()
	defer func() {
		exporterDuration.
			WithLabelValues(s.Name()).
			Observe(float64(time.Since(startTime)) / float64(time.Millisecond))
	}()
	s.ExportNodeInspection(data)
}
