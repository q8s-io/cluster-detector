package kube

import (
	"net/url"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubev1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/entity"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/kubernetes"
)

const (
	LocalPodsSize        = 10000
	DefaultGetPodTimeout = 10
)

var timeoutThreshold = config.Config.PodInspectionConfig.TimeoutThreshold

type PodInspectionSource struct {
	// Large local buffer, periodically read.
	// channel's element size does not exceed 65535 bytes, so the pointer is used
	localPodBuffer chan *entity.PodInspection
	// count pod num
	num         int
	stopChannel chan struct{}
	podsClient  kubev1core.PodInterface
}

func (this *PodInspectionSource) GetNewPodInspection() *entity.PodInspectionBatch {
	result := entity.PodInspectionBatch{
		TimeStamp:   time.Now(),
		Inspections: []*entity.PodInspection{},
	}
	// Using channel to realize writing and reading at the same time
	go this.inspection()
	// 使用外部的条件控制读信道的结束
	count := 0
PodLoop:
	for {
		if count >= this.num {
			break PodLoop
		}
		select {
		case inspection := <-this.localPodBuffer:
			result.Inspections = append(result.Inspections, inspection)
			count++
		// 防止写入信道出现崩溃，及时退出。
		case <-time.After(time.Second * DefaultGetPodTimeout):
			break PodLoop
		}
	}
	return &result
}

func (this *PodInspectionSource) inspection() {
	podList, listErr := this.podsClient.List(metav1.ListOptions{})
	if listErr != nil {
		klog.Errorf("Failed to list Pod: %s", listErr)
	}
	this.num = len(podList.Items)
	for _, pod := range podList.Items {
		podInspection := this.filter(&pod)
		if podInspection != nil {
			this.localPodBuffer <- podInspection
		}
	}
}

func (this *PodInspectionSource) filter(pod *v1.Pod) *entity.PodInspection {
	status := getPodStatus(pod)
	if status == "Running" || status == "Succeeded" {
		return nil
	}
	if status == "pending" {
		pendingTime := time.Now().Unix() - pod.Status.StartTime.Unix()
		// default: 5 minutes
		if timeoutThreshold == 0 {
			timeoutThreshold = 300
		}
		if pendingTime < int64(timeoutThreshold) {
			return nil
		}
	}

	podInspection := new(entity.PodInspection)
	podInspection.PodName = pod.Name
	podInspection.Status = status
	podInspection.Namespace = pod.Namespace
	podInspection.NodeName = pod.Spec.NodeName
	podInspection.PodIP = pod.Status.PodIP
	podInspection.HostIP = pod.Status.HostIP

	return podInspection
}

func NewPodInspectionSource(uri *url.URL) (*PodInspectionSource, error) {
	kubeClient, err := kubernetes.GetKubernetesClient(uri)
	if err != nil {
		klog.Errorf("Failed to create kubernetes client, because of %v", err)
		return nil, err
	}
	podsClient := kubeClient.CoreV1().Pods("")
	result := PodInspectionSource{
		localPodBuffer: make(chan *entity.PodInspection, LocalPodsSize),
		stopChannel:    make(chan struct{}),
		podsClient:     podsClient,
	}
	return &result, nil
}

// GetPodStatus returns the pod state
// Phase: Pending Running Succeeded Failed Unknown
// Reason: Terminating Evicted NodeLost Error
func getPodStatus(pod *v1.Pod) string {
	// Terminating
	if pod.DeletionTimestamp != nil {
		return "Terminating"
	}

	// not running
	if pod.Status.Phase != v1.PodRunning {
		return string(pod.Status.Phase)
	}

	ready := false
	notReadyReason := ""
	for _, c := range pod.Status.Conditions {
		if c.Type == v1.PodReady {
			ready = c.Status == v1.ConditionTrue
			notReadyReason = c.Reason
		}
	}

	if pod.Status.Reason != "" {
		return pod.Status.Reason
	}

	if notReadyReason != "" {
		return notReadyReason
	}

	if ready {
		return string(v1.PodRunning)
	}

	// Unknown?
	return "Unknown"
}
