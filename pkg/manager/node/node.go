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
	"github.com/q8s-io/cluster-detector/configs"
	"k8s.io/klog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/q8s-io/cluster-detector/pkg/core"
	"github.com/q8s-io/cluster-detector/pkg/manager"
)

var (
	// Last time of eventer housekeep since unix epoch in seconds
	lastHousekeepTimestamp = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "nodeInspection",
			Subsystem: "manager",
			Name:      "last_time_seconds",
			Help:      "Last time of eventer housekeep since unix epoch in seconds.",
		})

	// Time of latest scrape operation
	LatestScrapeTime = time.Now()
)

func init() {
	prometheus.MustRegister(lastHousekeepTimestamp)
}

type realManager struct {
	source    core.NodeInspectionSource
	sink      core.NodeSink
	frequency time.Duration
	stopChan  chan struct{}
}

func NewManager(source core.NodeInspectionSource, sink core.NodeSink, frequency time.Duration) (manager.Manager, error) {
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
	times := configs.Config.NodeInspectionConfig.Speed
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
	defer func() {
		lastHousekeepTimestamp.Set(float64(time.Now().Unix()))
	}()

	LatestScrapeTime = time.Now()

	// No parallelism. Assumes that the events are pushed to Heapster. Add parallelism
	// when this stops to be true.
	ips := rm.source.GetNewNodeInspection()
	klog.V(0).Infof("%s: \t Exporting %d inspection", rm.Name(), len(ips.Inspections))
	rm.sink.ExportNodeInspection(ips)
}
