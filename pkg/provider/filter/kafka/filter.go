package kafka

import (
	"strings"
	"sync"
	"time"

	"github.com/q8s-io/cluster-detector/pkg/entity"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/log"
)

const KafkaBufferSize = 10000

var FilterKafka = make(chan *log.LogMess, KafkaBufferSize)

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

type events chan *entity.EventInspection

type pods chan *entity.PodInspection

type nodes chan *entity.NodeInspection

type deletes chan *entity.DeleteInspection

func Filter(source interface{}, kafkaConfig *config.Runtime) {
	switch filter := source.(type) {
	case *chan *entity.EventInspection:
		f := events(*filter)
		f.eventKafkaFilter(&kafkaConfig.EventsConfig.KafkaEventConfig)
	case *chan *entity.PodInspection:
		f := pods(*filter)
		f.podKafkaFilter(&kafkaConfig.PodInspectionConfig.KafkaPodConfig)
	case *chan *entity.DeleteInspection:
		f := deletes((*filter))
		f.deleteKafkaFilter(&kafkaConfig.DeleteInspectionConfig.KafkaDeleteConfig)
	case *chan *entity.NodeInspection:
		f := nodes(*filter)
		f.nodeKafkaFilter()
	}
}

func (event *events) eventKafkaFilter(kafkaConfig *config.KafkaEventConfig) {
	skips := skipEvent{
		namespaces: kafkaConfig.Namespaces,
		kinds:      kafkaConfig.Kinds,
	}
	skips.Lock()
	defer skips.Unlock()
	go func() {
		for e := range *event {
			if !skips.skip(e) {
				mes := log.LogMess{
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

func (pod *pods) podKafkaFilter(kafkaConfig *config.KafkaPodConfig) {
	skips := skipPod{
		namespaces: kafkaConfig.Namespaces,
	}
	skips.Lock()
	defer skips.Unlock()
	go func() {
		for p := range *pod {
			if !skips.skip(p) {
				mes := log.LogMess{
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

func (delete *deletes) deleteKafkaFilter(kafkaConfig *config.KafkaDeleteConfig) {
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
				mes := log.LogMess{
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

func (node *nodes) nodeKafkaFilter() {
	skips := skipNode{}
	skips.Lock()
	defer skips.Unlock()
	go func() {
		for n := range *node {
			mes := log.LogMess{
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
