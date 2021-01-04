package kube

import (
	"context"
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	cliresource "k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/dynamic"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	deletecore "github.com/q8s-io/cluster-detector/pkg/entity"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
	"github.com/q8s-io/cluster-detector/pkg/provider/kube/determiner"
	"github.com/q8s-io/cluster-detector/pkg/provider/kube/unused"
)

const (
	LocalDeleteBufferSize = 10000
)

var UnusedResourceList chan *deletecore.DeleteInspection

func NewKubernetesSource() *chan *deletecore.DeleteInspection {
	UnusedResourceList = make(chan *deletecore.DeleteInspection, LocalDeleteBufferSize)
	return &UnusedResourceList
}

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
		time.Sleep(time.Second * time.Duration(config.Config.DeleteInspectionConfig.Speed))
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

func (r *runner) Run(ctx context.Context, f cmdutil.Factory) (err error) {
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

func StartWatch() {
	newRunner()
}
