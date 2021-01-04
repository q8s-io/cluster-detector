package controller

import (
	"github.com/q8s-io/cluster-detector/pkg/provider/kube/pod"
)

func RunPodInspection(){
	go pod.StartWatch()
}

