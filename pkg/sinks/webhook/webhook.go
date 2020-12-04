package webhook

/*import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/q8s-io/cluster-detector/configs"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/q8s-io/cluster-detector/pkg/common/filters"
	"github.com/q8s-io/cluster-detector/pkg/core"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

const (
	EVENT_WEBHOOK_SINK = "EventWebHookSink"
	NODE_WEBHOOK_SINK  = "NodeWebHookSink"
	POD_WEBHOOK_SINK   = "PodWebHookSink"
	CONTENT_TYPE_JSON  = "application/json"
	Warning            = "Warning"
	Normal             = "Normal"
)

type WebHookSink struct {
	endpoint string
}

type EventWebHookSink struct {
	filters map[string]filters.Filter
	WebHookSink
}

func (ws *EventWebHookSink) Name() string {
	return EVENT_WEBHOOK_SINK
}

func (ws *EventWebHookSink) ExportEvents(batch *core.EventBatch) {
	for _, event := range batch.Events {
		err := ws.Send(event)
		if err != nil {
			klog.Warningf("Failed to send event to WebHook sink,because of %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (ws *EventWebHookSink) Stop() {
	// not implement
	return
}

// send msg to generic webHook
func (ws *EventWebHookSink) Send(event *v1.Event) (err error) {
	for _, v := range ws.filters {
		if !v.Filter(event) {
			return
		}
	}
	sinkPoint := &core.EventSinkPoint{}
	point, err := sinkPoint.EventToPoint(event)
	if err != nil {
		klog.Warningf("Failed to convert event to point: %v", err)
	}
	return send(ws.WebHookSink, point)
}

func send(ws WebHookSink, msg interface{}) error {
	point_bytes, err := json.Marshal(msg)
	if err != nil {
		klog.Warningf("failed to marshal msg %v", msg)
		return nil
	}
	b := bytes.NewBuffer(point_bytes)
	resp, respErr := http.Post(ws.endpoint, CONTENT_TYPE_JSON, b)
	if respErr != nil {
		klog.Warningf("Failed to post events to WebHook Server, because of %v", respErr)
	} else {
		defer resp.Body.Close()
	}

	if resp != nil && resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		err = fmt.Errorf("failed to send msg to sink, because the response code is %d, body is : %v", resp.StatusCode, string(body))
		klog.Errorln(err)
		return err
	}
	return nil
}

// init WebHookSink with url params
func NewEventWebHookSink(WebHookEventConfig *configs.WebHookEventConfig) (*EventWebHookSink, error) {
	s := &EventWebHookSink{
		filters: make(map[string]filters.Filter),
	}

	if WebHookEventConfig.WebHookURL != "" {
		s.endpoint = WebHookEventConfig.WebHookURL
	} else {
		klog.Errorf("uri host's length is 0 and pls check your uri: %v", WebHookEventConfig.WebHookURL)
		return nil, fmt.Errorf("uri host is not valid.url: %v", WebHookEventConfig.WebHookURL)
	}

	level := Warning
	if WebHookEventConfig.Level != "" {
		level = WebHookEventConfig.Level
		s.filters["LevelFilter"] = filters.NewGenericFilter("Type", []string{level}, false)
	}

	if WebHookEventConfig.Namespaces != nil {
		s.filters["NamespacesFilter"] = filters.NewGenericFilter("Namespace", WebHookEventConfig.Namespaces, false)
	}

	if WebHookEventConfig.Kinds != nil {
		s.filters["KindsFilter"] = filters.NewGenericFilter("Kind", WebHookEventConfig.Kinds, false)
	}

	if WebHookEventConfig.Reason != nil {
		s.filters["ReasonsFilter"] = filters.NewGenericFilter("Reason", WebHookEventConfig.Reason, true)
	}
	return s, nil
}

// NodeInspection WebHookSink definition
type NodeWebHookSink WebHookSink

func (this *NodeWebHookSink) Name() string {
	return NODE_WEBHOOK_SINK
}

func (this *NodeWebHookSink) ExportNodeInspection(batch *core.NodeInspectionBatch) {
	webHook := WebHookSink{endpoint: this.endpoint}
	err := send(webHook, batch)
	if err != nil {
		klog.Warningf("Failed to send event to WebHook sink,because of %v", err)
	}
}

func (this *NodeWebHookSink) Stop() {
	// not implement
	return
}

func NewNodeWebHookSink(webHookConfig *configs.WebHook) (*NodeWebHookSink, error) {
	s := &NodeWebHookSink{}
	if webHookConfig.WebHookURL != "" {
		s.endpoint = webHookConfig.WebHookURL
	} else {
		klog.Errorf("uri host's length is 0 and pls check your uri: %v", webHookConfig.WebHookURL)
		return nil, fmt.Errorf("uri host is not valid.url: %v", webHookConfig.WebHookURL)
	}
	return s, nil
}

// PodInspection type
type PodWebHookSink struct {
	webHookSink WebHookSink
	namespaces  core.Namespaces
}

func (*PodWebHookSink) Name() string {
	return POD_WEBHOOK_SINK
}

func (*PodWebHookSink) Stop() {
	return
}

func (sink *PodWebHookSink) ExportPodInspection(podBatch *core.PodInspectionBatch) {
	batch := &core.PodInspectionBatch{
		TimeStamp:   podBatch.TimeStamp,
		Inspections: make([]*core.PodInspection, 0),
	}
	for _, inspection := range podBatch.Inspections {
		skip := sink.namespaces.SkipPod(inspection)
		if skip {
			continue
		}
		batch.Inspections = append(batch.Inspections, inspection)
	}
	webHook := WebHookSink{endpoint: sink.webHookSink.endpoint}
	err := send(webHook, batch)
	if err != nil {
		klog.Warningf("Failed to send event to WebHook sink,because of %v", err)
	}
}

func NewPodWebHookSink(webHookPodConfig *configs.WebHookPodConfig) (*PodWebHookSink, error) {
	sink := &PodWebHookSink{}
	if webHookPodConfig.WebHookURL != "" {
		sink.webHookSink.endpoint = webHookPodConfig.WebHookURL
		sink.namespaces = webHookPodConfig.Namespaces
	} else {
		klog.Errorf("uri host's length is 0 and pls check your uri: %v", webHookPodConfig.WebHookURL)
		return nil, fmt.Errorf("uri host is not valid.url: %v", webHookPodConfig.WebHookURL)
	}
	return sink, nil
}
*/