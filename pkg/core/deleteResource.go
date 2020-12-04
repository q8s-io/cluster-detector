package core

import (
	"time"
)


type DeleteInspectionBatch struct {
	TimeStamp   time.Time        `json:"timestamp"`
	Inspections []*DeleteInspection `json:"inspections"`
}

type DeleteInspection struct {
	Kind      string
	NameSpace string
	Name      string
}

type DeleteInspectionSource interface {
	GetNewDeleteInspection()*DeleteInspectionBatch
}

type DeleteSink interface {
	Name() string
	ExportDeleteInspection(batch *DeleteInspectionBatch)
	Stop()
}