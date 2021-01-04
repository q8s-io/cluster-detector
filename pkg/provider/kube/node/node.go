package node

// import (
// 	"time"
//
// 	v1 "k8s.io/api/core/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	kubev1core "k8s.io/client-go/kubernetes/typed/core/v1"
// 	"k8s.io/klog"
//
// 	nodecore "github.com/q8s-io/cluster-detector/pkg/entity"
// 	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
// )
//
// const (
// 	LocalNodesBufferSize  = 10000
// 	DefaultGetNodeTimeout = 10
// )
//
// type NodeInspectionSource struct {
// 	// Large local buffer, periodically read.
// 	// channel's element size does not exceed 65535 bytes, so the pointer is used
// 	localNodeBuffer chan *nodecore.NodeInspection
// 	// count node num
// 	num         int
// 	stopChannel chan struct{}
// 	nodeClient  kubev1core.NodeInterface
// }
//
// var NodeList chan *nodecore.NodeInspection
//
// func (this *NodeInspectionSource) GetNewNodeInspection() *nodecore.NodeInspectionBatch {
// 	result := nodecore.NodeInspectionBatch{
// 		Timestamp:   time.Now(),
// 		Inspections: []*nodecore.NodeInspection{},
// 	}
// 	// Using channel to realize writing and reading at the same time
// 	go this.inspection()
// 	// 使用外部的条件控制读取操作的结束
// 	count := 0
// NodeLoop:
// 	for {
// 		if count >= this.num {
// 			break NodeLoop
// 		}
// 		select {
// 		case inspection := <-this.localNodeBuffer:
// 			result.Inspections = append(result.Inspections, inspection)
// 			count++
// 		// 防止写入信道出现崩溃，及时退出。
// 		case <-time.After(time.Second * DefaultGetNodeTimeout):
// 			break NodeLoop
// 		}
// 	}
// 	return &result
// }
//
// func (this *NodeInspectionSource) inspection() {
// 	for {
// 		nodeList, listErr := this.nodeClient.List(metav1.ListOptions{})
// 		if listErr != nil {
// 			klog.Errorf("Failed to list Node: %s", listErr)
// 		}
// 		this.num = len(nodeList.Items)
// 		for _, node := range nodeList.Items {
// 			inspection := new(nodecore.NodeInspection)
// 			inspection.Name = node.Name
// 			isReady := false
// 			for _, condition := range node.Status.Conditions {
// 				if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
// 					isReady = true
// 				}
// 			}
// 			if isReady {
// 				inspection.Status = "Ready"
// 			} else {
// 				inspection.Status = "NotReady"
// 			}
// 			inspection.Conditions = node.Status.Conditions
// 			NodeList <- inspection
// 			//	this.localNodeBuffer <- inspection
// 		}
// 		time.Sleep(time.Second * time.Duration(config.Config.NodeInspectionConfig.Speed))
// 	}
// }

//
// func NewNodeInspectionSource(uri *url.URL) (*chan *nodecore.NodeInspection, error) {
// 	kubeClient, err := kubernetes.GetKubernetesClient(uri)
// 	NodeList = make(chan *nodecore.NodeInspection, LocalNodesBufferSize)
// 	if err != nil {
// 		klog.Errorf("Failed to create kubernetes client, because of %v", err)
// 		return nil, err
// 	}
// 	nodeClient := kubeClient.CoreV1().Nodes()
// 	result := NodeInspectionSource{
// 		localNodeBuffer: make(chan *nodecore.NodeInspection, LocalNodesBufferSize),
// 		stopChannel:     make(chan struct{}),
// 		nodeClient:      nodeClient,
// 	}
// 	go result.inspection()
// 	return &NodeList, nil
// }
