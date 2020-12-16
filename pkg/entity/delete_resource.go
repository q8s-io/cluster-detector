package entity

import (
	"time"
)

type DeleteInspection struct {
	Kind      string
	NameSpace string
	Name      string
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
