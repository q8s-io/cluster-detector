package core

import (
	kubeapi "k8s.io/api/core/v1"
	"time"
)

type NodeInspectionBatch struct {
	Timestamp   time.Time         `json:"timestamp"`
	Inspections []*NodeInspection `json:"inspections"`
}

type NodeInspection struct {
	Name       string
	Status     string
	Conditions []kubeapi.NodeCondition
}

type NodeInspectionSource interface {
	GetNewNodeInspection() *NodeInspectionBatch
}

type NodeSink interface {
	Name() string
	ExportNodeInspection(*NodeInspectionBatch)
	Stop()
}
