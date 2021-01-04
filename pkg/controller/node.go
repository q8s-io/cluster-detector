package controller

import (
	"github.com/q8s-io/cluster-detector/pkg/provider/kube/node"
)

func RunNodeInspection(){
	go node.StartWatch()
}