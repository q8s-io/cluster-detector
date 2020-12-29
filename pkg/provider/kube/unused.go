package kube

import (
	"context"
	"log"
	"net/url"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	cliresource "k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	deletecore "github.com/q8s-io/cluster-detector/pkg/entity"
	"github.com/q8s-io/cluster-detector/pkg/provider/kube/determiner"
	"github.com/q8s-io/cluster-detector/pkg/provider/kube/node"
	"github.com/q8s-io/cluster-detector/pkg/provider/kube/unused"
)

const (
	LocalDeleteBufferSize = 10000
)

type DeleteInspectionSource struct {
	// Large local buffer, periodically read.
	// channel's element size does not exceed 65535 bytes, so the pointer is used
	localNodeBuffer chan *deletecore.DeleteInspection
	// count node num
	num           int
	stopChannel   chan struct{}
	clientset     kubernetes.Interface
	dynamicClient dynamic.Interface
}

var UnusedResourceList chan *deletecore.DeleteInspection

type runner struct {
	namespace     string
	allNamespaces bool
	timeout       time.Duration
	deleteOpts    *metav1.DeleteOptions
	determiner    determiner.Determiner
	dynamicClient dynamic.Interface
	printer       printers.ResourcePrinter
	result        *cliresource.Result
}

func newRunner() {
	for {
		runner := &runner{
			allNamespaces: true,
		}
		f := cmdutil.NewFactory(genericclioptions.NewConfigFlags(true))
		_ = runner.Complete(f)
		err := runner.Run(context.Background(), f)
		if err != nil {
			log.Fatalf("get deleteInspections error err:%v\n", err.Error())
		}
		/*fmt.Println("Unused Resource sum: ",len(UnusedResourceList))*/
		time.Sleep(time.Second * 60)
		/*return dels*/
	}
}

func (r *runner) Complete(f cmdutil.Factory) (err error) {
	r.deleteOpts = &metav1.DeleteOptions{}
	r.namespace = metav1.NamespaceAll
	allResources := "pod,cm,secret,PersistentVolume,PersistentVolumeClaim,Job,PodDisruptionBudget,HorizontalPodAutoscaler"
	if err = r.completeResources(f, allResources); err != nil {
		return
	}
	clientset, err := f.KubernetesClientSet()
	if err != nil {
		return
	}
	r.dynamicClient, err = f.DynamicClient()
	if err != nil {
		return
	}
	resourceClient := resource.NewClient(clientset, r.dynamicClient)
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

func (r *runner) Run(ctx context.Context, f cmdutil.Factory) ( /*del []*deletecore.DeleteInspection,*/ err error) {
	if err := r.result.Visit(func(info *cliresource.Info, err error) error {
		if info.Namespace == metav1.NamespaceSystem {
			return nil
		}
		ok, err := r.determiner.DetermineDeletion(ctx, info)
		if err != nil {
			return err
		}
		if !ok {
			return nil // skip deletion
		}
		UnusedResourceList <- &deletecore.DeleteInspection{
			Kind:      info.Object.GetObjectKind().GroupVersionKind().Kind,
			NameSpace: info.Namespace,
			Name:      info.Name,
		}
		return nil
	}); err != nil {
		return err
	}
	log.Println("delete over")
	return
}

func NewDeleteInspectionSource(uri *url.URL) (*chan *deletecore.DeleteInspection, error) {
	UnusedResourceList = make(chan *deletecore.DeleteInspection, LocalDeleteBufferSize)
	go newRunner()
	return &UnusedResourceList, nil
}

func (delete *DeleteInspectionSource) GetNewDeleteInspection() *deletecore.DeleteInspectionBatch {
	result := deletecore.DeleteInspectionBatch{
		TimeStamp:   time.Now(),
		Inspections: []*deletecore.DeleteInspection{},
	}
	//delete.inspection()
	count := 0
DeleteLoop:
	for {
		if count >= len(UnusedResourceList) {
			break DeleteLoop
		}
		select {
		case inspection := <-UnusedResourceList:
			result.Inspections = append(result.Inspections, inspection)
			count++
		// 防止写入信道出现崩溃，及时退出。
		case <-time.After(time.Second * node.DefaultGetNodeTimeout):
			break DeleteLoop
		}
	}
	return &result
}
