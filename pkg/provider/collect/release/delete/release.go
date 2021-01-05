package delete

import (
	"context"
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/dynamic"
	"k8s.io/kubectl/pkg/cmd/util"

	"github.com/q8s-io/cluster-detector/pkg/entity"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/provider/collect/release"
	"github.com/q8s-io/cluster-detector/pkg/provider/collect/release/determiner"
)

var DeleteResourceListCh chan *entity.DeleteInspection

func NewKubernetesSource() {
	DeleteResourceListCh = make(chan *entity.DeleteInspection, entity.DefaultBufSize)
}

type runner struct {
	namespace     string
	allNamespaces bool
	timeout       time.Duration
	deleteOpts    *metav1.DeleteOptions
	determiner    determiner.Determiner
	dynamicClient dynamic.Interface
	printer       printers.ResourcePrinter
	result        *resource.Result
}

func newRunner() {
	for {
		runner := &runner{
			allNamespaces: true,
		}
		f := util.NewFactory(genericclioptions.NewConfigFlags(true))
		_ = runner.Complete(f)
		err := runner.Run(context.Background(), f)
		if err != nil {
			log.Fatalf("get deleteInspections error err:%v\n", err.Error())
		}
		time.Sleep(time.Second * time.Duration(config.Config.DeleteInspectionConfig.Speed))
	}
}

func (r *runner) Complete(f util.Factory) (err error) {
	r.deleteOpts = &metav1.DeleteOptions{}
	r.namespace = metav1.NamespaceAll
	allResources := "pod,cm,secret,PersistentVolume,PersistentVolumeClaim,Job,PodDisruptionBudget,HorizontalPodAutoscaler"
	if err = r.completeResources(f, allResources); err != nil {
		return
	}
	clientSet, err := f.KubernetesClientSet()
	if err != nil {
		return
	}
	r.dynamicClient, err = f.DynamicClient()
	if err != nil {
		return
	}
	resourceClient := release.NewClient(clientSet, r.dynamicClient)
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

func (r *runner) completeResources(f util.Factory, resourceTypes string) error {
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

func (r *runner) Run(ctx context.Context, f util.Factory) (err error) {
	if err := r.result.Visit(func(info *resource.Info, err error) error {
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
		DeleteResourceListCh <- &entity.DeleteInspection{
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

func StartWatch() {
	newRunner()
}
