package core

import (
	"time"
)

type PodInspectionBatch struct {
	TimeStamp   time.Time        `json:"timestamp"`
	Inspections []*PodInspection `json:"inspections"`
}

type PodInspection struct {
	PodName   string `json:"pod_name"`
	PodIP     string `json:"pod_ip"`
	HostIP    string `json:"host_ip"`
	NodeName  string `json:"node_name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
}

type Namespaces []string

func (namespaces Namespaces) SkipPod(inspection *PodInspection) bool {
	if namespaces != nil {
		skip := true
		for _, namespace := range namespaces {
			if namespace == inspection.Namespace {
				skip = false
			}
			if skip {
				return true
			}
		}
	}
	return false
}

type PodInspectionSource interface {
	GetNewPodInspection() *PodInspectionBatch
}

type PodSink interface {
	Name() string
	ExportPodInspection(*PodInspectionBatch)
	Stop()
}
