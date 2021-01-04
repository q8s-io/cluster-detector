package entity

import (
	"encoding/json"
	//"github.com/q8s-io/cluster-detector/pkg/provider/collect"
	"time"
)

type EventInspection struct {
	EventNamespace    string
	EventResourceName string
	EventKind         string
	EventType         string
	EventTime         time.Time
	EventInfo         interface{}
	// EventUID		  string
}

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
	ExportEvents(buffer *chan *EventInspection)
	// Stops the sink at earliest convenience.
	Stop()
}

// Sink data type
type EventSinkPoint struct {
	EventTimestamp time.Time
	EventValue     interface{}
	EventTags      map[string]string
}

func (_ EventSinkPoint) getEventValue(buffer *EventInspection) (string, error) {
	// TODO: check whether indenting is required.
	bytes, err := json.MarshalIndent(buffer, "", " ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (this EventSinkPoint) EventToPoint(event *EventInspection) (*EventSinkPoint, error) {
	value, err := this.getEventValue(event)
	if err != nil {
		return nil, err
	}
	point := EventSinkPoint{
		EventTimestamp: event.EventTime, // event.LastTimestamp.Time.UTC(),
		EventValue:     value,
		/*EventTags: map[string]string{
			"eventID": event.EventUID,
		},*/
	}
	/*if event.EventKind == "Pod" {
		point.EventTags[core.LabelPodId.Key] = string(event.InvolvedObject.UID)
		point.EventTags[core.LabelPodName.Key] = event.InvolvedObject.Name
	}
	point.EventTags[core.LabelHostname.Key] = event.Source.Host*/
	return &point, nil
}
