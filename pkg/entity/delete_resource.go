package entity

import (
	"time"
)

type DeleteInspection struct {
	NameSpace string
	Name      string
	Kind      string
}

type DeleteInspectionBatch struct {
	TimeStamp   time.Time           `json:"timestamp"`
	Inspections []*DeleteInspection `json:"inspections"`
}

type DeleteInspectionSource interface {
	GetNewDeleteInspection() *DeleteInspectionBatch
}

type DeleteSink interface {
	Name() string
	ExportDeleteInspection(buffer *chan *DeleteInspection)
	Stop()
}
