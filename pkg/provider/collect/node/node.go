package node

import (
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubev1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog"

	nodecore "github.com/q8s-io/cluster-detector/pkg/entity"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/kubernetes"
)

const (
	LocalNodesBufferSize  = 10000
	DefaultGetNodeTimeout = 10
)

type InspectionSource struct {
	// Large local buffer, periodically read.
	// channel's element size does not exceed 65535 bytes, so the pointer is used
	// localNodeBuffer chan *nodecore.NodeInspection
	// count node num
	num         int
	stopChannel chan struct{}
	nodeClient  kubev1core.NodeInterface
}

var NodeListCh chan *nodecore.NodeInspection

func NewKubernetesSource() {
	NodeListCh = make(chan *nodecore.NodeInspection, LocalNodesBufferSize)
}

func (this *InspectionSource) inspection() {
	for {
		nodeList, listErr := this.nodeClient.List(metav1.ListOptions{})
		if listErr != nil {
			klog.Errorf("Failed to list Node: %s", listErr)
		}
		this.num = len(nodeList.Items)
		for _, node := range nodeList.Items {
			inspection := new(nodecore.NodeInspection)
			inspection.Name = node.Name
			isReady := false
			for _, condition := range node.Status.Conditions {
				if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
					isReady = true
				}
			}
			if isReady {
				inspection.Status = "Ready"
			} else {
				inspection.Status = "NotReady"
			}
			inspection.Conditions = node.Status.Conditions
			NodeListCh <- inspection
			//	this.localNodeBuffer <- inspection
		}
		time.Sleep(time.Second * time.Duration(config.Config.NodeInspectionConfig.Speed))
	}
}

func newNodeInspectionSource() (*InspectionSource, error) {
	kubeCfg, err := kubernetes.GetKubeClientConfig(config.Config.Source.KubernetesURL)
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.GetKubernetesClient(kubeCfg)
	if err != nil {
		klog.Errorf("Failed to create kubernetes client, because of %v", err)
		return nil, err
	}
	return &InspectionSource{
		stopChannel: make(chan struct{}),
		nodeClient:  kubeClient.CoreV1().Nodes(),
	}, nil
}

func StartWatch() {
	nodeInspectionSource, err := newNodeInspectionSource()
	if err != nil {
		return
	}
	nodeInspectionSource.inspection()
}
