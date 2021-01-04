package kafka

import (
	"strings"
	"sync"
	"time"

	"github.com/q8s-io/cluster-detector/pkg/entity"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/log"
)

const BufferSize = 10000

var FilterKafka = make(chan *log.Mess, BufferSize)

type skipEvent struct {
	namespaces []string
	kinds      []string
	sync.RWMutex
}

type skipPod struct {
	namespaces []string
	sync.RWMutex
}

type skipDelete struct {
	namespaces []string
	kinds      []string
	sync.RWMutex
}

type skipNode struct {
	sync.RWMutex
}

type eventsCh chan *entity.EventInspection
type podsCh chan *entity.PodInspection
type nodesCh chan *entity.NodeInspection
type deletesCh chan *entity.DeleteInspection

func Filter(source interface{}, kafkaConfig *config.Runtime) {
	switch filter := source.(type) {
	case *chan *entity.EventInspection:
		f := eventsCh(*filter)
		f.eventKafkaFilter(&kafkaConfig.EventsConfig.KafkaEventConfig)
	case *chan *entity.PodInspection:
		f := podsCh(*filter)
		f.podKafkaFilter(&kafkaConfig.PodInspectionConfig.KafkaPodConfig)
	case *chan *entity.DeleteInspection:
		f := deletesCh(*filter)
		f.deleteKafkaFilter(&kafkaConfig.DeleteInspectionConfig.KafkaDeleteConfig)
	case *chan *entity.NodeInspection:
		f := nodesCh(*filter)
		f.nodeKafkaFilter()
	}
}

func (event *eventsCh) eventKafkaFilter(kafkaConfig *config.KafkaEventConfig) {
	skips := skipEvent{
		namespaces: kafkaConfig.Namespaces,
		kinds:      kafkaConfig.Kinds,
	}
	skips.Lock()
	defer skips.Unlock()
	go func() {
		for e := range *event {
			if !skips.skip(e) {
				mes := log.Mess{
					Namespace: e.EventNamespace,
					Name:      e.EventResourceName,
					Kind:      e.EventKind,
					Type:      e.EventType,
					Time:      time.Now(),
					Info:      e.EventInfo,
				}
				log.PrintLog(mes)
				FilterKafka <- &mes
			}
		}
	}()
}

func (s *skipEvent) skip(event *entity.EventInspection) bool {
	if s.namespaces != nil && event.EventNamespace != "" {
		skip := true
		for _, namespace := range s.namespaces {
			if namespace == event.EventNamespace {
				skip = false
			}
		}
		if skip {
			return true
		}
	}
	if s.kinds != nil {
		skip := true
		for _, kind := range s.kinds {
			if strings.ToLower(kind) == strings.ToLower(event.EventKind) {
				skip = false
			}
		}
		if skip {
			return true
		}
	}
	return false
}

func (pod *podsCh) podKafkaFilter(kafkaConfig *config.KafkaPodConfig) {
	skips := skipPod{
		namespaces: kafkaConfig.Namespaces,
	}
	skips.Lock()
	defer skips.Unlock()
	go func() {
		for p := range *pod {
			if !skips.skip(p) {
				mes := log.Mess{
					Namespace: p.Namespace,
					Name:      p.PodName,
					Kind:      "Pod",
					Type:      "Pod Inspections",
					Time:      time.Now(),
					Info: map[string]interface{}{
						"podStatus": p.Status,
						"podIP":     p.PodIP,
						"hostIP":    p.HostIP,
						"nodeName":  p.NodeName,
					},
				}
				log.PrintLog(mes)
				FilterKafka <- &mes
			}
		}
	}()

}

func (s *skipPod) skip(pod *entity.PodInspection) bool {
	if s.namespaces != nil {
		skip := true
		for _, namespace := range s.namespaces {
			if namespace == pod.Namespace {
				//	fmt.Println(namespace, "==", inspection.Namespace)
				skip = false
			}
		}
		if skip {
			return true
		}
	}

	return false
}

func (delete *deletesCh) deleteKafkaFilter(kafkaConfig *config.KafkaDeleteConfig) {
	skips := skipDelete{
		namespaces: kafkaConfig.Namespaces,
		kinds:      kafkaConfig.Kinds,
		RWMutex:    sync.RWMutex{},
	}
	skips.Lock()
	defer skips.Unlock()
	go func() {
		for d := range *delete {
			if !skips.skip(d) {
				mes := log.Mess{
					Namespace: d.NameSpace,
					Name:      d.Name,
					Kind:      d.Kind,
					Type:      "GCC",
					Time:      time.Now(),
					Info:      "delete resources message",
				}
				log.PrintLog(mes)
				FilterKafka <- &mes
			}
		}
	}()
}

func (s *skipDelete) skip(delete *entity.DeleteInspection) bool {
	if s.namespaces != nil && delete.NameSpace != "" {
		skip := true
		for _, namespace := range s.namespaces {
			if namespace == delete.NameSpace {
				skip = false
			}
		}
		if skip {
			return true
		}
	}
	if s.kinds != nil {
		skip := true
		for _, kind := range s.kinds {
			if strings.ToLower(kind) == strings.ToLower(delete.Kind) {
				skip = false
			}
		}
		if skip {
			return true
		}
	}
	return false
}

func (node *nodesCh) nodeKafkaFilter() {
	skips := skipNode{}
	skips.Lock()
	defer skips.Unlock()
	go func() {
		for n := range *node {
			mes := log.Mess{
				Namespace: "",
				Name:      n.Name,
				Kind:      "Node",
				Type:      "Node Inspections",
				Time:      time.Now(),
				Info:      n.Conditions,
			}
			log.PrintLog(mes)
			FilterKafka <- &mes
		}
	}()
}
