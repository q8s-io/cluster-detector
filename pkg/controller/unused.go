package controller

import (
	"github.com/q8s-io/cluster-detector/pkg/provider/kube"
)

func RunUnusedInspection(){
	go kube.StartWatch()
}
