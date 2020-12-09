package entity

import (
	"encoding/json"
	"time"

	"github.com/q8s-io/cluster-detector/pkg/metrics/core"

	"k8s.io/api/core/v1"
)

type EventBatch struct {
	// List of events included in the batch.
	Events []interface{}
}

// A place from where the events should be scraped.
type EventSource interface {
	// This is a mutable method.
	// Each call to this method clears the internal buffer so that each event can be obtained only once.
	GetNewEvents() *EventBatch
}

type EventSink interface {
	Name() string
	// Exports data to the external storage.
	// The function should be synchronous/blocking and finish only after the given EventBatch was written.
	// This will allow sink manager to push data only to these sinks that finished writing the previous data.
	ExportEvents(*EventBatch)
	// Stops the sink at earliest convenience.
	Stop()
}

// Sink data type
type EventSinkPoint struct {
	EventTimestamp time.Time
	EventValue     interface{}
	EventTags      map[string]string
}

func (_ EventSinkPoint) getEventValue(event *v1.Event) (string, error) {
	// TODO: check whether indenting is required.
	bytes, err := json.MarshalIndent(event, "", " ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (this EventSinkPoint) EventToPoint(event *v1.Event) (*EventSinkPoint, error) {
	value, err := this.getEventValue(event)
	if err != nil {
		return nil, err
	}
	point := EventSinkPoint{
		EventTimestamp: event.LastTimestamp.Time.UTC(),
		EventValue:     value,
		EventTags: map[string]string{
			"eventID": string(event.UID),
		},
	}
	if event.InvolvedObject.Kind == "Pod" {
		point.EventTags[core.LabelPodId.Key] = string(event.InvolvedObject.UID)
		point.EventTags[core.LabelPodName.Key] = event.InvolvedObject.Name
	}
	point.EventTags[core.LabelHostname.Key] = event.Source.Host
	return &point, nil
}
