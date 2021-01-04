package controller

import (
	"github.com/q8s-io/cluster-detector/pkg/provider/kube/event"
)

func RunEventsWatch() {
	go event.StartWatch()
}
