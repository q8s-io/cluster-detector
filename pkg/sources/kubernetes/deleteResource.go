package kubernetes

import (
	deletecore "github.com/q8s-io/cluster-detector/pkg/core"
	resource "github.com/q8s-io/cluster-detector/pkg/sources/kubernetes/deleteresource"
	"github.com/q8s-io/cluster-detector/pkg/sources/kubernetes/determiner"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	cliresource "k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"log"
	"net/url"
	k8s "github.com/q8s-io/cluster-detector/pkg/common/kubernetes"
	//"strings"
	"context"
	"time"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

const (
	LocalDeleteBufferSize  = 10000
	DefaultGetDeleteTimeout = 10
)

type DeleteInspectionSource struct {
	// Large local buffer, periodically read.
	// channel's element size does not exceed 65535 bytes, so the pointer is used
	localNodeBuffer chan *deletecore.DeleteInspection
	// count node num
	num         int
	stopChannel chan struct{}
	clientset     kubernetes.Interface
	dynamicClient dynamic.Interface
}

type runner struct {
	namespace        string
	allNamespaces    bool
	timeout          time.Duration
	deleteOpts *metav1.DeleteOptions
	determiner    determiner.Determiner
	dynamicClient dynamic.Interface
	printer       printers.ResourcePrinter
	result        *cliresource.Result
}

func newRunner() []*deletecore.DeleteInspection {
	runner:= &runner{
		allNamespaces: true,
	}
	f := cmdutil.NewFactory(genericclioptions.NewConfigFlags(true))
	runner.Complete(f)
	dels,err:=runner.Run(context.Background(), f)
	if err!=nil{
		log.Fatalf("get deleteInspections error err:%v\n",err.Error())
	}
	return dels
}

func (r *runner) Complete(f cmdutil.Factory) (err error) {
	r.deleteOpts = &metav1.DeleteOptions{}
	r.namespace=metav1.NamespaceAll
	allResources := "pod,cm,secret,PersistentVolume,PersistentVolumeClaim,Job,PodDisruptionBudget,HorizontalPodAutoscaler"
	if err = r.completeResources(f, allResources); err != nil {
		return
	}
	//fmt.Println(allResources)
	clientset, err := f.KubernetesClientSet()
	if err != nil {
		return
	}

	r.dynamicClient, err = f.DynamicClient()
	if err != nil {
		return
	}
	resourceClient := resource.NewClient(clientset, r.dynamicClient)

	//discoveryClient, err := f.ToDiscoveryClient()
	if err != nil {
		return err
	}
	//r.dryRunVerifier = cliresource.NewDryRunVerifier(r.dynamicClient, discoveryClient)

	namespace := r.namespace
	if r.allNamespaces {
		namespace = metav1.NamespaceAll
	}
	r.determiner, err = determiner.New(resourceClient, r.result, namespace)
	if err != nil {
		return
	}

	return
}

func (r *runner) completeResources(f cmdutil.Factory, resourceTypes string) error {
	r.result = f.
		NewBuilder().
		Unstructured().
		ContinueOnError().
		NamespaceParam(r.namespace).
		DefaultNamespace().
		AllNamespaces(r.allNamespaces).
		ResourceTypeOrNameArgs(true, resourceTypes).
		Flatten().
		Do()

	return r.result.Err()
}

func (r *runner) Run(ctx context.Context, f cmdutil.Factory,)(del []*deletecore.DeleteInspection, err error) {
	if err := r.result.Visit(func(info *cliresource.Info, err error) error {
		if info.Namespace == metav1.NamespaceSystem {
			return nil
		}
		//fmt.Println(info.Namespace,"/",info.Object.GetObjectKind().GroupVersionKind().Kind,"/",info.Name)
		ok, err := r.determiner.DetermineDeletion(ctx, info)
		if err != nil {
			return err
		}
		if !ok {
			return nil // skip deletion
		}
		del = append(del, &deletecore.DeleteInspection{
			Kind:      info.Object.GetObjectKind().GroupVersionKind().Kind,
			NameSpace: info.Namespace,
			Name:      info.Name,
		})
		return nil
	}); err != nil {
		return nil,err
	}
	//r.waitDeletion(uidMap, deletedInfos)
	log.Println("delete over")
	return
}

func NewDeleteInspectionSource(uri *url.URL) (*DeleteInspectionSource, error) {
	kubeClient, err := k8s.GetKubernetesClient(uri)
	dynamicClient,err:=k8s.GetKubernetesDynamicClient(uri)
	if err != nil {
		klog.Errorf("Failed to create kubernetes client, because of %v", err)
		return nil, err
	}
	result := DeleteInspectionSource{
		localNodeBuffer: make(chan *deletecore.DeleteInspection,LocalDeleteBufferSize),
		stopChannel:     make(chan struct{}),
		clientset: kubeClient,
		dynamicClient: dynamicClient,
	}
	return &result, nil
}

func (delete *DeleteInspectionSource)GetNewDeleteInspection()*deletecore.DeleteInspectionBatch{
	result:=deletecore.DeleteInspectionBatch{
		TimeStamp: time.Now(),
		Inspections: []*deletecore.DeleteInspection{},
	}
	delete.inspection()
	count:=0
DeleteLoop:
	for {
		if count >= delete.num {
			break DeleteLoop
		}
		select {
		case inspection := <-delete.localNodeBuffer:
			result.Inspections = append(result.Inspections, inspection)
			count++
		// 防止写入信道出现崩溃，及时退出。
		case <-time.After(time.Second * DefaultGetNodeTimeout):
			break DeleteLoop
		}
	}
	return &result
}

func (delete *DeleteInspectionSource)inspection(){
	dels:=newRunner()
	delete.num=len(dels)
	for _,v:=range dels{
		delete.localNodeBuffer<-v
	}
}

