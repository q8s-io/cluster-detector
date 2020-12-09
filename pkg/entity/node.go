package entity

import (
	"time"

	"k8s.io/api/core/v1"
)

type NodeInspection struct {
	Name       string
	Status     string
	Conditions []v1.NodeCondition
}

type NodeInspectionBatch struct {
	Timestamp   time.Time         `json:"timestamp"`
	Inspections []*NodeInspection `json:"inspections"`
}

type NodeInspectionSource interface {
	GetNewNodeInspection() *NodeInspectionBatch
}

type NodeSink interface {
	Name() string
	ExportNodeInspection(*NodeInspectionBatch)
	Stop()
}
