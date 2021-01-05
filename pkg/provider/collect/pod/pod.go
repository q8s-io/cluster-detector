package pod

import (
	"github.com/q8s-io/cluster-detector/pkg/entity"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/kubernetes"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubev1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog"
	"time"
)

var timeoutThreshold = config.Config.PodInspectionConfig.TimeoutThreshold

type PodInspectionSource struct {
	num         int
	stopChannel chan struct{}
	podsClient  kubev1core.PodInterface
}

var PodListCh chan *entity.PodInspection

func NewKubernetesSource() {
	PodListCh = make(chan *entity.PodInspection, entity.DefaultBufSize)
}

func (this *PodInspectionSource) inspection() {
	for {
		podList, listErr := this.podsClient.List(metav1.ListOptions{})
		if listErr != nil {
			klog.Errorf("Failed to list Pod: %s", listErr)
		}
		this.num = len(podList.Items)
		for _, pod := range podList.Items {
			podInspection := this.filter(&pod)
			if podInspection != nil {
				// this.localPodBuffer <- podInspection
				PodListCh <- podInspection
			}
		}
		time.Sleep(time.Second * time.Duration(config.Config.PodInspectionConfig.Speed))
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

func StartWatch() {
	podInspectionSource, err := newPodInspectionSource()
	if err != nil {
		return
	}
	podInspectionSource.inspection()
}

func newPodInspectionSource() (*PodInspectionSource, error) {
	kubeCfg, err := kubernetes.GetKubeClientConfig(config.Config.Source.KubernetesURL)
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.GetKubernetesClient(kubeCfg)
	if err != nil {
		klog.Errorf("Failed to create kubernetes client, because of %v", err)
		return nil, err
	}
	return &PodInspectionSource{
		stopChannel: make(chan struct{}),
		podsClient:  kubeClient.CoreV1().Pods(""),
	}, nil
}

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
