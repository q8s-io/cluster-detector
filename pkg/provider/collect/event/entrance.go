package event

import (
	"github.com/q8s-io/cluster-detector/pkg/entity"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/kubernetes"

	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

const (
	// Number of object pointers.
	// Big enough so it won't be hit anytime soon with reasonable GetNewEvents frequency.
	DELETE                     = "Deleted"
	JobEvent                   = "Job"
	PodEvent                   = "Pod"
	DeploymentEvent            = "Deployment"
	ReplicationControllerEvent = "ReplicationController"
	NodeEvent                  = "Node"
	ReplicaSetEvent            = "ReplicaSet"
	DaemonSetEvent             = "DaemonSet"
	StatefulSetEvent           = "StatefulSet"
	NameSpaceEvent             = "NameSpace"
	ConfigMapEvent             = "ConfigMap"
	SecretEvent                = "Secret"
	ServiceEvent               = "Service"
	IngressEvent               = "Ingress"
	PersistentVolumeEvent      = "PersistentVolume"
	ServiceAccountEvent        = "ServiceAccount"
	ClusterRoleEvent           = "ClusterRole"
)

type Client struct {
	KubeClient  k8s.Interface
	StopChannel chan struct{}
}

var List chan *entity.EventInspection

func NewKubernetesSource() *chan *entity.EventInspection {
	List = make(chan *entity.EventInspection, entity.DefaultBufSize)
	return &List
}

func GetKubeClient() (*Client, error) {
	cfgUrl := config.Config.Source.KubernetesURL

	kubeConfig, err := kubernetes.GetKubeClientConfig(cfgUrl)
	if err != nil {
		return nil, err
	}

	kubeClient, err := kubernetes.GetKubernetesClient(kubeConfig)
	if err != nil {
		klog.Error("Failed to create kubernetes client,because of %v", err)
		return nil, err
	}
	return &Client{
		KubeClient:  kubeClient,
		StopChannel: make(chan struct{}),
	}, nil
}

func StartWatch() {
	c, err := GetKubeClient()
	if err != nil && c != nil {
		return
	}
	// Outer loop, for reconnections
	go c.normalWatch()
	go c.deleteWatch()
	go c.namespaceWatch()
	go c.configMapWatch()
	go c.secretEventWatch()
	go c.serviceEventWatch()
	go c.ingressEventWatch()
	go c.persistentVolumeEventWatch()
	go c.serviceAccountEvent()
	go c.clusterRoleEventWatch()
	select {}
}
